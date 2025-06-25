package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-dap"
)

func main() {
	fmt.Println("🔍 Debugging DAP communication directly...")
	
	// Test 1: Can we launch dlv and connect?
	fmt.Println("\n--- Test 1: Launch dlv dap server ---")
	conn, cleanup, err := launchDelveDebug()
	if err != nil {
		log.Fatalf("Failed to launch delve: %v", err)
	}
	defer cleanup()
	
	fmt.Println("✓ Successfully connected to dlv dap server")
	
	// Test 2: Can we send an initialize request?
	fmt.Println("\n--- Test 2: Send initialize request ---")
	initReq := &dap.InitializeRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  1,
				Type: "request",
			},
			Command: "initialize",
		},
		Arguments: dap.InitializeRequestArguments{
			ClientID:         "debug-test",
			AdapterID:        "go",
			LinesStartAt1:    true,
			ColumnsStartAt1:  true,
		},
	}
	
	fmt.Printf("Sending initialize request: %+v\n", initReq)
	
	err = dap.WriteProtocolMessage(conn, initReq)
	if err != nil {
		log.Fatalf("Failed to write initialize request: %v", err)
	}
	
	fmt.Println("✓ Initialize request sent successfully")
	
	// Test 3: Can we read the response?
	fmt.Println("\n--- Test 3: Read initialize response ---")
	
	// Set a read timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	reader := bufio.NewReader(conn)
	respMsg, err := dap.ReadProtocolMessage(reader)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	
	fmt.Printf("✓ Received response: %T\n", respMsg)
	
	switch resp := respMsg.(type) {
	case *dap.InitializeResponse:
		fmt.Printf("✓ Initialize response received!\n")
		fmt.Printf("  Supports configuration done: %t\n", resp.Body.SupportsConfigurationDoneRequest)
		fmt.Printf("  Supports function breakpoints: %t\n", resp.Body.SupportsFunctionBreakpoints)
	case *dap.ErrorResponse:
		fmt.Printf("❌ Error response: %s\n", resp.Body.Error.Format)
	default:
		fmt.Printf("? Unexpected response type: %T\n", resp)
	}
	
	fmt.Println("\n✅ DAP communication test completed successfully!")
}

// launchDelveDebug is a simplified version for debugging
func launchDelveDebug() (net.Conn, func(), error) {
	// Find the path to the dlv executable.
	dlvPath, err := exec.LookPath("dlv")
	if err != nil {
		return nil, nil, fmt.Errorf("could not find 'dlv' executable: %w", err)
	}

	fmt.Printf("Found dlv at: %s\n", dlvPath)

	// Start the Delve DAP server
	cmd := exec.Command(dlvPath, "dap", "--log")
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("could not start dlv process: %w", err)
	}

	fmt.Printf("Started dlv process with PID: %d\n", cmd.Process.Pid)

	cleanup := func() {
		fmt.Println("Cleaning up dlv process...")
		_ = cmd.Process.Kill()
	}

	// Read stderr in background for debugging
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[dlv stderr] %s\n", scanner.Text())
		}
	}()

	// Read stdout to get the listening address
	scanner := bufio.NewScanner(stdout)
	const prefix = "DAP server listening at: "
	
	var addr string
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("[dlv stdout] %s\n", line)
		if strings.HasPrefix(line, prefix) {
			addr = strings.TrimPrefix(line, prefix)
			break
		}
	}
	
	if addr == "" {
		cleanup()
		return nil, nil, fmt.Errorf("could not find DAP server address in output")
	}
	
	fmt.Printf("DAP server listening at: %s\n", addr)
	
	// Small delay to ensure server is ready
	time.Sleep(100 * time.Millisecond)
	
	// Connect to the DAP server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("could not connect to dlv dap server: %w", err)
	}
	
	fmt.Println("Successfully connected to DAP server")
	
	return conn, cleanup, nil
}