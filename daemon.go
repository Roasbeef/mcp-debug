package mcpdebug

import (
	"github.com/lightningnetwork/lnd/actor"
	"github.com/roasbeef/mcp-debug/debugger"
	"github.com/roasbeef/mcp-debug/mcp"
	"github.com/roasbeef/mcp-debug/tui"
)

// MCPDebugService manages the lifecycle of the MCP debug server components.
type MCPDebugService struct {
	actorSystem *actor.ActorSystem
	debuggerKey actor.ServiceKey[*debugger.DebuggerCmd, *debugger.DebuggerResp]
	initialized bool
}

// NewMCPDebugService creates a new MCP debug service.
func NewMCPDebugService() *MCPDebugService {
	return &MCPDebugService{
		actorSystem: actor.NewActorSystem(),
		debuggerKey: actor.NewServiceKey[*debugger.DebuggerCmd, *debugger.DebuggerResp]("debugger"),
		initialized: false,
	}
}

// Start initializes the actor system and registers the debugger actor.
func (s *MCPDebugService) Start() error {
	if s.initialized {
		return nil
	}

	// Create a new debugger actor.
	debuggerActor := debugger.NewDebugger()

	// Register the debugger actor with the actor system.
	actor.RegisterWithSystem(
		s.actorSystem, "debugger", s.debuggerKey, 
		actor.NewFunctionBehavior[*debugger.DebuggerCmd, *debugger.DebuggerResp](debuggerActor.Receive),
	)

	s.initialized = true
	return nil
}

// Stop shuts down the actor system.
func (s *MCPDebugService) Stop() {
	if s.actorSystem != nil {
		s.actorSystem.Shutdown()
	}
}

// GetMCPServer creates a new MCP server with the configured debugger.
func (s *MCPDebugService) GetMCPServer() *mcp.MCPDebugServer {
	if !s.initialized {
		s.Start()
	}
	debuggerRef := actor.FindInReceptionist(s.actorSystem.Receptionist(), s.debuggerKey)[0]
	return mcp.NewMCPDebugServer(s.actorSystem, debuggerRef)
}

// GetActorSystem returns the actor system for direct access.
func (s *MCPDebugService) GetActorSystem() *actor.ActorSystem {
	return s.actorSystem
}

// RunTUI starts the TUI application.
func (s *MCPDebugService) RunTUI() error {
	mcpServer := s.GetMCPServer()
	return tui.RunTUI(mcpServer, s.actorSystem)
}

// Convenience functions for simple usage

// RunTUI runs the TUI application with a new service instance.
func RunTUI() error {
	service := NewMCPDebugService()
	defer service.Stop()
	return service.RunTUI()
}

// NewMCPServer creates a new MCP server with a new service instance.
func NewMCPServer() (*mcp.MCPDebugServer, *MCPDebugService) {
	service := NewMCPDebugService()
	mcpServer := service.GetMCPServer()
	return mcpServer, service
}
