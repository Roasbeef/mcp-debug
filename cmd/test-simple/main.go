package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
	mcpdebug "github.com/roasbeef/mcp-debug"
)

func main() {
	fmt.Println("🚀 Testing simple debug session...")
	
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
	
	// Add timeout to the future
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	result, err := future.Await(ctx).Unpack()
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
	
	// Test session initialization with timeout
	sessionRef := sessionRefs[0]
	fmt.Println("🔌 Initializing debug session...")
	
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	
	// Create InitializeRequest directly  
	initReq := &mcpdebug.DAPRequest{
		Request: &dap.InitializeRequest{
			Request: dap.Request{
				ProtocolMessage: dap.ProtocolMessage{
					Type: "request",
				},
				Command: "initialize",
			},
			Arguments: dap.InitializeRequestArguments{
				ClientID: "test-client",
			},
		},
	}
	
	future2 := sessionRef.Ask(ctx2, initReq)
	result2, err := future2.Await(ctx2).Unpack()
	if err != nil {
		log.Fatalf("Error initializing session: %v", err)
	}
	
	fmt.Printf("✓ Session initialized successfully!\n")
	fmt.Printf("  Response: %+v\n", result2)
	
	// Clean shutdown
	mcpdebug.System.Shutdown()
	fmt.Println("✓ System shutdown complete")
}