package debugger

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/fn/v2"
)

// Session is an actor that manages a single DAP debugging session.
type Session struct {
	conn    net.Conn
	cleanup func()

	// The actor's quit channel is used to signal that the session should be
	// terminated.
	quit chan struct{}

	// The following channels are used to route messages from the DAP server
	// back to the actor's message loop.
	responses chan dap.Message
	events    chan dap.Message
	errors    chan error
}

// NewSession creates a new debugging session actor.
// It launches a new Delve DAP server and connects to it.
func NewSession() (*Session, error) {
	log.Printf("[Session] Creating new debugging session...")
	
	conn, cleanup, err := launchDelve()
	if err != nil {
		log.Printf("[Session] Failed to launch Delve: %v", err)
		return nil, err
	}
	
	log.Printf("[Session] Successfully connected to Delve DAP server")

	s := &Session{
		conn:      conn,
		cleanup:   cleanup,
		quit:      make(chan struct{}),
		responses: make(chan dap.Message, 1),
		events:    make(chan dap.Message, 1),
		errors:    make(chan error, 1),
	}

	// Start the read loop immediately
	go s.readLoop()
	log.Printf("[Session] Started read loop")

	return s, nil
}

// Stop terminates the DAP session and cleans up resources.
func (s *Session) Stop() {
	close(s.quit)
	s.conn.Close()
	s.cleanup()
}

// readLoop is a long-running goroutine that reads messages from the DAP
// server and forwards them to the appropriate channels.
func (s *Session) readLoop() {
	reader := bufio.NewReader(s.conn)

	for {
		select {
		case <-s.quit:
			return
		default:
		}

		msg, err := dap.ReadProtocolMessage(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("[ReadLoop] Error reading DAP message: %v", err)
				select {
				case s.errors <- fmt.Errorf("error reading DAP message: %w", err):
				case <-s.quit:
					return
				}
			}
			return
		}

		log.Printf("[ReadLoop] Received message: %T", msg)
		
		switch m := msg.(type) {
		case dap.ResponseMessage:
			log.Printf("[ReadLoop] Forwarding response: %T", m)
			select {
			case s.responses <- m:
			case <-s.quit:
				return
			}
		case dap.EventMessage:
			log.Printf("[ReadLoop] Forwarding event: %T", m)
			select {
			case s.events <- m:
			case <-s.quit:
				return
			}
		default:
			// Log unexpected message types but don't fail
			log.Printf("[ReadLoop] Unexpected message type: %T", msg)
		}
	}
}

// Receive is the actor's message handler.
func (s *Session) Receive(actorCtx context.Context, msg *DAPRequest) fn.Result[*DAPResponse] {
	// Log the outgoing request
	log.Printf("[Session] Sending DAP request: %T", msg.Request)
	
	// First, send the request to the DAP server.
	if err := dap.WriteProtocolMessage(s.conn, msg.Request); err != nil {
		return fn.Err[*DAPResponse](fmt.Errorf("error writing DAP message: %w", err))
	}

	// Now, wait for the corresponding response.
	// We MUST wait for the actual response, not return events as responses.
	// Events should be handled separately or collected.
	for {
		select {
		case resp := <-s.responses:
			// We got a direct response to our request.
			log.Printf("[Session] Received DAP response: %T", resp)
			return fn.Ok(&DAPResponse{Response: resp})

		case event := <-s.events:
			// Log the event
			log.Printf("[Session] Received DAP event: %T", event)
			
			// Events are NOT responses! We should handle them but continue waiting
			// for the actual response to our request.
			switch e := event.(type) {
			case *dap.OutputEvent:
				// Log output but continue waiting for the actual response
				log.Printf("[Session] Output event (continuing): %s", e.Body.Output)
				continue
			case *dap.StoppedEvent:
				// Important event but not a response to our request
				log.Printf("[Session] Stopped event (continuing to wait for response): threadId=%d, reason=%s", 
					e.Body.ThreadId, e.Body.Reason)
				continue
			case *dap.TerminatedEvent, *dap.ExitedEvent:
				// These might indicate the session is ending
				log.Printf("[Session] Termination event (continuing): %T", event)
				continue
			default:
				// For other events, continue waiting
				log.Printf("[Session] Other event (continuing): %T", event)
				continue
			}

		case err := <-s.errors:
			// An error occurred in the read loop.
			log.Printf("[Session] Error in read loop: %v", err)
			return fn.Err[*DAPResponse](err)

		case <-actorCtx.Done():
			// The actor is shutting down.
			log.Printf("[Session] Actor context done: %v", actorCtx.Err())
			return fn.Err[*DAPResponse](actorCtx.Err())
		}
	}
}
