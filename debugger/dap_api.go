package debugger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-dap"
	"github.com/lightningnetwork/lnd/actor"
)

// InitializeSession sends an InitializeRequest to a Session actor and returns 
// the response. This establishes the DAP protocol connection and negotiates
// capabilities between the client and the debug adapter.
func InitializeSession(session actor.ActorRef[*DAPRequest, *DAPResponse], 
	clientID string) (*dap.InitializeResponse, error) {

	req := &dap.InitializeRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  1,
				Type: "request",
			},
			Command: "initialize",
		},
		Arguments: dap.InitializeRequestArguments{
			ClientID:        clientID,
			AdapterID:       "go",
			LinesStartAt1:   true,
			ColumnsStartAt1: true,
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.InitializeResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", 
			result.Response)
	}

	return resp, nil
}

// LaunchProgram launches a Go program for debugging using the provided
// configuration. This provides a high-level interface for program launching
// with comprehensive configuration options.
func LaunchProgram(session actor.ActorRef[*DAPRequest, *DAPResponse], 
	config LaunchConfig) (*dap.LaunchResponse, error) {

	// Determine launch mode based on program path
	// If it's a pre-built binary (ends with .test or is executable), use "exec" mode
	// Otherwise, use "test" or "debug" mode with appropriate build flags
	mode := "debug"
	
	// Check if this is a test file or test binary
	isTest := strings.HasSuffix(config.Program, "_test.go") || 
	          strings.HasSuffix(config.Program, ".test") ||
	          (len(config.Args) > 0 && strings.Contains(config.Args[0], "-test."))
	
	if strings.HasSuffix(config.Program, ".test") || 
	   strings.Contains(config.Program, "__debug_bin") {
		// Pre-built binary - use exec mode
		mode = "exec"
		log.Printf("[LaunchProgram] Using exec mode for pre-built binary: %s", config.Program)
	} else if isTest || strings.HasSuffix(config.Program, "_test.go") {
		// Test file - use test mode
		mode = "test"
		log.Printf("[LaunchProgram] Using test mode for test file: %s", config.Program)
	} else if !strings.HasSuffix(config.Program, ".go") {
		// Check if it's an executable file
		if fileInfo, err := os.Stat(config.Program); err == nil {
			if fileInfo.Mode()&0111 != 0 { // Check if executable
				mode = "exec"
			}
		}
	}
	
	log.Printf("[LaunchProgram] Using mode=%s for program=%s", mode, config.Program)
	
	// Build launch arguments from configuration
	launchArgs := map[string]interface{}{
		"name":    config.Name,
		"type":    "go",
		"request": "launch",
		"mode":    mode,
		"program": config.Program,
	}

	// Add optional configuration
	if len(config.Args) > 0 {
		launchArgs["args"] = config.Args
	}

	if len(config.Env) > 0 {
		launchArgs["env"] = config.Env
	}

	if config.WorkingDir != "" {
		launchArgs["cwd"] = config.WorkingDir
	}

	if config.StopOnEntry {
		launchArgs["stopOnEntry"] = true
	}

	// Handle build flags - automatically add debug flags if not in exec mode
	buildFlags := config.BuildFlags
	if mode != "exec" {
		// For debug and test modes, ensure we have debugging symbols
		hasDebugFlags := false
		for _, flag := range buildFlags {
			if strings.Contains(flag, "-gcflags") || strings.Contains(flag, "-N -l") {
				hasDebugFlags = true
				break
			}
		}
		
		if !hasDebugFlags {
			// Add flags to disable optimizations for better debugging
			debugFlags := []string{"-gcflags", "all=-N -l"}
			buildFlags = append(debugFlags, buildFlags...)
			log.Printf("[LaunchProgram] Auto-adding debug build flags: %v", debugFlags)
		}
	}
	
	if len(buildFlags) > 0 {
		launchArgs["buildFlags"] = buildFlags
	}

	launchArgsJSON, err := json.Marshal(launchArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal launch arguments: %w", 
			err)
	}

	req := &dap.LaunchRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  2,
				Type: "request",
			},
			Command: "launch",
		},
		Arguments: json.RawMessage(launchArgsJSON),
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.LaunchResponse)
	if !ok {
		// Check if it's an error response and extract the message
		if errResp, isErr := result.Response.(*dap.ErrorResponse); isErr {
			return nil, fmt.Errorf("launch failed: %s (id: %d)", 
				errResp.Body.Error.Format, errResp.Body.Error.Id)
		}
		return nil, fmt.Errorf("unexpected response type: %T", 
			result.Response)
	}

	return resp, nil
}

// AttachToProcess attaches the debugger to an existing running process
// using the provided configuration. This provides a high-level interface
// for process attachment with comprehensive configuration options.
func AttachToProcess(session actor.ActorRef[*DAPRequest, *DAPResponse], 
	config AttachConfig) (*dap.AttachResponse, error) {

	// Build attach arguments from configuration
	attachArgs := map[string]interface{}{
		"name":      config.Name,
		"type":      "go",
		"request":   "attach",
		"mode":      config.Mode,
		"processId": config.ProcessID,  // DAP expects "processId" not "pid"
	}

	// Add remote configuration if specified
	if config.Mode == "remote" {
		if config.Host != "" {
			attachArgs["host"] = config.Host
		}
		if config.Port != 0 {
			attachArgs["port"] = config.Port
		}
	}

	attachArgsJSON, err := json.Marshal(attachArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attach arguments: %w", 
			err)
	}

	req := &dap.AttachRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  2,
				Type: "request",
			},
			Command: "attach",
		},
		Arguments: json.RawMessage(attachArgsJSON),
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.AttachResponse)
	if !ok {
		// Check if it's an error response and extract the message
		if errResp, isErr := result.Response.(*dap.ErrorResponse); isErr {
			return nil, fmt.Errorf("attach failed: %s (id: %d)", 
				errResp.Body.Error.Format, errResp.Body.Error.Id)
		}
		return nil, fmt.Errorf("unexpected response type: %T", 
			result.Response)
	}

	return resp, nil
}

// ConfigurationDone indicates that the client has finished sending
// configuration requests and that the debug adapter should begin
// debugging the target program.
func ConfigurationDone(
	session actor.ActorRef[*DAPRequest, *DAPResponse],
) (*dap.ConfigurationDoneResponse, error) {

	req := &dap.ConfigurationDoneRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  3,
				Type: "request",
			},
			Command: "configurationDone",
		},
	}

	dapReq := &DAPRequest{Request: req}
	future := session.Ask(context.Background(), dapReq)
	result, err := future.Await(context.Background()).Unpack()
	if err != nil {
		return nil, err
	}

	resp, ok := result.Response.(*dap.ConfigurationDoneResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", 
			result.Response)
	}

	return resp, nil
}

// SetSourceBreakpoints is a convenience function for setting line-based
// breakpoints in a source file without complex configuration.
func SetSourceBreakpoints(session actor.ActorRef[*DAPRequest, *DAPResponse],
	sourcePath string, lines []int) (*dap.SetBreakpointsResponse, error) {

	// Convert simple line numbers to BreakpointLocation structs
	breakpoints := make([]BreakpointLocation, len(lines))
	for i, line := range lines {
		breakpoints[i] = BreakpointLocation{
			File: sourcePath,
			Line: line,
		}
	}

	return SetBreakpoints(session, breakpoints)
}

// SetSimpleFunctionBreakpoints is a convenience function for setting 
// function breakpoints by name without complex configuration.
func SetSimpleFunctionBreakpoints(
	session actor.ActorRef[*DAPRequest, *DAPResponse],
	functionNames []string) (*dap.SetFunctionBreakpointsResponse, error) {

	// Convert simple function names to FunctionBreakpoint structs
	breakpoints := make([]FunctionBreakpoint, len(functionNames))
	for i, name := range functionNames {
		breakpoints[i] = FunctionBreakpoint{
			Name: name,
		}
	}

	return SetFunctionBreakpoints(session, breakpoints)
}