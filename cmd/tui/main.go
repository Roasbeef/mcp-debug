package main

import (
	"log"
	"os"

	"github.com/lightningnetwork/lnd/actor"
	mcpdebug "github.com/roasbeef/mcp-debug"
)

func main() {
	// Start the daemon (this initializes the actor system and registers actors)
	mcpdebug.StartDaemon()
	
	// Create a router for debugger communication
	// This allows multiple debugger actors and better abstraction
	roundRobinStrategy := actor.NewRoundRobinStrategy[*mcpdebug.DebuggerCmd, *mcpdebug.DebuggerResp]()
	debuggerRouter := actor.NewRouter(
		mcpdebug.System.Receptionist(),
		mcpdebug.DebuggerKey,
		roundRobinStrategy,
		mcpdebug.System.DeadLetters(),
	)
	
	// Create MCP server using the router
	mcpServer := mcpdebug.NewMCPDebugServer(mcpdebug.System, debuggerRouter)
	
	// Set up cleanup
	defer mcpdebug.System.Shutdown()
	
	// Launch TUI with proper Bubble Tea components and actor router
	log.Println("Starting MCP Debug Server TUI Console...")
	if err := mcpdebug.RunTUI(mcpServer, mcpdebug.System); err != nil {
		log.Fatalf("TUI failed: %v", err)
		os.Exit(1)
	}
}