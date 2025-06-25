package mcpdebug

import (
	"bufio"
	"context"
	"fmt"
	"io"
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
	conn, cleanup, err := launchDelve()
	if err != nil {
		return nil, err
	}

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
				select {
				case s.errors <- fmt.Errorf("error reading DAP message: %w", err):
				case <-s.quit:
					return
				}
			}
			return
		}

		switch m := msg.(type) {
		case dap.ResponseMessage:
			select {
			case s.responses <- m:
			case <-s.quit:
				return
			}
		case dap.EventMessage:
			select {
			case s.events <- m:
			case <-s.quit:
				return
			}
		default:
			// Log unexpected message types but don't fail
		}
	}
}

// Receive is the actor's message handler.
func (s *Session) Receive(actorCtx context.Context, msg *DAPRequest) fn.Result[*DAPResponse] {
	// First, send the request to the DAP server.
	if err := dap.WriteProtocolMessage(s.conn, msg.Request); err != nil {
		return fn.Err[*DAPResponse](fmt.Errorf("error writing DAP message: %w", err))
	}

	// Now, wait for the corresponding response or an event.
	select {
	case resp := <-s.responses:
		// We got a direct response to our request.
		return fn.Ok(&DAPResponse{Response: resp})

	case event := <-s.events:
		// Sometimes we get an event before a response (e.g., a stopped
		// event after a continue request). We'll just return it.
		return fn.Ok(&DAPResponse{Response: event})

	case err := <-s.errors:
		// An error occurred in the read loop.
		return fn.Err[*DAPResponse](err)

	case <-actorCtx.Done():
		// The actor is shutting down.
		return fn.Err[*DAPResponse](actorCtx.Err())
	}
}
