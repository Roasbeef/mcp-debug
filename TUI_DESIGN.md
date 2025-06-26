# TUI Design Document

## Overview

The TUI (Terminal User Interface) for the MCP Debug Server provides a comprehensive monitoring and management console built with Bubble Tea. This interface serves as the primary way users interact with the server to monitor connections, manage debugging sessions, and execute interactive commands.

## Core Features

### 1. Server Status Dashboard
- **Real-time Status**: Display MCP server running state (active/stopped/error)
- **Connection Metrics**: Number of connected MCP clients with timestamps
- **Session Overview**: Active debugging sessions with session IDs and metadata
- **Performance Stats**: Total requests served, request rate, response times
- **Resource Usage**: Memory usage, goroutine count, connection pool status

### 2. Session Management View
- **Active Sessions List**: All current debugging sessions with details
  - Session ID and creation time
  - Program being debugged (path, process ID)
  - Connected client identifier
  - Session status (initializing, running, paused, terminated)
  - Current breakpoints and execution state
- **Session Controls**: Ability to terminate or inspect sessions
- **Session Activity**: Real-time updates on debugging operations
- **Session History**: Recently completed sessions

### 3. Client Connection Monitor
- **Connected Clients**: List of active MCP client connections
  - Client identifier and connection time
  - Client capabilities and protocol version
  - Last activity timestamp
  - Request/response statistics per client
- **Connection Health**: Monitor connection status and errors
- **Client Activity Log**: Recent requests and responses per client

### 4. Interactive Command Interface
- **Command Input**: Execute MCP tools directly from the TUI
- **Auto-completion**: Tab completion for MCP tool names and parameters
- **Command History**: Browse and re-execute previous commands
- **Response Display**: Real-time display of command results
- **JSON Formatting**: Pretty-printed JSON responses
- **Error Handling**: Clear display of command errors and validation

### 5. Log Viewer
- **Multi-level Logging**: Debug, Info, Warning, Error log levels
- **Log Filtering**: Filter by level, session, client, or keyword
- **Real-time Updates**: Live log streaming with auto-scroll
- **Log Export**: Save logs to file for analysis
- **Search Functionality**: Search through log history
- **Structured Display**: Formatted display of structured log entries

### 6. Statistics and Analytics
- **Performance Metrics**: Request latency, throughput, error rates
- **Usage Analytics**: Most used tools, session duration statistics
- **Trend Analysis**: Historical data visualization in ASCII charts
- **Health Monitoring**: System health indicators and alerts

## TUI Layout Structure

```
┌─────────────────────────────────────────────────────────────────┐
│ MCP Debug Server Console                    v1.0.0    [Running] │
├─────────────────────────────────────────────────────────────────┤
│ Status: ● Active  │ Clients: 2  │ Sessions: 3  │ Requests: 847  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ [Dashboard] [Sessions] [Clients] [Commands] [Logs] [Help]       │
│                                                                 │
│ Current View Content Area:                                      │
│ ┌─ Active Sessions ─────────────────────────────────────────┐   │
│ │ ● debug1  │ ./examples/simple     │ vscode    │ Running   │   │
│ │ ● debug2  │ ./myapp/main.go       │ cursor    │ Paused    │   │
│ │ ● debug3  │ ./server/cmd/main.go  │ claude    │ Stepping  │   │
│ │                                                           │   │
│ │ Selected: debug1 - Line 15, main.go                      │   │
│ │ Breakpoints: main.go:15, utils.go:42                     │   │
│ └───────────────────────────────────────────────────────────┘   │
│                                                                 │
│ ┌─ Command Input ───────────────────────────────────────────┐   │
│ │ > get_threads {"session_id": "debug1"}                   │   │
│ │                                          [Enter to execute] │
│ └───────────────────────────────────────────────────────────┘   │
│                                                                 │
│ ┌─ Activity Feed ───────────────────────────────────────────┐   │
│ │ 14:32:15 [debug1] Client connected: claude-code          │   │
│ │ 14:32:20 [debug1] Session initialized successfully       │   │
│ │ 14:32:25 [debug1] Breakpoint set: main.go:15             │   │
│ │ 14:32:30 [debug1] Execution paused at breakpoint         │   │
│ └───────────────────────────────────────────────────────────┘   │
│                                                                 │
│ Navigation: Tab/Shift+Tab, ↑↓ Select, Enter Activate, Ctrl+C Quit │
└─────────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. Main Application Model
```go
type Model struct {
    // Core state
    serverStatus   ServerStatus
    currentView    ViewType
    windowSize     tea.WindowSizeMsg
    
    // Sub-models
    dashboard      DashboardModel
    sessions       SessionsModel
    clients        ClientsModel
    commands       CommandsModel
    logs           LogsModel
    
    // Shared state
    mcpServer      *MCPDebugServer
    actorSystem    *actor.ActorSystem
    
    // UI state
    tabs           []string
    activeTab      int
    statusLine     StatusLineModel
}
```

### 2. Dashboard Model
```go
type DashboardModel struct {
    serverMetrics  ServerMetrics
    realtimeStats  RealtimeStats
    systemHealth   HealthIndicators
    chartData      []ChartPoint
}
```

### 3. Sessions Model
```go
type SessionsModel struct {
    activeSessions []SessionInfo
    selectedIndex  int
    detailView     bool
    sessionLogs    map[string][]LogEntry
}

type SessionInfo struct {
    ID             string
    ClientID       string
    ProgramPath    string
    Status         SessionStatus
    CreatedAt      time.Time
    LastActivity   time.Time
    Breakpoints    []BreakpointInfo
    CurrentFrame   *StackFrameInfo
}
```

### 4. Clients Model
```go
type ClientsModel struct {
    connectedClients []ClientInfo
    selectedIndex    int
    connectionStats  map[string]ClientStats
}

type ClientInfo struct {
    ID               string
    ConnectedAt      time.Time
    LastActivity     time.Time
    ProtocolVersion  string
    Capabilities     []string
    RequestCount     int
    ErrorCount       int
}
```

### 5. Commands Model
```go
type CommandsModel struct {
    input          textinput.Model
    history        []string
    historyIndex   int
    suggestions    []string
    response       string
    isExecuting    bool
    availableTools []ToolInfo
}
```

### 6. Logs Model
```go
type LogsModel struct {
    entries       []LogEntry
    filtered      []LogEntry
    selectedIndex int
    autoScroll    bool
    filters       LogFilters
    searchQuery   string
}

type LogEntry struct {
    Timestamp time.Time
    Level     LogLevel
    Component string
    SessionID string
    ClientID  string
    Message   string
    Data      map[string]interface{}
}
```

## Navigation and Interaction

### Key Bindings
- **Tab/Shift+Tab**: Navigate between main tabs
- **↑/↓**: Navigate within lists and menus
- **←/→**: Navigate between sub-views or panels
- **Enter**: Select/activate item or execute command
- **Esc**: Return to previous view or cancel operation
- **Space**: Toggle selection or pause/resume
- **Ctrl+C**: Quit application
- **Ctrl+R**: Refresh current view
- **Ctrl+L**: Clear current view/logs
- **/** or **Ctrl+F**: Open search/filter
- **?** or **F1**: Show help

### Tab Navigation
1. **Dashboard** - Overview and system status
2. **Sessions** - Active debugging sessions management
3. **Clients** - Connected MCP clients monitoring
4. **Commands** - Interactive command execution
5. **Logs** - System and session logs viewer
6. **Help** - Help and documentation

### View States
Each view can have multiple states:
- **List View**: Overview with navigation
- **Detail View**: Detailed information about selected item
- **Edit View**: Interactive editing or command input
- **Filter View**: Search and filtering interface

## Real-time Updates

### Update Sources
1. **MCP Server Events**: Connection, disconnection, errors
2. **Session Events**: Creation, state changes, termination
3. **Debug Events**: Breakpoint hits, step operations, variable changes
4. **System Events**: Performance metrics, resource usage
5. **Log Events**: New log entries from all components

### Update Mechanism
- Use Bubble Tea's `tea.Cmd` for async updates
- WebSocket or event channel integration for real-time data
- Periodic polling for system metrics
- Event-driven updates for debugging operations

## Error Handling and User Feedback

### Error Display
- **Status Line**: Brief error messages and status updates
- **Modal Dialogs**: Detailed error information with actions
- **Inline Messages**: Contextual errors within views
- **Log Integration**: Automatic error logging with context

### User Feedback
- **Loading Indicators**: Spinners for long operations
- **Progress Bars**: For operations with known duration
- **Status Messages**: Success/failure confirmations
- **Color Coding**: Status indicators using colors
- **Sound Alerts**: Optional audio feedback for critical events

## Performance Considerations

### Efficient Updates
- **Selective Rendering**: Only update changed components
- **Pagination**: Handle large lists with pagination
- **Lazy Loading**: Load details on demand
- **Debouncing**: Limit update frequency for high-volume events

### Memory Management
- **Log Rotation**: Limit in-memory log entries
- **Session Cleanup**: Remove terminated session data
- **Cache Management**: Efficient caching of frequently accessed data

## Accessibility

### Keyboard Navigation
- Full keyboard navigation without mouse dependency
- Consistent key bindings across all views
- Tab order that follows logical flow
- Escape sequences for all interactive elements

### Visual Design
- High contrast color schemes
- Clear visual hierarchy with borders and spacing
- Status indicators that don't rely solely on color
- Readable fonts and appropriate sizing

## Future Enhancements

### Advanced Features
- **Plugin System**: Extensible architecture for custom views
- **Configuration**: User-customizable layouts and key bindings
- **Themes**: Multiple color schemes and visual themes
- **Export**: Export data and reports in various formats
- **Remote Monitoring**: Monitor multiple server instances
- **Integration**: Webhook notifications and external tool integration

### Performance Monitoring
- **Metrics Dashboard**: Advanced performance visualization
- **Alerting**: Configurable alerts for system conditions
- **Reporting**: Automated reports and summaries
- **Profiling**: Built-in performance profiling tools