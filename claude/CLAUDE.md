# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an MCP (Model Context Protocol) debugger server that exposes debugging capabilities for Go applications as standardized MCP tools. The server integrates the Debug Adapter Protocol (DAP) with Delve debugger through an actor-based architecture using the Lightning Network's actor system. LLM clients can use this server to perform AI-powered debugging workflows.

**🎯 Current Status**: Fully restructured into clean package architecture with service-oriented design (completed in commits 8850603-f11b709).

## Package Architecture

The project is organized into focused packages with clean separation of concerns:

```
mcp-debug/
├── claude/           📚 Documentation & implementation notes  
├── debugger/         🔧 DAP/Delve integration with actor system
├── mcp/             🌐 MCP server exposing debugging as AI tools  
├── tui/             🖥️ Bubble Tea TUI for interactive debugging
├── cmd/             📦 Production command-line applications
├── internal/test/   🧪 Development & validation utilities
└── daemon.go        🏗️ Clean API with lifecycle management
```

### Package Responsibilities

- **claude/**: Documentation archive (implementation notes, design decisions)
- **debugger/**: DAP protocol, Delve integration, actor message handling
- **mcp/**: MCP server with 14 debugging tools for AI clients
- **tui/**: Bubble Tea terminal interface with real-time monitoring
- **cmd/**: Production applications (tui-console, mcp-server)
- **internal/test/**: Development validation tools

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

### Building (Updated for Package Structure)
```bash
# Interactive TUI console (recommended for monitoring)
go build -o tui-console ./cmd/tui
./tui-console

# Headless MCP server (for AI integration)
go build -o mcp-server ./cmd/mcp-server
./mcp-server

# Development validation tool
go build -o tui-validation ./internal/test/tui-validation
./tui-validation

# Build all applications
go mod tidy
make build-all  # if Makefile exists
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

The project uses LND's actor system with these patterns, now organized within the **debugger/** package:

### Package-Specific Actor Components

**debugger/** package contains:
- `messages.go`: `DebuggerCmd`, `DebuggerResp`, actor message types
- `debugger.go`: Main debugger actor implementation
- `session.go`: Session-specific actors for debug instances
- `dap_messages.go`: `DAPRequest`, `DAPResponse` for protocol handling

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
- **tui/**: Uses `mcp.MCPDebugServer.GetSessions()` for monitoring
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

The MCP server exposes the following standardized debugging tools:

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

### Usage Examples

#### Create Session and Launch Program
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "create_debug_session",
    "arguments": {
      "session_id": "debug1"
    }
  }
}
```

#### Attach to Existing Process
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "attach_to_process",
    "arguments": {
      "session_id": "debug1",
      "process_id": 12345,
      "name": "my-go-service"
    }
  }
}

## 🎯 Current State & Next Steps

### Recently Completed (Commits 8850603-f11b709)
✅ **Package Restructuring**: Complete transition to focused package architecture  
✅ **Service-Oriented API**: Clean lifecycle management with `MCPDebugService`  
✅ **Type-Safe Imports**: All packages properly reference each other  
✅ **Build Verification**: All applications compile and run successfully  
✅ **Documentation Update**: README and guides reflect new structure  

### Project Status
- **✅ Fully Functional**: All builds working, TUI and MCP server operational
- **✅ Clean Architecture**: Focused packages with single responsibilities
- **✅ Production Ready**: Proper service lifecycle and error handling
- **✅ Well Documented**: Comprehensive documentation in claude/ package

### Recommended Next Steps
1. **Enhanced Testing**: Add integration tests for each package
2. **Error Handling**: Improve error types and propagation across packages
3. **Configuration Management**: Add config files for deployment scenarios
4. **Performance Optimization**: Profile and optimize actor message passing
5. **Feature Extensions**: Add new debugging capabilities or AI integration features

### Key Files for Development

**Service Layer:**
- `daemon.go`: Main service API, start here for lifecycle management
- `cmd/tui/main.go`: TUI application entry point
- `cmd/mcp-server/main.go`: MCP server entry point

**Core Packages:**
- `debugger/`: All DAP protocol and debugging functionality
- `mcp/mcp_server.go`: MCP server with 14 debugging tools
- `tui/tui.go`: Bubble Tea interface implementation

**Documentation:**
- `claude/PACKAGE_RESTRUCTURING.md`: Detailed restructuring summary
- `claude/ACTOR.md`: Actor system patterns and usage
- `claude/TUI_DESIGN.md`: TUI architecture and components

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

The project is now in excellent shape with clean architecture, proper separation of concerns, and all builds working. Ready for feature development or production deployment! 🚀
```