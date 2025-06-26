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

## Package Architecture

The project is organized into focused packages that compose together cleanly:

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

### Service-Oriented Design

```go
// Clean lifecycle management
service := mcpdebug.NewMCPDebugService()
defer service.Stop()

// Get components as needed
mcpServer := service.GetMCPServer() 
actorSystem := service.GetActorSystem()

// Convenience functions for simple usage
mcpdebug.RunTUI()                    // Interactive TUI
mcpServer, service := mcpdebug.NewMCPServer()  // Headless server
```

### Architecture Flow

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   TUI Package   │    │   MCP Package    │    │ Debugger Package│
│                 │    │                  │    │                 │
│ • Dashboard     │◄──►│ • JSON-RPC 2.0   │◄──►│ • DAP Protocol  │
│ • Sessions      │    │ • 14 Tools       │    │ • Delve Backend │
│ • Commands      │    │ • Actor Router   │    │ • Type Wrappers │
│ • Logs          │    │ • Type Safety    │    │ • Session Mgmt  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         ▲                       ▲                       ▲
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌──────────────────┐
                    │   Root Package   │
                    │                  │
                    │ • MCPDebugService│
                    │ • Lifecycle Mgmt │
                    │ • Clean API      │
                    │ • Composition    │
                    └──────────────────┘
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

### Package Details

#### 📚 `claude/` - Documentation Archive
Contains all incremental implementation notes and milestone documentation:
- `ACTOR.md` - Actor system patterns and usage
- `TUI_DESIGN.md` - TUI architecture decisions  
- `CLEANUP_SUMMARY.md` - Project restructuring notes
- Implementation summaries and development history

#### 🔧 `debugger/` - DAP Integration
Core debugging functionality with actor-based message handling:
- `dap_*.go` - Debug Adapter Protocol implementations
- `debug_types.go` - Type-safe wrappers for debugging operations
- `session.go` - Debug session lifecycle management
- `messages.go` - Actor message definitions

#### 🌐 `mcp/` - AI Tool Server  
Model Context Protocol server exposing debugging as standardized tools:
- `mcp_server.go` - 14 debugging tools for AI clients
- JSON-RPC 2.0 over stdio transport
- Type-safe argument structures for all operations

#### 🖥️ `tui/` - Interactive Interface
Bubble Tea terminal user interface:
- `tui.go` - Complete dashboard with real-time metrics
- Multi-tab navigation (Dashboard, Sessions, Clients, Commands, Logs)
- Interactive command execution with history

#### 📦 `cmd/` - Applications
Production command-line programs:
```
cmd/
├── mcp-server/     # Headless MCP server for AI integration
└── tui/           # Interactive TUI console for monitoring
```

#### 🧪 `internal/test/` - Development Tools
Validation and testing utilities:
```
internal/test/
└── tui-validation/  # TUI component validation without TTY
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

- [`claude/TUI_USAGE.md`](claude/TUI_USAGE.md) - Complete TUI user guide
- [`claude/TUI_DESIGN.md`](claude/TUI_DESIGN.md) - TUI architecture and design patterns  
- [`claude/ACTOR.md`](claude/ACTOR.md) - Actor system usage patterns
- [`cmd/README.md`](cmd/README.md) - Binary descriptions and usage
- [`claude/CLEANUP_SUMMARY.md`](claude/CLEANUP_SUMMARY.md) - Package restructuring summary

## Key Features

✅ **Clean Package Architecture** - Focused packages with clear separation of concerns  
✅ **Service-Oriented Design** - Proper lifecycle management with `MCPDebugService`  
✅ **Type-Safe Composition** - All packages properly import and compose together  
✅ **Production Ready** - Comprehensive error handling and clean APIs  
✅ **Actor-Based Concurrency** - LND's proven actor system with router patterns  
✅ **Modern TUI Framework** - Bubble Tea with official components  
✅ **AI Integration Ready** - MCP server exposes debugging as standardized tools  
✅ **Real-Time Monitoring** - Live metrics, session tracking, and log streaming

## Development Philosophy

This project demonstrates best practices for building **composable, service-oriented Go applications**:

- **Package Boundaries**: Each package has a single responsibility and clean interfaces
- **Lifecycle Management**: Proper service initialization, cleanup, and resource management  
- **Type Safety**: Strong typing throughout with exported interfaces between packages
- **Actor Patterns**: Proven concurrency patterns from Lightning Network development
- **API Design**: Clean, discoverable APIs that hide complexity while enabling power users

Perfect for learning modern Go architecture patterns, actor-based concurrency, and building AI-integrated development tools.