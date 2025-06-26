package debugger

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
	"github.com/lightningnetwork/lnd/fn/v2"
	"github.com/stretchr/testify/require"
)

// MockExecutionSession extends MockSession to handle execution commands.
type MockExecutionSession struct {
	*MockSession
}

// NewMockExecutionSession creates a new mock execution session.
func NewMockExecutionSession() *MockExecutionSession {
	return &MockExecutionSession{
		MockSession: NewMockSession(),
	}
}

// Receive implements the actor Receive method for execution commands.
func (m *MockExecutionSession) Receive(actorCtx context.Context, 
	msg *DAPRequest) fn.Result[*DAPResponse] {
	
	// Record the request
	m.requests = append(m.requests, msg.Request)
	
	// Extract command from the request
	var command string
	switch req := msg.Request.(type) {
	case *dap.ContinueRequest:
		command = req.Command
	case *dap.NextRequest:
		command = req.Command
	case *dap.StepInRequest:
		command = req.Command
	case *dap.StepOutRequest:
		command = req.Command
	case *dap.PauseRequest:
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

// TestContinue tests the Continue function.
func TestContinue(t *testing.T) {
	// Create a mock execution session
	mockSession := NewMockExecutionSession()
	
	// Set up the expected response
	expectedResp := &dap.ContinueResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  10,
				Type: "response",
			},
			Command:    "continue",
			Success:    true,
			RequestSeq: 10,
		},
		Body: dap.ContinueResponseBody{
			AllThreadsContinued: true,
		},
	}
	mockSession.SetResponse("continue", expectedResp)
	
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
	
	// Test Continue
	threadID := 1
	resp, err := Continue(sessionRef, threadID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "continue", resp.Command)
	require.True(t, resp.Success)
	require.True(t, resp.Body.AllThreadsContinued)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	continueReq, ok := requests[0].(*dap.ContinueRequest)
	require.True(t, ok)
	require.Equal(t, "continue", continueReq.Command)
	require.Equal(t, threadID, continueReq.Arguments.ThreadId)
	
	// Cleanup
	system.Shutdown()
}

// TestNext tests the Next function.
func TestNext(t *testing.T) {
	// Create a mock execution session
	mockSession := NewMockExecutionSession()
	
	// Set up the expected response
	expectedResp := &dap.NextResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  11,
				Type: "response",
			},
			Command:    "next",
			Success:    true,
			RequestSeq: 11,
		},
	}
	mockSession.SetResponse("next", expectedResp)
	
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
	
	// Test Next
	threadID := 1
	resp, err := Next(sessionRef, threadID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "next", resp.Command)
	require.True(t, resp.Success)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	nextReq, ok := requests[0].(*dap.NextRequest)
	require.True(t, ok)
	require.Equal(t, "next", nextReq.Command)
	require.Equal(t, threadID, nextReq.Arguments.ThreadId)
	
	// Cleanup
	system.Shutdown()
}

// TestStepIn tests the StepIn function.
func TestStepIn(t *testing.T) {
	// Create a mock execution session
	mockSession := NewMockExecutionSession()
	
	// Set up the expected response
	expectedResp := &dap.StepInResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  12,
				Type: "response",
			},
			Command:    "stepIn",
			Success:    true,
			RequestSeq: 12,
		},
	}
	mockSession.SetResponse("stepIn", expectedResp)
	
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
	
	// Test StepIn
	threadID := 1
	resp, err := StepIn(sessionRef, threadID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "stepIn", resp.Command)
	require.True(t, resp.Success)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	stepInReq, ok := requests[0].(*dap.StepInRequest)
	require.True(t, ok)
	require.Equal(t, "stepIn", stepInReq.Command)
	require.Equal(t, threadID, stepInReq.Arguments.ThreadId)
	
	// Cleanup
	system.Shutdown()
}

// TestStepOut tests the StepOut function.
func TestStepOut(t *testing.T) {
	// Create a mock execution session
	mockSession := NewMockExecutionSession()
	
	// Set up the expected response
	expectedResp := &dap.StepOutResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  13,
				Type: "response",
			},
			Command:    "stepOut",
			Success:    true,
			RequestSeq: 13,
		},
	}
	mockSession.SetResponse("stepOut", expectedResp)
	
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
	
	// Test StepOut
	threadID := 1
	resp, err := StepOut(sessionRef, threadID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "stepOut", resp.Command)
	require.True(t, resp.Success)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	stepOutReq, ok := requests[0].(*dap.StepOutRequest)
	require.True(t, ok)
	require.Equal(t, "stepOut", stepOutReq.Command)
	require.Equal(t, threadID, stepOutReq.Arguments.ThreadId)
	
	// Cleanup
	system.Shutdown()
}

// TestPause tests the Pause function.
func TestPause(t *testing.T) {
	// Create a mock execution session
	mockSession := NewMockExecutionSession()
	
	// Set up the expected response
	expectedResp := &dap.PauseResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  14,
				Type: "response",
			},
			Command:    "pause",
			Success:    true,
			RequestSeq: 14,
		},
	}
	mockSession.SetResponse("pause", expectedResp)
	
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
	
	// Test Pause
	threadID := 1
	resp, err := Pause(sessionRef, threadID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "pause", resp.Command)
	require.True(t, resp.Success)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	pauseReq, ok := requests[0].(*dap.PauseRequest)
	require.True(t, ok)
	require.Equal(t, "pause", pauseReq.Command)
	require.Equal(t, threadID, pauseReq.Arguments.ThreadId)
	
	// Cleanup
	system.Shutdown()
}

// TestExecutionCommands tests multiple execution commands in sequence.
func TestExecutionCommands(t *testing.T) {
	// Create a mock execution session
	mockSession := NewMockExecutionSession()
	
	// Set up responses for all execution commands
	mockSession.SetResponse("continue", &dap.ContinueResponse{
		Response: dap.Response{Command: "continue", Success: true},
		Body:     dap.ContinueResponseBody{AllThreadsContinued: true},
	})
	mockSession.SetResponse("next", &dap.NextResponse{
		Response: dap.Response{Command: "next", Success: true},
	})
	mockSession.SetResponse("stepIn", &dap.StepInResponse{
		Response: dap.Response{Command: "stepIn", Success: true},
	})
	mockSession.SetResponse("stepOut", &dap.StepOutResponse{
		Response: dap.Response{Command: "stepOut", Success: true},
	})
	mockSession.SetResponse("pause", &dap.PauseResponse{
		Response: dap.Response{Command: "pause", Success: true},
	})
	
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
	
	threadID := 1
	
	// Test Continue
	continueResp, err := Continue(sessionRef, threadID)
	require.NoError(t, err)
	require.True(t, continueResp.Success)
	require.True(t, continueResp.Body.AllThreadsContinued)
	
	// Test Next
	nextResp, err := Next(sessionRef, threadID)
	require.NoError(t, err)
	require.True(t, nextResp.Success)
	
	// Test StepIn
	stepInResp, err := StepIn(sessionRef, threadID)
	require.NoError(t, err)
	require.True(t, stepInResp.Success)
	
	// Test StepOut
	stepOutResp, err := StepOut(sessionRef, threadID)
	require.NoError(t, err)
	require.True(t, stepOutResp.Success)
	
	// Test Pause
	pauseResp, err := Pause(sessionRef, threadID)
	require.NoError(t, err)
	require.True(t, pauseResp.Success)
	
	// Verify all requests were sent
	requests := mockSession.GetRequests()
	require.Len(t, requests, 5)
	
	// Verify command sequence
	expectedCommands := []string{"continue", "next", "stepIn", "stepOut", "pause"}
	for i, req := range requests {
		switch typedReq := req.(type) {
		case *dap.ContinueRequest:
			require.Equal(t, expectedCommands[i], typedReq.Command)
			require.Equal(t, threadID, typedReq.Arguments.ThreadId)
		case *dap.NextRequest:
			require.Equal(t, expectedCommands[i], typedReq.Command)
			require.Equal(t, threadID, typedReq.Arguments.ThreadId)
		case *dap.StepInRequest:
			require.Equal(t, expectedCommands[i], typedReq.Command)
			require.Equal(t, threadID, typedReq.Arguments.ThreadId)
		case *dap.StepOutRequest:
			require.Equal(t, expectedCommands[i], typedReq.Command)
			require.Equal(t, threadID, typedReq.Arguments.ThreadId)
		case *dap.PauseRequest:
			require.Equal(t, expectedCommands[i], typedReq.Command)
			require.Equal(t, threadID, typedReq.Arguments.ThreadId)
		default:
			t.Fatalf("unexpected request type: %T", req)
		}
	}
	
	// Cleanup
	system.Shutdown()
}