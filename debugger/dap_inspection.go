package debugger

import (
	"context"
	"fmt"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
)

// GetThreadsInfo retrieves information about all threads in the debugged
// program, returning a slice of ThreadInfo wrapper types for easier handling.
func GetThreadsInfo(session actor.ActorRef[*DAPRequest, *DAPResponse],
) ([]ThreadInfo, error) {

	resp, err := GetThreads(session)
	if err != nil {
		return nil, err
	}

	// Convert DAP threads to ThreadInfo wrapper types
	threads := make([]ThreadInfo, len(resp.Body.Threads))
	for i, thread := range resp.Body.Threads {
		threads[i] = ThreadInfo{
			ID:   thread.Id,
			Name: thread.Name,
		}
	}

	return threads, nil
}

// GetStackFrames retrieves the call stack for the specified thread,
// returning a slice of StackFrame wrapper types for easier handling.
func GetStackFrames(session actor.ActorRef[*DAPRequest, *DAPResponse],
	threadID int) ([]StackFrame, error) {

	resp, err := GetStackTrace(session, threadID)
	if err != nil {
		return nil, err
	}

	// Convert DAP stack frames to StackFrame wrapper types
	frames := make([]StackFrame, len(resp.Body.StackFrames))
	for i, frame := range resp.Body.StackFrames {
		frames[i] = StackFrame{
			ID:     frame.Id,
			Name:   frame.Name,
			Line:   frame.Line,
			Column: frame.Column,
			Source: SourceInfo{
				Path: frame.Source.Path,
				Name: frame.Source.Name,
			},
		}
	}

	return frames, nil
}

// GetVariableScopes retrieves the variable scopes for the specified frame,
// returning a slice of VariableScope wrapper types for easier handling.
func GetVariableScopes(session actor.ActorRef[*DAPRequest, *DAPResponse],
	frameID int) ([]VariableScope, error) {

	resp, err := GetScopes(session, frameID)
	if err != nil {
		return nil, err
	}

	// Convert DAP scopes to VariableScope wrapper types
	scopes := make([]VariableScope, len(resp.Body.Scopes))
	for i, scope := range resp.Body.Scopes {
		scopes[i] = VariableScope{
			Name:               scope.Name,
			VariablesReference: scope.VariablesReference,
			Expensive:          scope.Expensive,
		}
	}

	return scopes, nil
}

// GetVariableList retrieves variables for the specified scope or variable
// reference, returning a slice of Variable wrapper types for easier handling.
func GetVariableList(session actor.ActorRef[*DAPRequest, *DAPResponse],
	variablesReference int) ([]Variable, error) {

	resp, err := GetVariables(session, variablesReference)
	if err != nil {
		return nil, err
	}

	// Convert DAP variables to Variable wrapper types
	variables := make([]Variable, len(resp.Body.Variables))
	for i, variable := range resp.Body.Variables {
		variables[i] = Variable{
			Name:                variable.Name,
			Value:               variable.Value,
			Type:                variable.Type,
			VariablesReference:  variable.VariablesReference,
			IndexedVariables:    variable.IndexedVariables,
			NamedVariables:      variable.NamedVariables,
		}
	}

	return variables, nil
}

// EvaluateExpressionResult evaluates an expression and returns the result
// as an EvaluationResult wrapper type for easier handling.
func EvaluateExpressionResult(session actor.ActorRef[*DAPRequest, *DAPResponse],
	expression string, frameID int) (*EvaluationResult, error) {

	resp, err := EvaluateExpression(session, expression, frameID)
	if err != nil {
		return nil, err
	}

	// Convert DAP evaluation result to EvaluationResult wrapper type
	result := &EvaluationResult{
		Result:              resp.Body.Result,
		Type:                resp.Body.Type,
		VariablesReference:  resp.Body.VariablesReference,
		IndexedVariables:    resp.Body.IndexedVariables,
		NamedVariables:      resp.Body.NamedVariables,
	}

	return result, nil
}

// GetThreads retrieves information about all threads in the debugged program.
// This is useful for understanding the program's execution state and for
// targeting specific threads with debugging operations.
func GetThreads(session actor.ActorRef[*DAPRequest, *DAPResponse],
) (*dap.ThreadsResponse, error) {

	req := &dap.ThreadsRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "threads",
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.ThreadsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// GetStackTrace retrieves the call stack for the specified thread. This
// provides information about the current execution stack including function
// names, source locations, and frame IDs for variable inspection.
func GetStackTrace(session actor.ActorRef[*DAPRequest, *DAPResponse],
	threadID int) (*dap.StackTraceResponse, error) {

	req := &dap.StackTraceRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "stackTrace",
		},
		Arguments: dap.StackTraceArguments{
			ThreadId: threadID,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.StackTraceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// GetScopes retrieves the variable scopes available in the specified stack
// frame. Scopes typically include local variables, function arguments, and
// global variables.
func GetScopes(session actor.ActorRef[*DAPRequest, *DAPResponse],
	frameID int) (*dap.ScopesResponse, error) {

	req := &dap.ScopesRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "scopes",
		},
		Arguments: dap.ScopesArguments{
			FrameId: frameID,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.ScopesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// GetVariables retrieves the variables available in the specified scope or
// variable reference. This is used to inspect variable values and their
// properties during debugging.
func GetVariables(session actor.ActorRef[*DAPRequest, *DAPResponse],
	variablesReference int) (*dap.VariablesResponse, error) {

	req := &dap.VariablesRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "variables",
		},
		Arguments: dap.VariablesArguments{
			VariablesReference: variablesReference,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.VariablesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}

// EvaluateExpression evaluates an expression in the context of the specified
// frame and returns the result. This is useful for inspecting complex
// expressions or calling functions during debugging.
func EvaluateExpression(session actor.ActorRef[*DAPRequest, *DAPResponse],
	expression string, frameID int) (*dap.EvaluateResponse, error) {

	req := &dap.EvaluateRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Type: "request",
			},
			Command: "evaluate",
		},
		Arguments: dap.EvaluateArguments{
			Expression: expression,
			FrameId:    frameID,
			Context:    "watch", // Can be "watch", "repl", "hover", etc.
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.EvaluateResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T",
			result.Response)
	}

	return resp, nil
}