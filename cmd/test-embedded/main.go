package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
	mcpdebug "github.com/roasbeef/mcp-debug"
)

func main() {
	fmt.Println("🚀 Testing embedded Delve DAP server...")
	
	// Start the daemon
	mcpdebug.StartDaemon()
	
	// Find the debugger actor
	debuggerRefs := actor.FindInReceptionist(
		mcpdebug.System.Receptionist(), mcpdebug.DebuggerKey,
	)
	if len(debuggerRefs) == 0 {
		log.Fatal("Error: Could not find debugger actor factory")
	}
	debuggerRef := debuggerRefs[0]
	
	fmt.Println("✓ Actor system started successfully")
	fmt.Println("✓ Debugger factory actor found")
	
	// Test creating a debug session
	startMsg := &mcpdebug.DebuggerMsg{
		Command: &mcpdebug.StartDebuggerCmd{},
	}
	
	fmt.Println("🔧 Creating new debug session...")
	future := debuggerRef.Ask(context.Background(), startMsg)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		log.Fatalf("Error creating debug session: %v", err)
	}
	
	fmt.Printf("✓ Debug session created with ID: %s\n", result.Status)
	
	// Try to find the session actor
	sessionKey := actor.NewServiceKey[*mcpdebug.DAPRequest, *mcpdebug.DAPResponse](result.Status)
	sessionRefs := actor.FindInReceptionist(mcpdebug.System.Receptionist(), sessionKey)
	
	if len(sessionRefs) == 0 {
		log.Fatal("Error: Could not find session actor")
	}
	
	fmt.Printf("✓ Session actor found for session: %s\n", result.Status)
	
	// Test session initialization
	sessionRef := sessionRefs[0]
	fmt.Println("🔌 Initializing debug session...")
	
	// Use retry logic to handle timing issues
	var initResp *dap.InitializeResponse
	err = mcpdebug.RetryWithBackoff(context.Background(), mcpdebug.DefaultRetryConfig, func() error {
		var retryErr error
		initResp, retryErr = mcpdebug.InitializeSession(sessionRef, "test-client")
		return retryErr
	})
	if err != nil {
		log.Fatalf("Error initializing session: %v", err)
	}
	
	fmt.Printf("✓ Session initialized successfully!\n")
	fmt.Printf("  Supports configuration done: %t\n", initResp.Body.SupportsConfigurationDoneRequest)
	fmt.Printf("  Supports function breakpoints: %t\n", initResp.Body.SupportsFunctionBreakpoints)
	fmt.Printf("  Supports conditional breakpoints: %t\n", initResp.Body.SupportsConditionalBreakpoints)
	fmt.Printf("  Supports terminate request: %t\n", initResp.Body.SupportsTerminateRequest)
	
	// Clean shutdown
	mcpdebug.System.Shutdown()
	fmt.Println("✓ System shutdown complete")
}