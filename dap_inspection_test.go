package mcpdebug

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
	"github.com/lightningnetwork/lnd/fn/v2"
	"github.com/stretchr/testify/require"
)

// MockInspectionSession extends MockSession to handle inspection commands.
type MockInspectionSession struct {
	*MockSession
}

// NewMockInspectionSession creates a new mock inspection session.
func NewMockInspectionSession() *MockInspectionSession {
	return &MockInspectionSession{
		MockSession: NewMockSession(),
	}
}

// Receive implements the actor Receive method for inspection commands.
func (m *MockInspectionSession) Receive(actorCtx context.Context, 
	msg *DAPRequest) fn.Result[*DAPResponse] {
	
	// Record the request
	m.requests = append(m.requests, msg.Request)
	
	// Extract command from the request
	var command string
	switch req := msg.Request.(type) {
	case *dap.ThreadsRequest:
		command = req.Command
	case *dap.StackTraceRequest:
		command = req.Command
	case *dap.ScopesRequest:
		command = req.Command
	case *dap.VariablesRequest:
		command = req.Command
	case *dap.EvaluateRequest:
		command = req.Command
	default:
		return fn.Err[*DAPResponse](
			fmt.Errorf("unknown request type"))
	}
	
	// Return the predefined response
	if response, exists := m.responses[command]; exists {
		return fn.Ok(&DAPResponse{Response: response})
	}
	
	return fn.Err[*DAPResponse](
		fmt.Errorf("no mock response for command: %s", command))
}

// TestGetThreads tests the GetThreads function.
func TestGetThreads(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.ThreadsResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  20,
				Type: "response",
			},
			Command:    "threads",
			Success:    true,
			RequestSeq: 20,
		},
		Body: dap.ThreadsResponseBody{
			Threads: []dap.Thread{
				{Id: 1, Name: "main"},
				{Id: 2, Name: "goroutine 2"},
			},
		},
	}
	mockSession.SetResponse("threads", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetThreads
	resp, err := GetThreads(sessionRef)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "threads", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.Threads, 2)
	require.Equal(t, 1, resp.Body.Threads[0].Id)
	require.Equal(t, "main", resp.Body.Threads[0].Name)
	
	// Cleanup
	system.Shutdown()
}

// TestGetThreadsInfo tests the GetThreadsInfo wrapper function.
func TestGetThreadsInfo(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.ThreadsResponse{
		Response: dap.Response{
			Command: "threads",
			Success: true,
		},
		Body: dap.ThreadsResponseBody{
			Threads: []dap.Thread{
				{Id: 1, Name: "main"},
				{Id: 2, Name: "worker"},
				{Id: 3, Name: "background"},
			},
		},
	}
	mockSession.SetResponse("threads", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetThreadsInfo wrapper function
	threads, err := GetThreadsInfo(sessionRef)
	require.NoError(t, err)
	require.Len(t, threads, 3)
	
	// Verify ThreadInfo wrapper types
	require.Equal(t, 1, threads[0].ID)
	require.Equal(t, "main", threads[0].Name)
	require.Equal(t, 2, threads[1].ID)
	require.Equal(t, "worker", threads[1].Name)
	require.Equal(t, 3, threads[2].ID)
	require.Equal(t, "background", threads[2].Name)
	
	// Cleanup
	system.Shutdown()
}

// TestGetStackTrace tests the GetStackTrace function.
func TestGetStackTrace(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.StackTraceResponse{
		Response: dap.Response{
			Command: "stackTrace",
			Success: true,
		},
		Body: dap.StackTraceResponseBody{
			StackFrames: []dap.StackFrame{
				{
					Id:     1,
					Name:   "main",
					Line:   15,
					Column: 5,
					Source: &dap.Source{
						Path: "/path/to/main.go",
						Name: "main.go",
					},
				},
				{
					Id:     2,
					Name:   "processData",
					Line:   42,
					Column: 10,
					Source: &dap.Source{
						Path: "/path/to/utils.go",
						Name: "utils.go",
					},
				},
			},
		},
	}
	mockSession.SetResponse("stackTrace", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetStackTrace
	threadID := 1
	resp, err := GetStackTrace(sessionRef, threadID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "stackTrace", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.StackFrames, 2)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	stackTraceReq, ok := requests[0].(*dap.StackTraceRequest)
	require.True(t, ok)
	require.Equal(t, "stackTrace", stackTraceReq.Command)
	require.Equal(t, threadID, stackTraceReq.Arguments.ThreadId)
	
	// Cleanup
	system.Shutdown()
}

// TestGetStackFrames tests the GetStackFrames wrapper function.
func TestGetStackFrames(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.StackTraceResponse{
		Response: dap.Response{
			Command: "stackTrace",
			Success: true,
		},
		Body: dap.StackTraceResponseBody{
			StackFrames: []dap.StackFrame{
				{
					Id:     1,
					Name:   "main",
					Line:   15,
					Column: 5,
					Source: &dap.Source{
						Path: "/path/to/main.go",
						Name: "main.go",
					},
				},
				{
					Id:     2,
					Name:   "helper",
					Line:   28,
					Column: 1,
					Source: &dap.Source{
						Path: "/path/to/helper.go",
						Name: "helper.go",
					},
				},
			},
		},
	}
	mockSession.SetResponse("stackTrace", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetStackFrames wrapper function
	threadID := 1
	frames, err := GetStackFrames(sessionRef, threadID)
	require.NoError(t, err)
	require.Len(t, frames, 2)
	
	// Verify StackFrame wrapper types
	require.Equal(t, 1, frames[0].ID)
	require.Equal(t, "main", frames[0].Name)
	require.Equal(t, 15, frames[0].Line)
	require.Equal(t, 5, frames[0].Column)
	require.Equal(t, "/path/to/main.go", frames[0].Source.Path)
	require.Equal(t, "main.go", frames[0].Source.Name)
	
	require.Equal(t, 2, frames[1].ID)
	require.Equal(t, "helper", frames[1].Name)
	require.Equal(t, 28, frames[1].Line)
	require.Equal(t, 1, frames[1].Column)
	require.Equal(t, "/path/to/helper.go", frames[1].Source.Path)
	require.Equal(t, "helper.go", frames[1].Source.Name)
	
	// Cleanup
	system.Shutdown()
}

// TestGetScopes tests the GetScopes function.
func TestGetScopes(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.ScopesResponse{
		Response: dap.Response{
			Command: "scopes",
			Success: true,
		},
		Body: dap.ScopesResponseBody{
			Scopes: []dap.Scope{
				{
					Name:               "Local",
					VariablesReference: 1001,
					Expensive:          false,
				},
				{
					Name:               "Global",
					VariablesReference: 1002,
					Expensive:          true,
				},
			},
		},
	}
	mockSession.SetResponse("scopes", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetScopes
	frameID := 1
	resp, err := GetScopes(sessionRef, frameID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "scopes", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.Scopes, 2)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	scopesReq, ok := requests[0].(*dap.ScopesRequest)
	require.True(t, ok)
	require.Equal(t, "scopes", scopesReq.Command)
	require.Equal(t, frameID, scopesReq.Arguments.FrameId)
	
	// Cleanup
	system.Shutdown()
}

// TestGetVariableScopes tests the GetVariableScopes wrapper function.
func TestGetVariableScopes(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.ScopesResponse{
		Response: dap.Response{
			Command: "scopes",
			Success: true,
		},
		Body: dap.ScopesResponseBody{
			Scopes: []dap.Scope{
				{
					Name:               "Arguments",
					VariablesReference: 2001,
					Expensive:          false,
				},
				{
					Name:               "Locals",
					VariablesReference: 2002,
					Expensive:          false,
				},
				{
					Name:               "Globals",
					VariablesReference: 2003,
					Expensive:          true,
				},
			},
		},
	}
	mockSession.SetResponse("scopes", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetVariableScopes wrapper function
	frameID := 1
	scopes, err := GetVariableScopes(sessionRef, frameID)
	require.NoError(t, err)
	require.Len(t, scopes, 3)
	
	// Verify VariableScope wrapper types
	require.Equal(t, "Arguments", scopes[0].Name)
	require.Equal(t, 2001, scopes[0].VariablesReference)
	require.False(t, scopes[0].Expensive)
	
	require.Equal(t, "Locals", scopes[1].Name)
	require.Equal(t, 2002, scopes[1].VariablesReference)
	require.False(t, scopes[1].Expensive)
	
	require.Equal(t, "Globals", scopes[2].Name)
	require.Equal(t, 2003, scopes[2].VariablesReference)
	require.True(t, scopes[2].Expensive)
	
	// Cleanup
	system.Shutdown()
}

// TestGetVariables tests the GetVariables function.
func TestGetVariables(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.VariablesResponse{
		Response: dap.Response{
			Command: "variables",
			Success: true,
		},
		Body: dap.VariablesResponseBody{
			Variables: []dap.Variable{
				{
					Name:               "count",
					Value:              "42",
					Type:               "int",
					VariablesReference: 0,
				},
				{
					Name:               "data",
					Value:              "{...}",
					Type:               "map[string]interface{}",
					VariablesReference: 3001,
					NamedVariables:     3,
				},
			},
		},
	}
	mockSession.SetResponse("variables", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetVariables
	variablesReference := 2001
	resp, err := GetVariables(sessionRef, variablesReference)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "variables", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.Variables, 2)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	variablesReq, ok := requests[0].(*dap.VariablesRequest)
	require.True(t, ok)
	require.Equal(t, "variables", variablesReq.Command)
	require.Equal(t, variablesReference, variablesReq.Arguments.VariablesReference)
	
	// Cleanup
	system.Shutdown()
}

// TestGetVariableList tests the GetVariableList wrapper function.
func TestGetVariableList(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.VariablesResponse{
		Response: dap.Response{
			Command: "variables",
			Success: true,
		},
		Body: dap.VariablesResponseBody{
			Variables: []dap.Variable{
				{
					Name:               "x",
					Value:              "10",
					Type:               "int",
					VariablesReference: 0,
				},
				{
					Name:               "items",
					Value:              "[]string{...}",
					Type:               "[]string",
					VariablesReference: 4001,
					IndexedVariables:   5,
				},
				{
					Name:               "config",
					Value:              "&Config{...}",
					Type:               "*Config",
					VariablesReference: 4002,
					NamedVariables:     8,
				},
			},
		},
	}
	mockSession.SetResponse("variables", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test GetVariableList wrapper function
	variablesReference := 2001
	variables, err := GetVariableList(sessionRef, variablesReference)
	require.NoError(t, err)
	require.Len(t, variables, 3)
	
	// Verify Variable wrapper types
	require.Equal(t, "x", variables[0].Name)
	require.Equal(t, "10", variables[0].Value)
	require.Equal(t, "int", variables[0].Type)
	require.Equal(t, 0, variables[0].VariablesReference)
	require.Equal(t, 0, variables[0].IndexedVariables)
	require.Equal(t, 0, variables[0].NamedVariables)
	
	require.Equal(t, "items", variables[1].Name)
	require.Equal(t, "[]string{...}", variables[1].Value)
	require.Equal(t, "[]string", variables[1].Type)
	require.Equal(t, 4001, variables[1].VariablesReference)
	require.Equal(t, 5, variables[1].IndexedVariables)
	require.Equal(t, 0, variables[1].NamedVariables)
	
	require.Equal(t, "config", variables[2].Name)
	require.Equal(t, "&Config{...}", variables[2].Value)
	require.Equal(t, "*Config", variables[2].Type)
	require.Equal(t, 4002, variables[2].VariablesReference)
	require.Equal(t, 0, variables[2].IndexedVariables)
	require.Equal(t, 8, variables[2].NamedVariables)
	
	// Cleanup
	system.Shutdown()
}

// TestEvaluateExpression tests the EvaluateExpression function.
func TestEvaluateExpression(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.EvaluateResponse{
		Response: dap.Response{
			Command: "evaluate",
			Success: true,
		},
		Body: dap.EvaluateResponseBody{
			Result:             "52",
			Type:               "int",
			VariablesReference: 0,
		},
	}
	mockSession.SetResponse("evaluate", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test EvaluateExpression
	expression := "x + 10"
	frameID := 1
	resp, err := EvaluateExpression(sessionRef, expression, frameID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "evaluate", resp.Command)
	require.True(t, resp.Success)
	require.Equal(t, "52", resp.Body.Result)
	require.Equal(t, "int", resp.Body.Type)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	evaluateReq, ok := requests[0].(*dap.EvaluateRequest)
	require.True(t, ok)
	require.Equal(t, "evaluate", evaluateReq.Command)
	require.Equal(t, expression, evaluateReq.Arguments.Expression)
	require.Equal(t, frameID, evaluateReq.Arguments.FrameId)
	require.Equal(t, "watch", evaluateReq.Arguments.Context)
	
	// Cleanup
	system.Shutdown()
}

// TestEvaluateExpressionResult tests the EvaluateExpressionResult wrapper function.
func TestEvaluateExpressionResult(t *testing.T) {
	// Create a mock inspection session
	mockSession := NewMockInspectionSession()
	
	// Set up the expected response
	expectedResp := &dap.EvaluateResponse{
		Response: dap.Response{
			Command: "evaluate",
			Success: true,
		},
		Body: dap.EvaluateResponseBody{
			Result:             "map[string]interface{}{...}",
			Type:               "map[string]interface{}",
			VariablesReference: 5001,
			NamedVariables:     4,
		},
	}
	mockSession.SetResponse("evaluate", expectedResp)
	
	// Create actor system and register mock session
	system := actor.NewActorSystem()
	sessionKey := actor.NewServiceKey[*DAPRequest, *DAPResponse]("session")
	actor.RegisterWithSystem(
		system, "session", sessionKey,
		actor.NewFunctionBehavior[*DAPRequest, *DAPResponse](
			mockSession.Receive),
	)
	
	// Get reference to the session actor
	sessionRef := actor.FindInReceptionist(
		system.Receptionist(), sessionKey)[0]
	
	// Test EvaluateExpressionResult wrapper function
	expression := "config.GetAll()"
	frameID := 1
	result, err := EvaluateExpressionResult(sessionRef, expression, frameID)
	require.NoError(t, err)
	require.NotNil(t, result)
	
	// Verify EvaluationResult wrapper type
	require.Equal(t, "map[string]interface{}{...}", result.Result)
	require.Equal(t, "map[string]interface{}", result.Type)
	require.Equal(t, 5001, result.VariablesReference)
	require.Equal(t, 0, result.IndexedVariables)
	require.Equal(t, 4, result.NamedVariables)
	
	// Cleanup
	system.Shutdown()
}