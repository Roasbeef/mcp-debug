package debugger

import (
	"testing"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
	"github.com/stretchr/testify/require"
)

// TestSetBreakpoints tests the SetBreakpoints function with wrapper types.
func TestSetBreakpoints(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.SetBreakpointsResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  5,
				Type: "response",
			},
			Command:    "setBreakpoints",
			Success:    true,
			RequestSeq: 5,
		},
		Body: dap.SetBreakpointsResponseBody{
			Breakpoints: []dap.Breakpoint{
				{
					Id:       1,
					Verified: true,
					Line:     10,
					Source: &dap.Source{
						Path: "/path/to/file.go",
					},
				},
				{
					Id:       2,
					Verified: true,
					Line:     20,
					Column:   5,
					Source: &dap.Source{
						Path: "/path/to/file.go",
					},
				},
			},
		},
	}
	mockSession.SetResponse("setBreakpoints", expectedResp)
	
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
	
	// Test SetBreakpoints with BreakpointLocation wrapper types
	breakpoints := []BreakpointLocation{
		{
			File: "/path/to/file.go",
			Line: 10,
		},
		{
			File:         "/path/to/file.go",
			Line:         20,
			Column:       5,
			Condition:    "x > 0",
			HitCondition: ">= 3",
			LogMessage:   "Hit breakpoint at line 20",
		},
	}
	
	resp, err := SetBreakpoints(sessionRef, breakpoints)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "setBreakpoints", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.Breakpoints, 2)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	setBreakpointsReq, ok := requests[0].(*dap.SetBreakpointsRequest)
	require.True(t, ok)
	require.Equal(t, "setBreakpoints", setBreakpointsReq.Command)
	require.Equal(t, "/path/to/file.go", setBreakpointsReq.Arguments.Source.Path)
	require.Len(t, setBreakpointsReq.Arguments.Breakpoints, 2)
	
	// Check first breakpoint (simple line breakpoint)
	bp1 := setBreakpointsReq.Arguments.Breakpoints[0]
	require.Equal(t, 10, bp1.Line)
	require.Equal(t, 0, bp1.Column) // Should be 0 for line-only breakpoint
	require.Empty(t, bp1.Condition)
	require.Empty(t, bp1.HitCondition)
	require.Empty(t, bp1.LogMessage)
	
	// Check second breakpoint (full configuration)
	bp2 := setBreakpointsReq.Arguments.Breakpoints[1]
	require.Equal(t, 20, bp2.Line)
	require.Equal(t, 5, bp2.Column)
	require.Equal(t, "x > 0", bp2.Condition)
	require.Equal(t, ">= 3", bp2.HitCondition)
	require.Equal(t, "Hit breakpoint at line 20", bp2.LogMessage)
	
	// Cleanup
	system.Shutdown()
}

// TestSetBreakpointsValidation tests that SetBreakpoints validates input.
func TestSetBreakpointsValidation(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
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
	
	// Test with empty breakpoints slice
	_, err := SetBreakpoints(sessionRef, []BreakpointLocation{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no breakpoints provided")
	
	// Test with breakpoints for different files
	breakpoints := []BreakpointLocation{
		{File: "/path/to/file1.go", Line: 10},
		{File: "/path/to/file2.go", Line: 20},
	}
	_, err = SetBreakpoints(sessionRef, breakpoints)
	require.Error(t, err)
	require.Contains(t, err.Error(), "all breakpoints must be for the same file")
	
	// Cleanup
	system.Shutdown()
}

// TestSetFunctionBreakpoints tests the SetFunctionBreakpoints function.
func TestSetFunctionBreakpoints(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.SetFunctionBreakpointsResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  6,
				Type: "response",
			},
			Command:    "setFunctionBreakpoints",
			Success:    true,
			RequestSeq: 6,
		},
		Body: dap.SetFunctionBreakpointsResponseBody{
			Breakpoints: []dap.Breakpoint{
				{
					Id:       3,
					Verified: true,
				},
				{
					Id:       4,
					Verified: true,
				},
			},
		},
	}
	mockSession.SetResponse("setFunctionBreakpoints", expectedResp)
	
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
	
	// Test SetFunctionBreakpoints with FunctionBreakpoint wrapper types
	functionBreakpoints := []FunctionBreakpoint{
		{
			Name: "main",
		},
		{
			Name:         "processData",
			Condition:    "count > 10",
			HitCondition: "== 5",
		},
	}
	
	resp, err := SetFunctionBreakpoints(sessionRef, functionBreakpoints)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "setFunctionBreakpoints", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.Breakpoints, 2)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	setFuncBreakpointsReq, ok := requests[0].(*dap.SetFunctionBreakpointsRequest)
	require.True(t, ok)
	require.Equal(t, "setFunctionBreakpoints", setFuncBreakpointsReq.Command)
	require.Len(t, setFuncBreakpointsReq.Arguments.Breakpoints, 2)
	
	// Check first function breakpoint (simple)
	fbp1 := setFuncBreakpointsReq.Arguments.Breakpoints[0]
	require.Equal(t, "main", fbp1.Name)
	require.Empty(t, fbp1.Condition)
	require.Empty(t, fbp1.HitCondition)
	
	// Check second function breakpoint (with conditions)
	fbp2 := setFuncBreakpointsReq.Arguments.Breakpoints[1]
	require.Equal(t, "processData", fbp2.Name)
	require.Equal(t, "count > 10", fbp2.Condition)
	require.Equal(t, "== 5", fbp2.HitCondition)
	
	// Cleanup
	system.Shutdown()
}

// TestSetSourceBreakpoints tests the convenience function for simple line breakpoints.
func TestSetSourceBreakpoints(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.SetBreakpointsResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  7,
				Type: "response",
			},
			Command:    "setBreakpoints",
			Success:    true,
			RequestSeq: 7,
		},
		Body: dap.SetBreakpointsResponseBody{
			Breakpoints: []dap.Breakpoint{
				{Id: 5, Verified: true, Line: 10},
				{Id: 6, Verified: true, Line: 20},
				{Id: 7, Verified: true, Line: 30},
			},
		},
	}
	mockSession.SetResponse("setBreakpoints", expectedResp)
	
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
	
	// Test SetSourceBreakpoints convenience function
	sourcePath := "/path/to/main.go"
	lines := []int{10, 20, 30}
	
	resp, err := SetSourceBreakpoints(sessionRef, sourcePath, lines)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "setBreakpoints", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.Breakpoints, 3)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	setBreakpointsReq, ok := requests[0].(*dap.SetBreakpointsRequest)
	require.True(t, ok)
	require.Equal(t, sourcePath, setBreakpointsReq.Arguments.Source.Path)
	require.Len(t, setBreakpointsReq.Arguments.Breakpoints, 3)
	
	// Verify line numbers
	for i, expectedLine := range lines {
		require.Equal(t, expectedLine, 
			setBreakpointsReq.Arguments.Breakpoints[i].Line)
	}
	
	// Cleanup
	system.Shutdown()
}

// TestSetSimpleFunctionBreakpoints tests the convenience function for simple function breakpoints.
func TestSetSimpleFunctionBreakpoints(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.SetFunctionBreakpointsResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  8,
				Type: "response",
			},
			Command:    "setFunctionBreakpoints",
			Success:    true,
			RequestSeq: 8,
		},
		Body: dap.SetFunctionBreakpointsResponseBody{
			Breakpoints: []dap.Breakpoint{
				{Id: 8, Verified: true},
				{Id: 9, Verified: true},
			},
		},
	}
	mockSession.SetResponse("setFunctionBreakpoints", expectedResp)
	
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
	
	// Test SetSimpleFunctionBreakpoints convenience function
	functionNames := []string{"main", "processData"}
	
	resp, err := SetSimpleFunctionBreakpoints(sessionRef, functionNames)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "setFunctionBreakpoints", resp.Command)
	require.True(t, resp.Success)
	require.Len(t, resp.Body.Breakpoints, 2)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	setFuncBreakpointsReq, ok := requests[0].(*dap.SetFunctionBreakpointsRequest)
	require.True(t, ok)
	require.Len(t, setFuncBreakpointsReq.Arguments.Breakpoints, 2)
	
	// Verify function names
	for i, expectedName := range functionNames {
		require.Equal(t, expectedName, 
			setFuncBreakpointsReq.Arguments.Breakpoints[i].Name)
	}
	
	// Cleanup
	system.Shutdown()
}

// TestBreakpointLocation validates the BreakpointLocation wrapper type.
func TestBreakpointLocation(t *testing.T) {
	bp := BreakpointLocation{
		File:         "/path/to/file.go",
		Line:         42,
		Column:       10,
		Condition:    "x > 5",
		HitCondition: ">= 2",
		LogMessage:   "Breakpoint hit at line 42",
	}
	
	// Verify all fields are set correctly
	require.Equal(t, "/path/to/file.go", bp.File)
	require.Equal(t, 42, bp.Line)
	require.Equal(t, 10, bp.Column)
	require.Equal(t, "x > 5", bp.Condition)
	require.Equal(t, ">= 2", bp.HitCondition)
	require.Equal(t, "Breakpoint hit at line 42", bp.LogMessage)
}

// TestFunctionBreakpoint validates the FunctionBreakpoint wrapper type.
func TestFunctionBreakpoint(t *testing.T) {
	fbp := FunctionBreakpoint{
		Name:         "calculateSum",
		Condition:    "count > 0",
		HitCondition: "== 1",
	}
	
	// Verify all fields are set correctly
	require.Equal(t, "calculateSum", fbp.Name)
	require.Equal(t, "count > 0", fbp.Condition)
	require.Equal(t, "== 1", fbp.HitCondition)
}