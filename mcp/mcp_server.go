package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lightningnetwork/lnd/actor"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/roasbeef/mcp-debug/debugger"
)

// CreateSessionArgs represents the arguments for creating a debug session.
type CreateSessionArgs struct {
	SessionID string `json:"session_id"`
}

// InitializeSessionArgs represents the arguments for initializing a session.
type InitializeSessionArgs struct {
	SessionID string `json:"session_id"`
	ClientID  string `json:"client_id"`
}

// LaunchProgramArgs represents the arguments for launching a program.
type LaunchProgramArgs struct {
	SessionID   string   `json:"session_id"`
	Program     string   `json:"program"`
	Name        string   `json:"name,omitempty"`
	Args        []string `json:"args,omitempty"`
	Env         []string `json:"env,omitempty"`
	WorkingDir  string   `json:"working_dir,omitempty"`
	StopOnEntry bool     `json:"stop_on_entry,omitempty"`
	BuildFlags  []string `json:"build_flags,omitempty"`
}

// SetBreakpointsArgs represents the arguments for setting breakpoints.
type SetBreakpointsArgs struct {
	SessionID string `json:"session_id"`
	File      string `json:"file"`
	Lines     []int  `json:"lines"`
}

// ExecutionControlArgs represents the arguments for execution control commands.
type ExecutionControlArgs struct {
	SessionID string `json:"session_id"`
	ThreadID  int    `json:"thread_id"`
}

// GetThreadsArgs represents the arguments for getting threads.
type GetThreadsArgs struct {
	SessionID string `json:"session_id"`
}

// GetStackFramesArgs represents the arguments for getting stack frames.
type GetStackFramesArgs struct {
	SessionID string `json:"session_id"`
	ThreadID  int    `json:"thread_id"`
}

// GetVariablesArgs represents the arguments for getting variables.
type GetVariablesArgs struct {
	SessionID string `json:"session_id"`
	FrameID   int    `json:"frame_id"`
}

// EvaluateExpressionArgs represents the arguments for evaluating expressions.
type EvaluateExpressionArgs struct {
	SessionID  string `json:"session_id"`
	Expression string `json:"expression"`
	FrameID    int    `json:"frame_id"`
}

// MCPDebugServer wraps our debugging functionality as an MCP server.
type MCPDebugServer struct {
	server   *server.MCPServer
	debugger actor.ActorRef[*debugger.DebuggerCmd, *debugger.DebuggerResp]
	sessions map[string]actor.ActorRef[*debugger.DAPRequest, *debugger.DAPResponse]
	actorSys *actor.ActorSystem
}

// NewMCPDebugServer creates a new MCP server for debugging operations.
func NewMCPDebugServer(actorSys *actor.ActorSystem,
	debuggerRef actor.ActorRef[*debugger.DebuggerCmd, *debugger.DebuggerResp]) *MCPDebugServer {

	mcpServer := server.NewMCPServer(
		"Go Debug Adapter Protocol Server",
		"1.0.0",
	)

	mds := &MCPDebugServer{
		server:   mcpServer,
		debugger: debuggerRef,
		sessions: make(map[string]actor.ActorRef[*debugger.DAPRequest, *debugger.DAPResponse]),
		actorSys: actorSys,
	}

	// Register all debugging tools
	mds.registerTools()

	return mds
}

// registerTools registers all available debugging tools with the MCP server.
func (mds *MCPDebugServer) registerTools() {
	// Session management tools
	mds.registerCreateSessionTool()
	mds.registerInitializeSessionTool()

	// Program control tools
	mds.registerLaunchProgramTool()
	mds.registerConfigurationDoneTool()

	// Breakpoint tools
	mds.registerSetBreakpointsTool()

	// Execution control tools
	mds.registerContinueTool()
	mds.registerNextTool()
	mds.registerStepInTool()
	mds.registerStepOutTool()
	mds.registerPauseTool()

	// Inspection tools
	mds.registerGetThreadsTool()
	mds.registerGetStackFramesTool()
	mds.registerGetVariablesTool()
	mds.registerEvaluateExpressionTool()
}

// registerCreateSessionTool registers the create debugging session tool.
func (mds *MCPDebugServer) registerCreateSessionTool() {
	tool := mcp.NewTool("create_debug_session",
		mcp.WithDescription("Create a new debugging session"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("Unique identifier for the session")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args CreateSessionArgs) (*mcp.CallToolResult, error) {

		sessionID := args.SessionID

		// Check if session already exists
		if _, exists := mds.sessions[sessionID]; exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s already exists", sessionID)),
				},
				IsError: true,
			}, nil
		}

		// Create debug session
		cmd := &debugger.CreateSessionCmd{}
		future := mds.debugger.Ask(ctx, &debugger.DebuggerCmd{Cmd: cmd})
		result, err := future.Await(ctx).Unpack()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to create session: %v", err)),
				},
				IsError: true,
			}, nil
		}

		// Store session reference
		createResp, ok := result.Resp.(*debugger.CreateSessionResp)
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Invalid response from debugger"),
				},
				IsError: true,
			}, nil
		}

		mds.sessions[sessionID] = createResp.Session

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Created debugging session: %s", sessionID)),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

// registerInitializeSessionTool registers the initialize session tool.
func (mds *MCPDebugServer) registerInitializeSessionTool() {
	tool := mcp.NewTool("initialize_session",
		mcp.WithDescription("Initialize a debugging session with DAP protocol"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithString("client_id", mcp.Required(),
			mcp.Description("Client identifier")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args InitializeSessionArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		// Initialize the session
		resp, err := debugger.InitializeSession(session, args.ClientID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to initialize session: %v", err)),
				},
				IsError: true,
			}, nil
		}

		respJSON, _ := json.Marshal(resp)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Session initialized successfully. Response: %s",
					string(respJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

// registerLaunchProgramTool registers the launch program tool.
func (mds *MCPDebugServer) registerLaunchProgramTool() {
	tool := mcp.NewTool("launch_program",
		mcp.WithDescription("Launch a Go program for debugging"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithString("program", mcp.Required(),
			mcp.Description("Path to the Go program to debug")),
		mcp.WithString("name",
			mcp.Description("Name for the debug session")),
		mcp.WithArray("args",
			mcp.Description("Command line arguments for the program"),
			mcp.Items(map[string]any{"type": "string"})),
		mcp.WithArray("env",
			mcp.Description("Environment variables (KEY=value format)"),
			mcp.Items(map[string]any{"type": "string"})),
		mcp.WithString("working_dir",
			mcp.Description("Working directory for the program")),
		mcp.WithBoolean("stop_on_entry",
			mcp.Description("Stop at program entry point")),
		mcp.WithArray("build_flags",
			mcp.Description("Go build flags"),
			mcp.Items(map[string]any{"type": "string"})),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args LaunchProgramArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		// Build launch configuration
		config := debugger.LaunchConfig{
			Name:        getStringOrDefault(args.Name, "Debug Session"),
			Program:     args.Program,
			Args:        args.Args,
			Env:         args.Env,
			WorkingDir:  args.WorkingDir,
			StopOnEntry: args.StopOnEntry,
			BuildFlags:  args.BuildFlags,
		}

		// Launch the program
		resp, err := debugger.LaunchProgram(session, config)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to launch program: %v", err)),
				},
				IsError: true,
			}, nil
		}

		respJSON, _ := json.Marshal(resp)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Program launched successfully. Response: %s",
					string(respJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

// registerConfigurationDoneTool registers the configuration done tool.
func (mds *MCPDebugServer) registerConfigurationDoneTool() {
	tool := mcp.NewTool("configuration_done",
		mcp.WithDescription("Signal that configuration is complete and debugging can begin"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args GetThreadsArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		// Send configuration done
		resp, err := debugger.ConfigurationDone(session)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to send configuration done: %v", err)),
				},
				IsError: true,
			}, nil
		}

		respJSON, _ := json.Marshal(resp)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Configuration done. Response: %s", string(respJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

// registerSetBreakpointsTool registers the set breakpoints tool.
func (mds *MCPDebugServer) registerSetBreakpointsTool() {
	tool := mcp.NewTool("set_breakpoints",
		mcp.WithDescription("Set breakpoints in source code"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithString("file", mcp.Required(),
			mcp.Description("Source file path")),
		mcp.WithArray("lines", mcp.Required(),
			mcp.Description("Line numbers for breakpoints"),
			mcp.Items(map[string]any{"type": "integer"})),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args SetBreakpointsArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		// Set breakpoints
		resp, err := debugger.SetSourceBreakpoints(session, args.File, args.Lines)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to set breakpoints: %v", err)),
				},
				IsError: true,
			}, nil
		}

		respJSON, _ := json.Marshal(resp.Body.Breakpoints)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Breakpoints set successfully. Breakpoints: %s",
					string(respJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

// registerContinueTool registers the continue execution tool.
func (mds *MCPDebugServer) registerContinueTool() {
	tool := mcp.NewTool("continue_execution",
		mcp.WithDescription("Continue program execution"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithNumber("thread_id", mcp.Required(),
			mcp.Description("Thread ID to continue")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args ExecutionControlArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		// Continue execution
		resp, err := debugger.Continue(session, args.ThreadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to continue execution: %v", err)),
				},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Continued execution. All threads continued: %t",
					resp.Body.AllThreadsContinued)),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

// registerGetThreadsTool registers the get threads tool.
func (mds *MCPDebugServer) registerGetThreadsTool() {
	tool := mcp.NewTool("get_threads",
		mcp.WithDescription("Get information about all threads in the debugged program"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args GetThreadsArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		// Get threads
		threads, err := debugger.GetThreadsInfo(session)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to get threads: %v", err)),
				},
				IsError: true,
			}, nil
		}

		threadsJSON, _ := json.Marshal(threads)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Threads: %s", string(threadsJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

// Serve starts the MCP server using stdio transport.
func (mds *MCPDebugServer) Serve() error {
	log.Printf("Starting Go DAP MCP Server...")
	return server.ServeStdio(mds.server)
}

// Helper functions

// getStringOrDefault returns the string value or a default if empty.
func getStringOrDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

// Placeholder implementations for remaining tools to keep file manageable
func (mds *MCPDebugServer) registerNextTool() {
	tool := mcp.NewTool("step_next",
		mcp.WithDescription("Step over (execute next line without entering function calls)"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithNumber("thread_id", mcp.Required(),
			mcp.Description("Thread ID to step")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args ExecutionControlArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		_, err := debugger.Next(session, args.ThreadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to step next: %v", err)),
				},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Stepped to next line successfully"),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

func (mds *MCPDebugServer) registerStepInTool() {
	tool := mcp.NewTool("step_in",
		mcp.WithDescription("Step into function calls"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithNumber("thread_id", mcp.Required(),
			mcp.Description("Thread ID to step")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args ExecutionControlArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		_, err := debugger.StepIn(session, args.ThreadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to step in: %v", err)),
				},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Stepped into function successfully"),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

func (mds *MCPDebugServer) registerStepOutTool() {
	tool := mcp.NewTool("step_out",
		mcp.WithDescription("Step out of current function"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithNumber("thread_id", mcp.Required(),
			mcp.Description("Thread ID to step")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args ExecutionControlArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		_, err := debugger.StepOut(session, args.ThreadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to step out: %v", err)),
				},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Stepped out of function successfully"),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

func (mds *MCPDebugServer) registerPauseTool() {
	tool := mcp.NewTool("pause_execution",
		mcp.WithDescription("Pause program execution"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithNumber("thread_id", mcp.Required(),
			mcp.Description("Thread ID to pause")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args ExecutionControlArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		_, err := debugger.Pause(session, args.ThreadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to pause execution: %v", err)),
				},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Paused execution successfully"),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

func (mds *MCPDebugServer) registerGetStackFramesTool() {
	tool := mcp.NewTool("get_stack_frames",
		mcp.WithDescription("Get stack frames for a specific thread"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithNumber("thread_id", mcp.Required(),
			mcp.Description("Thread ID to get stack frames for")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args GetStackFramesArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		frames, err := debugger.GetStackFrames(session, args.ThreadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to get stack frames: %v", err)),
				},
				IsError: true,
			}, nil
		}

		framesJSON, _ := json.Marshal(frames)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Stack frames: %s", string(framesJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

func (mds *MCPDebugServer) registerGetVariablesTool() {
	tool := mcp.NewTool("get_variables",
		mcp.WithDescription("Get variables for a specific scope"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithNumber("frame_id", mcp.Required(),
			mcp.Description("Frame ID to get variables for")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args GetVariablesArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		scopes, err := debugger.GetVariableScopes(session, args.FrameID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to get variable scopes: %v", err)),
				},
				IsError: true,
			}, nil
		}

		allVariables := make(map[string][]debugger.Variable)
		for _, scope := range scopes {
			variables, err := debugger.GetVariableList(session, scope.VariablesReference)
			if err != nil {
				continue
			}
			allVariables[scope.Name] = variables
		}

		variablesJSON, _ := json.Marshal(allVariables)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Variables: %s", string(variablesJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}

func (mds *MCPDebugServer) registerEvaluateExpressionTool() {
	tool := mcp.NewTool("evaluate_expression",
		mcp.WithDescription("Evaluate an expression in the context of a specific frame"),
		mcp.WithString("session_id", mcp.Required(),
			mcp.Description("Session identifier")),
		mcp.WithString("expression", mcp.Required(),
			mcp.Description("Expression to evaluate")),
		mcp.WithNumber("frame_id", mcp.Required(),
			mcp.Description("Frame ID for evaluation context")),
	)

	handler := mcp.NewTypedToolHandler(func(ctx context.Context,
		request mcp.CallToolRequest, args EvaluateExpressionArgs) (*mcp.CallToolResult, error) {

		session, exists := mds.sessions[args.SessionID]
		if !exists {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Session %s not found", args.SessionID)),
				},
				IsError: true,
			}, nil
		}

		result, err := debugger.EvaluateExpressionResult(session, args.Expression, args.FrameID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf(
						"Failed to evaluate expression: %v", err)),
				},
				IsError: true,
			}, nil
		}

		resultJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf(
					"Evaluation result: %s", string(resultJSON))),
			},
		}, nil
	})

	mds.server.AddTool(tool, handler)
}




// GetSessions returns a copy of the current sessions map for monitoring.
func (mds *MCPDebugServer) GetSessions() map[string]actor.ActorRef[*debugger.DAPRequest, *debugger.DAPResponse] {
	sessionsCopy := make(map[string]actor.ActorRef[*debugger.DAPRequest, *debugger.DAPResponse])
	for k, v := range mds.sessions {
		sessionsCopy[k] = v
	}
	return sessionsCopy
}
