package mcpdebug

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-delve/delve/service"
	"github.com/go-delve/delve/service/dap"
	delvedebugger "github.com/go-delve/delve/service/debugger"
)

// launchDelve starts a Delve DAP server and returns connection.
// Currently uses external process - can be switched to embedded later.
func launchDelve() (net.Conn, func(), error) {
	// For now, use the external approach that we know works
	return launchDelveExternal()
}

// launchDelveEmbedded starts an embedded Delve DAP server using the delve library.
// This is work-in-progress - has connection timing issues.
func launchDelveEmbedded() (net.Conn, func(), error) {
	// Create a TCP listener on localhost
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create listener: %w", err)
	}

	// Create disconnect channel
	disconnectCh := make(chan struct{})
	
	// Create readiness channel to signal when server is ready
	readyCh := make(chan error, 1)

	// Create service config for the DAP server
	config := &service.Config{
		Listener:       listener,
		DisconnectChan: disconnectCh,
		Debugger: delvedebugger.Config{
			// Basic debugger configuration
			WorkingDir: ".",
		},
	}

	// Create the embedded DAP server
	server := dap.NewServer(config)

	// Start the server in a goroutine
	go func() {
		defer listener.Close()
		defer close(readyCh)
		
		// Signal that we're ready to start
		readyCh <- nil
		
		// Run the DAP server (this will accept connections on the listener)
		server.Run()
	}()

	// Wait for server to signal readiness
	if err := <-readyCh; err != nil {
		listener.Close()
		return nil, nil, fmt.Errorf("server failed to start: %w", err)
	}

	// Connect to the embedded server with retry logic
	addr := listener.Addr().String()
	var conn net.Conn
	
	connectErr := RetryWithBackoff(context.Background(), RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}, func() error {
		var dialErr error
		conn, dialErr = net.Dial("tcp", addr)
		return dialErr
	})
	
	if connectErr != nil {
		listener.Close()
		server.Stop()
		return nil, nil, fmt.Errorf("failed to connect to embedded server at %s: %w", addr, connectErr)
	}

	// Cleanup function
	cleanup := func() {
		conn.Close()
		server.Stop()
	}

	return conn, cleanup, nil
}