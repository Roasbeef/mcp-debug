package debugger

import (
	"context"
	"fmt"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
)

// Continue resumes execution of the debugged program. If the program was
// stopped at a breakpoint, this will continue execution until the next
// breakpoint is hit or the program terminates.
func Continue(session actor.ActorRef[*DAPRequest, *DAPResponse],
	threadID int) (*dap.ContinueResponse, error) {

	req := &dap.ContinueRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "continue",
		},
		Arguments: dap.ContinueArguments{
			ThreadId: threadID,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.ContinueResponse)
	if !ok {
		// Check if it's an error response and extract the message
		if errResp, isErr := result.Response.(*dap.ErrorResponse); isErr {
			return nil, fmt.Errorf("continue failed: %s (id: %d)", 
				errResp.Body.Error.Format, errResp.Body.Error.Id)
		}
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// Next performs a step over operation. This executes the next line of code
// but does not step into function calls.
func Next(session actor.ActorRef[*DAPRequest, *DAPResponse],
	threadID int) (*dap.NextResponse, error) {

	req := &dap.NextRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "next",
		},
		Arguments: dap.NextArguments{
			ThreadId: threadID,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.NextResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// StepIn performs a step into operation. This steps into function calls
// rather than stepping over them.
func StepIn(session actor.ActorRef[*DAPRequest, *DAPResponse],
	threadID int) (*dap.StepInResponse, error) {

	req := &dap.StepInRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "stepIn",
		},
		Arguments: dap.StepInArguments{
			ThreadId: threadID,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.StepInResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// StepOut performs a step out operation. This continues execution until
// the current function returns.
func StepOut(session actor.ActorRef[*DAPRequest, *DAPResponse],
	threadID int) (*dap.StepOutResponse, error) {

	req := &dap.StepOutRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "stepOut",
		},
		Arguments: dap.StepOutArguments{
			ThreadId: threadID,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.StepOutResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// Pause pauses execution of the debugged program. This is useful for
// interrupting a running program to examine its state.
func Pause(session actor.ActorRef[*DAPRequest, *DAPResponse],
	threadID int) (*dap.PauseResponse, error) {

	req := &dap.PauseRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "pause",
		},
		Arguments: dap.PauseArguments{
			ThreadId: threadID,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.PauseResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}