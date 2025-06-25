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

// MockSession implements a mock DAP session for testing.
type MockSession struct {
	responses map[string]dap.Message
	requests  []dap.Message
}

// NewMockSession creates a new mock session with predefined responses.
func NewMockSession() *MockSession {
	return &MockSession{
		responses: make(map[string]dap.Message),
		requests:  make([]dap.Message, 0),
	}
}

// SetResponse sets a mock response for a given command.
func (m *MockSession) SetResponse(command string, response dap.Message) {
	m.responses[command] = response
}

// GetRequests returns all requests that were sent to the mock session.
func (m *MockSession) GetRequests() []dap.Message {
	return m.requests
}

// Receive implements the actor Receive method for the mock session.
func (m *MockSession) Receive(actorCtx context.Context, 
	msg *DAPRequest) fn.Result[*DAPResponse] {
	
	// Record the request
	m.requests = append(m.requests, msg.Request)
	
	// Extract command from the request
	var command string
	switch req := msg.Request.(type) {
	case *dap.InitializeRequest:
		command = req.Command
	case *dap.LaunchRequest:
		command = req.Command
	case *dap.AttachRequest:
		command = req.Command
	case *dap.ConfigurationDoneRequest:
		command = req.Command
	case *dap.SetBreakpointsRequest:
		command = req.Command
	case *dap.SetFunctionBreakpointsRequest:
		command = req.Command
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
			fmt.Errorf("unknown request type: %T", req))
	}
	
	// Return the predefined response
	if response, exists := m.responses[command]; exists {
		return fn.Ok(&DAPResponse{Response: response})
	}
	
	return fn.Err[*DAPResponse](
		fmt.Errorf("no mock response for command: %s", command))
}

// TestInitializeSession tests the InitializeSession function.
func TestInitializeSession(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.InitializeResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  1,
				Type: "response",
			},
			Command:    "initialize",
			Success:    true,
			RequestSeq: 1,
		},
		Body: dap.Capabilities{
			SupportsConfigurationDoneRequest: true,
			SupportsEvaluateForHovers:        true,
			SupportsStepBack:                 false,
		},
	}
	mockSession.SetResponse("initialize", expectedResp)
	
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
	
	// Test InitializeSession
	resp, err := InitializeSession(sessionRef, "test-client")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "initialize", resp.Command)
	require.True(t, resp.Success)
	require.True(t, resp.Body.SupportsConfigurationDoneRequest)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	initReq, ok := requests[0].(*dap.InitializeRequest)
	require.True(t, ok)
	require.Equal(t, "initialize", initReq.Command)
	require.Equal(t, "test-client", initReq.Arguments.ClientID)
	require.Equal(t, "go", initReq.Arguments.AdapterID)
	require.True(t, initReq.Arguments.LinesStartAt1)
	require.True(t, initReq.Arguments.ColumnsStartAt1)
	
	// Cleanup
	system.Shutdown()
}

// TestLaunchProgram tests the LaunchProgram function.
func TestLaunchProgram(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.LaunchResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  2,
				Type: "response",
			},
			Command:    "launch",
			Success:    true,
			RequestSeq: 2,
		},
	}
	mockSession.SetResponse("launch", expectedResp)
	
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
	
	// Test LaunchProgram with configuration
	config := LaunchConfig{
		Name:        "Test Session",
		Program:     "/path/to/program",
		Args:        []string{"arg1", "arg2"},
		Env:         []string{"VAR1=value1", "VAR2=value2"},
		WorkingDir:  "/tmp",
		StopOnEntry: true,
		BuildFlags:  []string{"-race"},
	}
	
	resp, err := LaunchProgram(sessionRef, config)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "launch", resp.Command)
	require.True(t, resp.Success)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	launchReq, ok := requests[0].(*dap.LaunchRequest)
	require.True(t, ok)
	require.Equal(t, "launch", launchReq.Command)
	
	// Cleanup
	system.Shutdown()
}

// TestAttachToProcess tests the AttachToProcess function.
func TestAttachToProcess(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.AttachResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  3,
				Type: "response",
			},
			Command:    "attach",
			Success:    true,
			RequestSeq: 3,
		},
	}
	mockSession.SetResponse("attach", expectedResp)
	
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
	
	// Test AttachToProcess with configuration
	config := AttachConfig{
		Name:      "Attach Session",
		ProcessID: 12345,
		Mode:      "local",
	}
	
	resp, err := AttachToProcess(sessionRef, config)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "attach", resp.Command)
	require.True(t, resp.Success)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	attachReq, ok := requests[0].(*dap.AttachRequest)
	require.True(t, ok)
	require.Equal(t, "attach", attachReq.Command)
	
	// Cleanup
	system.Shutdown()
}

// TestConfigurationDone tests the ConfigurationDone function.
func TestConfigurationDone(t *testing.T) {
	// Create a mock session
	mockSession := NewMockSession()
	
	// Set up the expected response
	expectedResp := &dap.ConfigurationDoneResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  4,
				Type: "response",
			},
			Command:    "configurationDone",
			Success:    true,
			RequestSeq: 4,
		},
	}
	mockSession.SetResponse("configurationDone", expectedResp)
	
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
	
	// Test ConfigurationDone
	resp, err := ConfigurationDone(sessionRef)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "configurationDone", resp.Command)
	require.True(t, resp.Success)
	
	// Verify the request was sent correctly
	requests := mockSession.GetRequests()
	require.Len(t, requests, 1)
	
	configReq, ok := requests[0].(*dap.ConfigurationDoneRequest)
	require.True(t, ok)
	require.Equal(t, "configurationDone", configReq.Command)
	
	// Cleanup
	system.Shutdown()
}

// TestLaunchConfig validates the LaunchConfig wrapper type.
func TestLaunchConfig(t *testing.T) {
	config := LaunchConfig{
		Name:        "Test Config",
		Program:     "/path/to/program",
		Args:        []string{"arg1", "arg2"},
		Env:         []string{"VAR=value"},
		WorkingDir:  "/tmp",
		StopOnEntry: true,
		BuildFlags:  []string{"-race", "-v"},
	}
	
	// Verify all fields are set correctly
	require.Equal(t, "Test Config", config.Name)
	require.Equal(t, "/path/to/program", config.Program)
	require.Equal(t, []string{"arg1", "arg2"}, config.Args)
	require.Equal(t, []string{"VAR=value"}, config.Env)
	require.Equal(t, "/tmp", config.WorkingDir)
	require.True(t, config.StopOnEntry)
	require.Equal(t, []string{"-race", "-v"}, config.BuildFlags)
}

// TestAttachConfig validates the AttachConfig wrapper type.
func TestAttachConfig(t *testing.T) {
	config := AttachConfig{
		Name:      "Test Attach",
		ProcessID: 12345,
		Mode:      "local",
		Host:      "localhost",
		Port:      8080,
	}
	
	// Verify all fields are set correctly
	require.Equal(t, "Test Attach", config.Name)
	require.Equal(t, 12345, config.ProcessID)
	require.Equal(t, "local", config.Mode)
	require.Equal(t, "localhost", config.Host)
	require.Equal(t, 8080, config.Port)
}