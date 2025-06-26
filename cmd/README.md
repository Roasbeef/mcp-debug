# MCP Debug Server Commands

This directory contains the main executable programs for the MCP Debug Server.

## Available Programs

### 🖥️ TUI Console (Recommended)
```bash
go build -o tui-console ./cmd/tui
./tui-console
```

**Primary interface** - Interactive Terminal User Interface with:
- Real-time server monitoring dashboard
- Debugging session management
- MCP client connection tracking  
- Interactive command execution
- Live log viewing with filtering
- Built-in help system

**Features:**
- Tab navigation (Dashboard, Sessions, Clients, Commands, Logs)
- Keyboard shortcuts (Tab, Arrow keys, Ctrl+C to quit)
- Real-time metrics (uptime, request count, error tracking)
- Interactive MCP tool execution
- Proper Bubble Tea components with actor router integration

### 🔧 Standalone MCP Server
```bash
go build -o mcp-server ./cmd/mcp-server  
./mcp-server
```

**Headless server** - Pure MCP server for integration with external clients:
- JSON-RPC 2.0 over stdio transport
- Complete DAP (Debug Adapter Protocol) support
- 14 debugging tools (create_debug_session, launch_program, set_breakpoints, etc.)
- Actor-based architecture with LND patterns
- Suitable for integration with AI/LLM clients

## Recommended Usage

**For interactive debugging and monitoring:**
```bash
./tui-console
```

**For programmatic/API access:**
```bash
./mcp-server
```

## Internal Tools

Development and validation tools are located in:
- `internal/test/tui-validation/` - TUI component testing utility

## Architecture

Both programs use:
- **LND Actor System** with router pattern for scalable message handling
- **DAP Protocol** for standardized debugging operations  
- **MCP Protocol** for AI-friendly tool integration
- **Type-safe message interfaces** throughout

The TUI provides the complete user experience, while the standalone server enables headless integration scenarios.