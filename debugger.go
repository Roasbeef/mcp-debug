package mcpdebug

import (
	"context"
	"fmt"

	"github.com/lightningnetwork/lnd/actor"
	"github.com/lightningnetwork/lnd/fn/v2"
)

// debugger is an actor that is responsible for creating and managing debugger
// sessions.
type debugger struct {
	nextSessionID int
	system        *actor.ActorSystem
}

// newDebugger creates a new debugger actor factory.
func newDebugger(system *actor.ActorSystem) *debugger {
	return &debugger{
		system: system,
	}
}

// NewDebugger creates a new debugger actor that doesn't require the system
// to be passed in (it will get it from the actor context).
func NewDebugger() *debugger {
	return &debugger{}
}

// Receive is the message handler for the debugger actor.
func (d *debugger) Receive(actorCtx context.Context, msg *DebuggerCmd) fn.Result[*DebuggerResp] {
	switch cmd := msg.Cmd.(type) {
	case *StartDebuggerCmd:
		// Create a new session.
		session, err := NewSession()
		if err != nil {
			return fn.Err[*DebuggerResp](fmt.Errorf("could not create session: %w", err))
		}

		// Create a unique service key for this session.
		d.nextSessionID++
		sessionID := fmt.Sprintf("session-%d", d.nextSessionID)
		sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse](sessionID)

		// Register the session actor with the system.
		actor.RegisterWithSystem(
			d.system, sessionID, sessionKey, actor.NewFunctionBehavior(session.Receive),
		)

		// Get the session reference from the receptionist
		sessionRef := actor.FindInReceptionist(d.system.Receptionist(), sessionKey)[0]
		
		// Return the session reference to the caller
		return fn.Ok(&DebuggerResp{
			Resp: &CreateSessionResp{Session: sessionRef},
		})

	case *CreateSessionCmd:
		// Create a new session.
		session, err := NewSession()
		if err != nil {
			return fn.Err[*DebuggerResp](fmt.Errorf("could not create session: %w", err))
		}

		// Create a unique service key for this session.
		d.nextSessionID++
		sessionID := fmt.Sprintf("session-%d", d.nextSessionID)
		sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse](sessionID)

		// Register the session actor with the system.
		actor.RegisterWithSystem(
			d.system, sessionID, sessionKey, actor.NewFunctionBehavior(session.Receive),
		)

		// Get the session reference from the receptionist
		sessionRef := actor.FindInReceptionist(d.system.Receptionist(), sessionKey)[0]
		
		// Return the session reference to the caller
		return fn.Ok(&DebuggerResp{
			Resp: &CreateSessionResp{Session: sessionRef},
		})

	default:
		return fn.Err[*DebuggerResp](fmt.Errorf("unknown command type: %T", cmd))
	}
}
