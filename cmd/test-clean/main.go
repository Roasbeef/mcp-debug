package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lightningnetwork/lnd/actor"
	mcpdebug "github.com/roasbeef/mcp-debug"
)

func main() {
	fmt.Println("🚀 Testing clean debug session...")
	
	// Start the daemon
	mcpdebug.StartDaemon()
	defer mcpdebug.System.Shutdown()
	
	// Find the debugger actor
	debuggerRefs := actor.FindInReceptionist(
		mcpdebug.System.Receptionist(), mcpdebug.DebuggerKey,
	)
	if len(debuggerRefs) == 0 {
		log.Fatal("Error: Could not find debugger actor factory")
	}
	debuggerRef := debuggerRefs[0]
	
	fmt.Println("✓ Actor system started")
	
	// Create debug session
	startMsg := &mcpdebug.DebuggerMsg{
		Command: &mcpdebug.StartDebuggerCmd{},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	future := debuggerRef.Ask(ctx, startMsg)
	result, err := future.Await(ctx).Unpack()
	if err != nil {
		log.Fatalf("Error creating debug session: %v", err)
	}
	
	fmt.Printf("✓ Debug session created: %s\n", result.Status)
	
	// Find session actor
	sessionKey := actor.NewServiceKey[*mcpdebug.DAPRequest, *mcpdebug.DAPResponse](result.Status)
	sessionRefs := actor.FindInReceptionist(mcpdebug.System.Receptionist(), sessionKey)
	
	if len(sessionRefs) == 0 {
		log.Fatal("Error: Could not find session actor")
	}
	sessionRef := sessionRefs[0]
	
	// Initialize session
	fmt.Println("🔌 Initializing session...")
	initResp, err := mcpdebug.InitializeSession(sessionRef, "test-client")
	if err != nil {
		log.Fatalf("Error initializing session: %v", err)
	}
	
	fmt.Printf("✅ Session initialized!\n")
	fmt.Printf("  Supports configuration done: %t\n", initResp.Body.SupportsConfigurationDoneRequest)
	fmt.Printf("  Supports function breakpoints: %t\n", initResp.Body.SupportsFunctionBreakpoints)
	fmt.Printf("  Supports conditional breakpoints: %t\n", initResp.Body.SupportsConditionalBreakpoints)
	
	fmt.Println("🎉 All tests passed!")
}