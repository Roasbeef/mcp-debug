package mcpdebug

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// launchDelveExternal starts a new Delve DAP process with retry logic.
// This is a working implementation using external dlv process.
func launchDelveExternal() (net.Conn, func(), error) {
	// Use retry logic to launch Delve
	var conn net.Conn
	var cleanup func()
	
	err := RetryWithBackoff(context.Background(), DefaultRetryConfig, func() error {
		var retryErr error
		conn, cleanup, retryErr = launchDelveOnceExternal()
		return retryErr
	})
	
	if err != nil {
		return nil, nil, fmt.Errorf("failed to launch Delve after retries: %w", err)
	}
	
	return conn, cleanup, nil
}

// launchDelveOnceExternal performs a single attempt to launch external Delve
func launchDelveOnceExternal() (net.Conn, func(), error) {
	// Find the path to the dlv executable.
	dlvPath, err := exec.LookPath("dlv")
	if err != nil {
		return nil, nil, fmt.Errorf("could not find 'dlv' executable: %w", err)
	}

	// Start the Delve DAP server. It will listen on a random free port.
	cmd := exec.Command(dlvPath, "dap")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("could not start dlv process: %w", err)
	}

	// The cleanup function will be returned to the caller to be executed when
	// the session is over.
	cleanup := func() {
		_ = cmd.Process.Kill()
	}

	// Delve prints the listening address to stdout. We need to read it.
	// Use a context with timeout to avoid blocking indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrCh := make(chan string, 1)
	errCh := make(chan error, 1)
	
	go func() {
		scanner := bufio.NewScanner(stdout)
		const prefix = "DAP server listening at: "
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, prefix) {
				addr := strings.TrimPrefix(line, prefix)
				addrCh <- addr
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("error reading stdout: %w", err)
		} else {
			errCh <- fmt.Errorf("unexpected end of stdout")
		}
	}()

	var addr string
	select {
	case addr = <-addrCh:
		// We got the address successfully.
	case err := <-errCh:
		cleanup()
		return nil, nil, err
	case <-ctx.Done():
		cleanup()
		return nil, nil, fmt.Errorf("timed out waiting for dlv dap address")
	}

	// Connect to the DAP server with retry logic
	var conn net.Conn
	connectErr := RetryWithBackoff(ctx, RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   2.0,
	}, func() error {
		var dialErr error
		conn, dialErr = net.Dial("tcp", addr)
		return dialErr
	})
	
	if connectErr != nil {
		cleanup()
		return nil, nil, fmt.Errorf("could not connect to dlv dap server at %s: %w", addr, connectErr)
	}

	return conn, cleanup, nil
}