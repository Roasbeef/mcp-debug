# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an MCP (Model Context Protocol) debugger server that exposes debugging capabilities for Go applications as standardized MCP tools. The server integrates the Debug Adapter Protocol (DAP) with Delve debugger through an actor-based architecture using the Lightning Network's actor system. LLM clients can use this server to perform AI-powered debugging workflows.

## Architecture

The system uses an actor-based architecture with the following key components:

- **MCP Server**: Exposes debugging functionality through MCP protocol
- **DAP Adapter**: Communicates with debuggers using Debug Adapter Protocol  
- **Actor System**: Manages concurrent debugging sessions and components
- **Delve Integration**: Uses Go's Delve debugger as the backend
- **TUI Interface**: Terminal-based user interface using Bubble Tea

### Core Components

- `daemon.go`: Main actor system initialization and debugger factory
- `debugger.go`: Actor responsible for creating and managing debug sessions
- `session.go`: Individual debugging session management
- `dap*.go`: Debug Adapter Protocol implementation
- `cmd/mcp-debugger/main.go`: TUI application entry point
- `actor/`: Actor system implementation (from LND)

## Development Commands

### Building
```bash
go build -o mcp-debugger ./cmd/mcp-debugger
```

### Running
```bash
./mcp-debugger
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./actor/...

# Run with verbose output
go test -v ./...
```

### MCP Server
```bash
# Build and run the MCP server
go build -o mcp-server ./cmd/mcp-server
./mcp-server

# Test MCP server functionality
./test_mcp_server.sh
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

The project uses LND's actor system with these patterns:

### Message Types
- All messages implement `actor.Message` interface
- Embed `actor.BaseMessage` for convenience
- Define `MessageType()` method for debugging

### Service Keys
- Type-safe identifiers for actor registration/discovery
- Format: `actor.NewServiceKey[RequestType, ResponseType](name)`
- Used for actor lookup in receptionist

### Actor Lifecycle
- Actors created via `actor.NewActor()` with `ActorConfig`
- Registered with system using `actor.RegisterWithSystem()`
- Managed by global `System` variable in `daemon.go`

### Communication Patterns
- **Tell**: Fire-and-forget messages via `ActorRef.Tell()`
- **Ask**: Request-response via `ActorRef.Ask()` returning `Future[R]`

## Key Service Keys

- `DebuggerKey`: Main debugger factory actor
- Session-specific keys: Created dynamically as `"session-{id}"`

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

## Common Workflows

### Adding New Debug Commands
1. Define message types in `messages.go`
2. Add handling in session actor's `Receive` method
3. Update DAP protocol mapping
4. Add TUI command if needed

### Session Management
- Sessions created by debugger factory actor
- Each session gets unique service key
- Sessions handle DAP requests/responses
- Automatic cleanup on session end

## MCP Tools

The MCP server exposes the following standardized debugging tools:

### Session Management
- `create_debug_session` - Create a new debugging session
- `initialize_session` - Initialize DAP protocol with client capabilities

### Program Control
- `launch_program` - Launch a Go program for debugging with full configuration
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
- `CreateSessionArgs`, `LaunchProgramArgs`, `SetBreakpointsArgs`, etc.
- JSON schema validation for all parameters
- Comprehensive error handling and reporting

### Usage Example
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