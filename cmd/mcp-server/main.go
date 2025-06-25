package main

import (
	"log"
	"os"

	mcpdebug "github.com/roasbeef/mcp-debug"
	"github.com/lightningnetwork/lnd/actor"
)

func main() {
	// Create actor system
	actorSys := actor.NewActorSystem()
	defer actorSys.Shutdown()

	// Create debugger actor
	debuggerKey := actor.NewServiceKey[*mcpdebug.DebuggerCmd, *mcpdebug.DebuggerResp]("debugger")
	debugger := mcpdebug.NewDebugger()
	
	actor.RegisterWithSystem(
		actorSys, "debugger", debuggerKey,
		actor.NewFunctionBehavior[*mcpdebug.DebuggerCmd, *mcpdebug.DebuggerResp](debugger.Receive),
	)

	// Get debugger reference
	debuggerRef := actor.FindInReceptionist(actorSys.Receptionist(), debuggerKey)[0]

	// Create MCP server
	mcpServer := mcpdebug.NewMCPDebugServer(actorSys, debuggerRef)

	// Start serving
	if err := mcpServer.Serve(); err != nil {
		log.Fatalf("MCP server error: %v", err)
		os.Exit(1)
	}
}