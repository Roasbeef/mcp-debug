package mcpdebug

import (
	"github.com/lightningnetwork/lnd/actor"
)

// DebuggerCommand is an interface for commands sent to the debugger actor.
type DebuggerCommand interface {
	isDebuggerCommand()
}

// StartDebuggerCmd is a command to start the debugger.
type StartDebuggerCmd struct{}

func (c *StartDebuggerCmd) isDebuggerCommand() {}

// StopDebuggerCmd is a command to stop the debugger.
type StopDebuggerCmd struct{}

func (c *StopDebuggerCmd) isDebuggerCommand() {}

// CreateSessionCmd is a command to create a new debug session.
type CreateSessionCmd struct{}

func (c *CreateSessionCmd) isDebuggerCommand() {}

// DebuggerCmd is the message sent to the debugger actor.
type DebuggerCmd struct {
	actor.BaseMessage
	Cmd DebuggerCommand
}

// MessageType returns the string identifier for this message type.
func (m *DebuggerCmd) MessageType() string {
	return "DebuggerCmd"
}

// DebuggerResponse is an interface for responses from debugger commands.
type DebuggerResponse interface {
	isDebuggerResponse()
}

// CreateSessionResp is the response from creating a session.
type CreateSessionResp struct {
	Session actor.ActorRef[*DAPRequest, *DAPResponse]
}

func (r *CreateSessionResp) isDebuggerResponse() {}

// DebuggerResp is the response from the debugger actor.
type DebuggerResp struct {
	actor.BaseMessage
	Resp DebuggerResponse
}

// MessageType returns the string identifier for this message type.
func (m *DebuggerResp) MessageType() string {
	return "DebuggerResp"
}