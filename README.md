# MCP Debug Server

A comprehensive debugging server that bridges the Debug Adapter Protocol (DAP) with the Model Context Protocol (MCP), enabling AI-powered debugging workflows for Go applications.

## Quick Start

### 🖥️ Interactive TUI Console (Recommended)
```bash
go build -o tui-console ./cmd/tui
./tui-console
```

### 🔧 Headless MCP Server
```bash
go build -o mcp-server ./cmd/mcp-server
./mcp-server
```

## Features

### 🎯 **TUI Console** 
Interactive terminal interface with real-time monitoring:
- **Dashboard**: Server metrics, uptime, request/error tracking
- **Sessions**: Active debugging session management with tables
- **Clients**: MCP client connection monitoring  
- **Commands**: Interactive MCP tool execution with history
- **Logs**: Real-time log streaming with auto-scroll
- **Help**: Integrated keyboard shortcuts and documentation

### 🔌 **MCP Server Integration**
14 debugging tools accessible via Model Context Protocol:
- `create_debug_session` - Initialize new debugging sessions
- `launch_program` - Start Go programs for debugging
- `set_breakpoints` - Manage source code breakpoints
- `continue_execution`, `step_next`, `step_in`, `step_out` - Execution control
- `get_threads`, `get_stack_frames`, `get_variables` - Program inspection
- `evaluate_expression` - Runtime expression evaluation

### ⚡ **Actor-Based Architecture**
- **LND Actor System** with router pattern for scalable concurrency
- **Round-robin load balancing** for multiple debugger instances
- **Type-safe message interfaces** with comprehensive error handling
- **Clean separation** between protocol layers and business logic

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   TUI Console   │    │   MCP Server     │    │  Debug Targets  │
│                 │    │                  │    │                 │
│ • Dashboard     │◄──►│ • JSON-RPC 2.0   │◄──►│ • Go Programs   │
│ • Sessions      │    │ • 14 Tools       │    │ • Delve Backend │
│ • Commands      │    │ • Actor Router   │    │ • DAP Protocol  │
│ • Logs          │    │ • Type Safety    │    │ • Breakpoints   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Technology Stack

- **Backend**: Go with Lightning Network's actor system
- **Protocols**: DAP (Debug Adapter Protocol) + MCP (Model Context Protocol)  
- **TUI**: Bubble Tea with native components (tables, viewports, textinput)
- **Debugging**: Delve debugger integration
- **Architecture**: Actor model with router pattern for concurrency

## Development

### Build All
```bash
go mod tidy
go build -o tui-console ./cmd/tui
go build -o mcp-server ./cmd/mcp-server
```

### Test Components
```bash
# Test TUI components
go build -o tui-test ./internal/test/tui-validation
./tui-test

# Run test suite
go test -v ./...
```

### Project Structure
```
cmd/
├── mcp-server/     # Headless MCP server
└── tui/           # Interactive TUI console

internal/
└── test/          # Development and validation tools

*.go               # Core implementation
*_test.go         # Test suites
*.md              # Documentation
```

## Usage Examples

### TUI Console Workflow
1. Start TUI: `./tui-console`
2. Navigate with Tab key between views
3. Use Commands tab to execute: `create_debug_session {"session_id": "debug1"}`
4. Monitor sessions in Sessions tab
5. View logs in Logs tab
6. Press `?` for help

### MCP Server Integration
```bash
# Start server
./mcp-server

# MCP client can now send JSON-RPC requests:
{
  "jsonrpc": "2.0",
  "method": "tools/call", 
  "params": {
    "name": "create_debug_session",
    "arguments": {"session_id": "debug1"}
  }
}
```

## Documentation

- [`TUI_USAGE.md`](TUI_USAGE.md) - Complete TUI user guide
- [`TUI_DESIGN.md`](TUI_DESIGN.md) - TUI architecture and design patterns  
- [`ACTOR.md`](ACTOR.md) - Actor system usage patterns
- [`cmd/README.md`](cmd/README.md) - Binary descriptions and usage

## Key Improvements

✅ **Proper Bubble Tea Architecture** - Uses official components with composition  
✅ **Actor Router Pattern** - LND's router abstraction for simplified actor selection  
✅ **Real Metrics** - Live uptime, request counts, error tracking  
✅ **Type Safety** - Strongly typed throughout with proper constants  
✅ **Production Ready** - Comprehensive error handling and clean architecture

This project demonstrates best practices for building robust, concurrent Go applications with modern TUI frameworks and actor-based architectures.