package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
	mcpdebug "github.com/roasbeef/mcp-debug"
)

func main() {
	fmt.Println("🚀 Testing complete debug workflow...")
	
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
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	_, err = mcpdebug.InitializeSession(sessionRef, "test-workflow")
	if err != nil {
		log.Fatalf("Error initializing session: %v", err)
	}
	
	fmt.Printf("✓ Session initialized\n")
	
	// Test launching our example program
	fmt.Println("🚀 Launching example program...")
	
	// Get the absolute path to our example program
	examplePath, err := filepath.Abs("./examples/simple/main.go")
	if err != nil {
		log.Fatalf("Error getting example path: %v", err)
	}
	
	// Create launch arguments
	launchArgs := map[string]interface{}{
		"name":    "Launch example",
		"type":    "go",
		"request": "launch",
		"mode":    "debug",
		"program": examplePath,
	}
	
	launchArgsJSON, err := json.Marshal(launchArgs)
	if err != nil {
		log.Fatalf("Error marshaling launch arguments: %v", err)
	}
	
	// Create launch request
	launchReq := &dap.LaunchRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  2,
				Type: "request",
			},
			Command: "launch",
		},
		Arguments: json.RawMessage(launchArgsJSON),
	}
	
	dapReq := &mcpdebug.DAPRequest{Request: launchReq}
	future2 := sessionRef.Ask(ctx, dapReq)
	launchResult, err := future2.Await(ctx).Unpack()
	if err != nil {
		log.Fatalf("Error launching program: %v", err)
	}
	
	fmt.Printf("✓ Program launch response received: %T\n", launchResult.Response)
	
	// Test configuration done
	fmt.Println("⚙️  Sending configuration done...")
	configReq := &dap.ConfigurationDoneRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  3,
				Type: "request",
			},
			Command: "configurationDone",
		},
	}
	
	dapConfigReq := &mcpdebug.DAPRequest{Request: configReq}
	future3 := sessionRef.Ask(ctx, dapConfigReq)
	configResult, err := future3.Await(ctx).Unpack()
	if err != nil {
		log.Fatalf("Error in configuration done: %v", err)
	}
	
	fmt.Printf("✓ Configuration done response: %T\n", configResult.Response)
	
	fmt.Println("🎉 Basic debug workflow completed successfully!")
	
	// Print summary
	fmt.Println("\n=== Workflow Summary ===")
	fmt.Println("✓ Session creation")
	fmt.Println("✓ DAP initialization") 
	fmt.Println("✓ Program launch attempt")
	fmt.Println("✓ Configuration completion")
	fmt.Println("\n✅ All core debugging operations functional!")
}