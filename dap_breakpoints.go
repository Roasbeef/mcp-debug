package mcpdebug

import (
	"context"
	"fmt"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
)

// SetBreakpoints sets breakpoints using the provided breakpoint locations.
// This function replaces any existing breakpoints for the file with the new
// set of breakpoints and provides comprehensive breakpoint configuration.
func SetBreakpoints(session actor.ActorRef[*DAPRequest, *DAPResponse],
	breakpoints []BreakpointLocation) (*dap.SetBreakpointsResponse, error) {

	if len(breakpoints) == 0 {
		return nil, fmt.Errorf("no breakpoints provided")
	}

	// All breakpoints must be for the same file
	sourcePath := breakpoints[0].File
	for _, bp := range breakpoints[1:] {
		if bp.File != sourcePath {
			return nil, fmt.Errorf(
				"all breakpoints must be for the same file")
		}
	}

	// Convert breakpoint locations to DAP source breakpoints
	sourceBreakpoints := make([]dap.SourceBreakpoint, len(breakpoints))
	for i, bp := range breakpoints {
		sourceBreakpoints[i] = dap.SourceBreakpoint{
			Line:         bp.Line,
			Column:       bp.Column,
			Condition:    bp.Condition,
			HitCondition: bp.HitCondition,
			LogMessage:   bp.LogMessage,
		}
	}

	req := &dap.SetBreakpointsRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "setBreakpoints",
		},
		Arguments: dap.SetBreakpointsArguments{
			Source: dap.Source{
				Path: sourcePath,
			},
			Breakpoints: sourceBreakpoints,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.SetBreakpointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// SetFunctionBreakpoints sets breakpoints on function names using the
// provided function breakpoint configurations. This is useful for setting
// breakpoints on functions without knowing their exact source location.
func SetFunctionBreakpoints(
	session actor.ActorRef[*DAPRequest, *DAPResponse],
	functionBreakpoints []FunctionBreakpoint) (*dap.SetFunctionBreakpointsResponse, error) {

	// Convert function breakpoint configurations to DAP function breakpoints
	dapFunctionBreakpoints := make([]dap.FunctionBreakpoint, len(functionBreakpoints))
	for i, bp := range functionBreakpoints {
		dapFunctionBreakpoints[i] = dap.FunctionBreakpoint{
			Name:         bp.Name,
			Condition:    bp.Condition,
			HitCondition: bp.HitCondition,
		}
	}

	req := &dap.SetFunctionBreakpointsRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "setFunctionBreakpoints",
		},
		Arguments: dap.SetFunctionBreakpointsArguments{
			Breakpoints: dapFunctionBreakpoints,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.SetFunctionBreakpointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}