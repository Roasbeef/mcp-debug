# MCP Debug Server TUI Usage Guide

## Overview

The MCP Debug Server TUI (Terminal User Interface) provides a comprehensive monitoring and management console for the debugging server. Built with **Bubble Tea** using proper component composition and the **LND Actor Router** pattern, it offers real-time monitoring of debugging sessions, client connections, and interactive command execution.

## Key Improvements

✅ **Proper Bubble Tea Architecture**: Uses official Bubble Tea components (tables, textinput, viewport, help)  
✅ **Actor Router Pattern**: Leverages LND's actor router for simplified actor selection and load balancing  
✅ **Component Composition**: Follows Bubble Tea best practices with proper model-view-update pattern  
✅ **Type-Safe Key Bindings**: Uses Bubble Tea's key binding system with help integration  
✅ **Responsive Design**: Automatically adapts to terminal window size changes

## Starting the TUI

### Build and Run
```bash
# Build the TUI console
go build -o tui-console ./cmd/tui

# Run the TUI
./tui-console
```

### Using the Test Version
```bash
# For environments without TTY support
go build -o tui-test ./cmd/tui-test
./tui-test
```

## Interface Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ MCP Debug Server Console                    v1.0.0    [Running] │
├─────────────────────────────────────────────────────────────────┤
│ Status: ● Running  │ Clients: 1  │ Sessions: 1  │ Requests: 15  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ [Dashboard] [Sessions] [Clients] [Commands] [Logs] [Help]       │
│                                                                 │
│ Current View Content Area                                       │
│                                                                 │
│ Navigation: Tab/Shift+Tab, ↑↓ Select, Enter Activate, Ctrl+C Quit │
└─────────────────────────────────────────────────────────────────┘
```

## Navigation

### Keyboard Shortcuts
- **Tab/Shift+Tab**: Navigate between main tabs
- **↑/↓**: Navigate within lists and menus
- **Enter**: Select/activate item or execute command
- **Ctrl+C** or **Q**: Quit application
- **Ctrl+R**: Refresh current view
- **/** or **Ctrl+F**: Open search/filter (when available)

### Available Views

## 1. Dashboard View
Shows server overview and real-time statistics:
- Server status and uptime
- Connected client count
- Active debugging sessions
- Request/response statistics
- Error rates and performance metrics
- Quick action guidance

## 2. Sessions View
Monitor active debugging sessions:
- Session ID and creation time
- Program path and process information
- Connected client details
- Session status (running, paused, terminated)
- Breakpoint information
- Current execution context

**Demo Session Example:**
```
● demo_session_1  │ ./examples/simple/simple  │ claude-code    │ Running
  Created: 5m ago │ Breakpoints: 2           │ Last: just now
```

## 3. Clients View
Track connected MCP clients:
- Client identifier and connection time
- Protocol version and capabilities
- Request/response statistics per client
- Last activity timestamps
- Connection health status

**Demo Client Example:**
```
● claude-code     │ Connected: 10m ago │ Requests: 15
  Version: 1.0.0  │ Errors: 0         │ Last: just now
```

## 4. Commands View
Interactive command execution interface:

### Available MCP Tools
```bash
# Session Management
create_debug_session {"session_id": "debug1"}
initialize_session {"session_id": "debug1", "client_id": "client1"}

# Program Control
launch_program {"session_id": "debug1", "program": "./path/to/program"}
configuration_done {"session_id": "debug1"}

# Breakpoint Management
set_breakpoints {"session_id": "debug1", "file": "./main.go", "lines": [15, 25]}

# Execution Control
continue_execution {"session_id": "debug1", "thread_id": 1}
step_next {"session_id": "debug1", "thread_id": 1}
step_in {"session_id": "debug1", "thread_id": 1}
step_out {"session_id": "debug1", "thread_id": 1}
pause_execution {"session_id": "debug1", "thread_id": 1}

# Inspection
get_threads {"session_id": "debug1"}
get_stack_frames {"session_id": "debug1", "thread_id": 1}
get_variables {"session_id": "debug1", "frame_id": 1}
evaluate_expression {"session_id": "debug1", "expression": "myVar", "frame_id": 1}
```

### Command Execution
1. Type commands in the input field
2. Press **Enter** to execute
3. View responses in the result area
4. Browse command history with ↑↓ keys

### Built-in Commands
- **help** or **?**: Show available commands
- Commands are executed as JSON-formatted MCP tool calls

## 5. Logs View
Real-time log monitoring:
- System and session logs
- Filterable by level (Info, Warning, Error)
- Auto-scroll to latest entries
- Searchable log history
- Structured log display

**Log Entry Format:**
```
[INFO] 14:32:15 MCP Server: Client connected: claude-code
[INFO] 14:32:20 Debug Session: Session created: demo_session_1
[INFO] 14:32:25 Breakpoints: Breakpoint set at main.go:15
```

## 6. Help View
Documentation and keyboard shortcuts reference.

## Real-time Features

### Automatic Updates
- Server metrics refresh every 5 seconds
- Session status updates in real-time
- Client activity monitoring
- Live log streaming with auto-scroll

### Interactive Elements
- **Command History**: Navigate previous commands with ↑↓
- **List Navigation**: Use arrow keys to browse sessions/clients/logs
- **Auto-refresh**: Press Ctrl+R to manually refresh current view
- **Status Indicators**: Color-coded status displays

## Example Workflow

### 1. Start a Debugging Session
```bash
# In the Commands tab, execute:
create_debug_session {"session_id": "my_debug"}
initialize_session {"session_id": "my_debug", "client_id": "my_client"}
launch_program {"session_id": "my_debug", "program": "./examples/simple/simple"}
configuration_done {"session_id": "my_debug"}
```

### 2. Set Breakpoints
```bash
set_breakpoints {"session_id": "my_debug", "file": "./examples/simple/main.go", "lines": [15]}
```

### 3. Control Execution
```bash
continue_execution {"session_id": "my_debug", "thread_id": 1}
get_threads {"session_id": "my_debug"}
get_variables {"session_id": "my_debug", "frame_id": 1}
```

### 4. Monitor Progress
- Switch to **Sessions** tab to see session status
- Check **Logs** tab for execution details
- Use **Dashboard** for overall system health

## Troubleshooting

### TTY Issues
If you see "could not open a new TTY" error:
- Use the test version: `./tui-test`
- Run in a proper terminal environment
- Check terminal compatibility

### Actor System Issues
- Ensure the daemon is properly initialized
- Check that debugger actors are registered
- Verify actor system shutdown on exit

### Performance
- Large log files may affect performance
- Use log filtering to focus on relevant entries
- Refresh views manually if auto-update is slow

## Architecture Notes

### Components
- **Bubble Tea Framework**: Official terminal UI framework with proper component composition
- **Actor Router System**: LND's actor router for distributed message handling and load balancing
- **MCP Server Integration**: Model Context Protocol server with type-safe tool execution
- **Responsive Components**: Tables, viewports, and inputs that adapt to terminal size

### Design Patterns
- **Model-View-Update**: Bubble Tea's recommended architecture pattern
- **Router-based Actor Communication**: Uses LND's router abstraction for simplified actor selection
- **Component Composition**: Proper use of Bubble Tea's built-in components instead of custom implementations
- **Type-Safe Key Bindings**: Leverage Bubble Tea's key binding system with integrated help
- **Message-Driven Updates**: Event-driven state updates with proper command batching

### Technical Improvements
- **Actor Router Pattern**: Replaces manual actor selection with round-robin routing
- **Native Bubble Tea Components**: Uses `table.Model`, `textinput.Model`, `viewport.Model`, `help.Model`
- **Proper Update Handling**: Follows Bubble Tea patterns for component updates and command batching
- **Responsive Layout**: Components automatically resize based on terminal window changes
- **Clean Separation**: Model state management separate from view rendering logic

The improved TUI provides a comprehensive interface for monitoring and controlling the MCP Debug Server, leveraging both Bubble Tea and LND actor system best practices for a robust and maintainable terminal application.