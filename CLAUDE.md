# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an MCP (Model Context Protocol) debugger server that exposes debugging capabilities for Go applications as standardized MCP tools. The server integrates the Debug Adapter Protocol (DAP) with Delve debugger through an actor-based architecture using the Lightning Network's actor system. LLM clients can use this server to perform AI-powered debugging workflows.

**🎯 Current Status**: Production-ready MCP debugger with comprehensive TUI, clean package architecture, and real-time metrics tracking.

## Package Architecture

The project is organized into focused packages with clean separation of concerns:

```
mcp-debug/
├── agent_planning/   📚 Documentation & implementation notes  
├── debugger/         🔧 DAP/Delve integration with actor system
├── mcp/             🌐 MCP server exposing debugging as AI tools  
├── tui/             🖥️ Bubble Tea TUI with real-time monitoring
├── cmd/             📦 Production command-line applications
│   ├── dlv-mcp-server/ 🔌 Headless MCP server for API integration
│   └── tui/         💻 Interactive TUI console (main interface)
├── internal/        🔒 Internal packages
│   └── test/        🧪 Development & validation utilities
├── examples/        📖 Example programs for debugging
├── CLAUDE.md        📋 Claude Code guidance (this file)
└── daemon.go        🏗️ Clean API with lifecycle management
```

### Package Responsibilities

- **agent_planning/**: Documentation archive (implementation notes, design decisions)
- **debugger/**: DAP protocol, Delve integration, actor message handling
- **mcp/**: MCP server with 15 debugging tools for AI clients
- **tui/**: Bubble Tea terminal interface with real-time monitoring and metrics
- **cmd/dlv-mcp-server/**: Headless MCP server for API/LLM integration
- **cmd/tui/**: Interactive TUI console with dashboard, sessions, clients, commands, and logs
- **internal/test/**: Development validation tools (tui-validation)
- **examples/**: Sample Go programs for debugging demonstrations

### Core Service API

The system now uses `MCPDebugService` for proper lifecycle management:

```go
// Service-oriented approach (NEW)
service := mcpdebug.NewMCPDebugService()
defer service.Stop()
mcpServer := service.GetMCPServer()

// Convenience functions
mcpdebug.RunTUI()                    // Interactive TUI
mcpServer, service := mcpdebug.NewMCPServer()  // Headless server
```

## Development Commands

### Building
```bash
# Interactive TUI console (recommended for monitoring)
go build -o tui-console ./cmd/tui
./tui-console

# Headless MCP server (for AI/LLM integration)
go build -o dlv-mcp-server ./cmd/dlv-mcp-server
./dlv-mcp-server

# Development validation tool
go build -o tui-validation ./internal/test/tui-validation
./tui-validation

# Build and test example programs
go build -o simple ./examples/simple
./simple  # Can be used for debugging tests
```

### Testing
```bash
# Run all tests
go test ./...

# Test specific packages
go test ./debugger/...
go test ./mcp/...
go test ./tui/...

# Run with verbose output
go test -v ./...

# Test with coverage
go test -cover ./...
```

### MCP Server Testing
```bash
# Test MCP server functionality
cd mcp && ./test_mcp_server.sh
```

### Code Quality
```bash
# Format code
go fmt ./...

# Lint (if golangci-lint is available)
golangci-lint run

# Vet code
go vet ./...
```

## Actor System Architecture

The project uses LND's actor system with router pattern for load balancing:

### Package-Specific Actor Components

**debugger/** package contains:
- `messages.go`: `DebuggerCmd`, `DebuggerResp`, actor message types
- `debugger.go`: Main debugger actor implementation
- `session.go`: Session-specific actors for debug instances
- `dap_messages.go`: `DAPRequest`, `DAPResponse` for protocol handling

### Actor Router Pattern (Used in TUI)
- Round-robin load balancing across debugger instances
- Automatic failover and recovery
- Efficient message distribution

### Message Types (in debugger/ package)
- All messages implement `actor.Message` interface
- Embed `actor.BaseMessage` for convenience
- Define `MessageType()` method for debugging
- Exported types: `DebuggerCmd`, `DebuggerResp`, `DAPRequest`, `DAPResponse`

### Service Keys (managed by MCPDebugService)
- Type-safe identifiers: `actor.NewServiceKey[*debugger.DebuggerCmd, *debugger.DebuggerResp]`
- Main key: `DebuggerKey` in service initialization
- Session keys: Created dynamically as needed

### Actor Lifecycle (Service-Managed)
- Actors managed by `MCPDebugService` in `daemon.go`
- Automatic initialization with `service.Start()`
- Clean shutdown with `service.Stop()`
- No more global variables or manual system management

### Communication Patterns
- **Tell**: Fire-and-forget messages via `ActorRef.Tell()`
- **Ask**: Request-response via `ActorRef.Ask()` returning `Future[R]`
- **Router Pattern**: Round-robin load balancing for multiple debugger instances

## Package Integration Points

- **mcp/**: Uses `debugger.DebuggerCmd` and `debugger.DebuggerResp` for actor communication
- **tui/**: 
  - Uses `mcp.MCPDebugServer.GetSessions()` for real-time session monitoring
  - Implements Bubble Tea components (table, viewport, textinput, help)
  - Tracks real metrics (uptime, requests, errors)
- **cmd/**: Uses `mcpdebug.RunTUI()` and `mcpdebug.NewMCPServer()` convenience functions

## Code Style Guidelines

This project follows LND's development guidelines:

### Formatting
- 80 character line limit
- Tabs as 8 spaces
- Logical code stanzas with spacing
- Function parameters wrapped with closing paren on new line

### Documentation
- All functions must have comments starting with function name
- Complete sentences for godoc compatibility
- Explain intention, not just implementation

### Git Commits
- Prefix commits with affected package (e.g., `daemon:`, `actor:`)
- 50 character summary, 72 character body wrapping
- Present tense commit messages

## Dependencies

Key external dependencies:
- `github.com/lightningnetwork/lnd/actor`: Actor system framework
- `github.com/google/go-dap`: Debug Adapter Protocol implementation
- `github.com/charmbracelet/bubbletea`: Terminal UI framework
- `github.com/mark3labs/mcp-go`: MCP protocol support

## Testing Patterns

- Use `testify` for assertions
- Actor tests should use test actor systems
- Mock external dependencies (Delve, DAP clients)
- Integration tests for full debug workflows

## Common Workflows (Updated for Package Structure)

### Adding New Debug Commands
1. Define message types in `debugger/messages.go`
2. Add handling in `debugger/session.go` actor's `Receive` method
3. Update DAP protocol mapping in `debugger/dap_*.go`
4. Add MCP tool in `mcp/mcp_server.go` if needed
5. Add TUI command in `tui/tui.go` if needed

### Session Management
- Sessions created by debugger factory actor in `debugger/` package
- Each session gets unique service key managed by `MCPDebugService`
- Sessions handle DAP requests/responses via `debugger.DAPRequest`/`debugger.DAPResponse`
- Automatic cleanup on session end through service lifecycle

### Package Development Workflow
1. **Identify the right package**: Use package responsibilities to locate code
2. **Check imports**: Ensure proper package prefixes (`debugger.`, `mcp.`, `tui.`)
3. **Use service API**: Always use `MCPDebugService` for lifecycle management
4. **Test package isolation**: Ensure each package can be tested independently
5. **Follow LND guidelines**: Maintain commit structure with package prefixes

## MCP Tools

The MCP server exposes 15 standardized debugging tools:

### Session Management
- `create_debug_session` - Create a new debugging session
- `initialize_session` - Initialize DAP protocol with client capabilities

### Program Control
- `launch_program` - Launch a Go program for debugging with full configuration
- `attach_to_process` - Attach to an existing running process for debugging
- `configuration_done` - Signal that configuration is complete

### Breakpoint Management
- `set_breakpoints` - Set source code breakpoints by file and line numbers

### Execution Control
- `continue_execution` - Continue program execution
- `step_next` - Step over (execute next line without entering functions)
- `step_in` - Step into function calls
- `step_out` - Step out of current function
- `pause_execution` - Pause program execution

### Inspection Tools
- `get_threads` - Get information about all threads
- `get_stack_frames` - Get call stack for a specific thread
- `get_variables` - Get variable values for a specific scope
- `evaluate_expression` - Evaluate expressions in context

### Type-Safe Arguments
All tools use strongly-typed argument structures:
- `CreateSessionArgs`, `LaunchProgramArgs`, `AttachToProcessArgs`, `SetBreakpointsArgs`, etc.
- JSON schema validation for all parameters
- Comprehensive error handling and reporting


## Current State

The project is production-ready with:
- Clean package architecture with single responsibilities
- Full Bubble Tea TUI with real-time metrics
- 15 MCP debugging tools exposed via server
- Actor router pattern for load balancing
- Process attachment support for debugging running processes


### Key Files for Development

**Service Layer:**
- `daemon.go`: Main service API, start here for lifecycle management
- `cmd/tui/main.go`: TUI application entry point
- `cmd/dlv-mcp-server/main.go`: MCP server entry point

**Core Packages:**
- `debugger/`: All DAP protocol and debugging functionality
- `mcp/mcp_server.go`: MCP server with 15 debugging tools
- `tui/tui.go`: Bubble Tea interface with real-time monitoring

**Documentation:**
- `agent_planning/TUI_IMPLEMENTATION_SUMMARY.md`: Complete TUI implementation details
- `agent_planning/CLEANUP_SUMMARY.md`: Project structure cleanup summary
- `agent_planning/METRICS_FIXES.md`: Real metrics implementation
- `agent_planning/PACKAGE_RESTRUCTURING.md`: Package architecture details
- `agent_planning/ACTOR.md`: Actor system patterns and usage
- `agent_planning/TUI_DESIGN.md`: TUI architecture and components
- `agent_planning/ATTACH_PROCESS_PLANNING.md`: Process attachment implementation

### Development Tips
- **Package Focus**: Each package has single responsibility, use that to guide changes
- **Service First**: Always start with `MCPDebugService` for proper initialization
- **Type Safety**: Use package prefixes (`debugger.`, `mcp.`, `tui.`) consistently
- **LND Guidelines**: Follow commit conventions with package prefixes
- **Test Isolation**: Each package should be testable independently

### Debugging Common Issues
- **Import errors**: Check package declaration and import path structure
- **Type not found**: Verify type is exported (capitalized) from correct package
- **Service issues**: Ensure `MCPDebugService` lifecycle is properly managed
- **Actor communication**: Use typed actor references with correct message types

The project is production-ready with clean architecture and comprehensive debugging capabilities.