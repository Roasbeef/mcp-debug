package debugger

import (
	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
)

// DAPRequest is a message wrapper for sending a DAP request to a Session actor.
// It contains the raw DAP request that should be sent to the Delve server.
type DAPRequest struct {
	actor.BaseMessage
	Request dap.Message
}

// MessageType returns the string identifier for this message type.
func (r *DAPRequest) MessageType() string {
	return "DAPRequest"
}

// DAPResponse is a message wrapper for receiving a DAP response from a Session
// actor. It contains the raw DAP response from the Delve server.
type DAPResponse struct {
	actor.BaseMessage
	Response dap.Message
}

// MessageType returns the string identifier for this message type.
func (r *DAPResponse) MessageType() string {
	return "DAPResponse"
}
