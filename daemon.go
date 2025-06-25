package mcpdebug

import (
	"github.com/lightningnetwork/lnd/actor"
)

var (
	// System is the main actor system.
	System = actor.NewActorSystem()

	// DebuggerKey is the service key for the debugger actor.
	DebuggerKey = actor.NewServiceKey[*DebuggerCmd, *DebuggerResp]("debugger")
)

// StartDaemon starts the actor system and the debugger actor.
func StartDaemon() {
	// Create a new debugger actor.
	debugger := newDebugger(System)

	// Register the debugger actor with the actor system.
	actor.RegisterWithSystem(
		System, "debugger", DebuggerKey, actor.NewFunctionBehavior(debugger.Receive),
	)
}
