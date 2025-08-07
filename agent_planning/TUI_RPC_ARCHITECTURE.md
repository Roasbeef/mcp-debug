# TUI/Server gRPC Architecture Design

## Executive Summary

This document outlines the architectural transformation of the MCP Debug Server's TUI from an embedded component to a standalone monitoring client that communicates via gRPC. This enables remote monitoring, multiple concurrent clients, and proper separation between the UI and server components.

## Problem Statement

The current TUI implementation has critical limitations:
- **Process Isolation**: TUI and MCP server run as separate processes with isolated actor systems
- **No State Sharing**: Sessions created via MCP tools are invisible to the TUI
- **No Event Propagation**: Debug events cannot reach the TUI for real-time updates
- **Local Only**: TUI must run on the same machine as the server
- **Single Instance**: Only one TUI can monitor the server

## Proposed Solution: gRPC-Based Architecture

### Why gRPC?

After evaluating options (gRPC vs JSON-RPC), gRPC is superior for our use case:
- **Built-in Streaming**: Native support for server-side, client-side, and bidirectional streaming
- **Performance**: Protocol Buffers are 3-10x faster than JSON
- **Real-time Updates**: Server can push events without polling
- **Type Safety**: Strong typing through protobuf definitions
- **Language Agnostic**: Clients can be written in any supported language

## Architecture Overview

```
┌─────────────────┐         gRPC          ┌──────────────┐
│   TUI Console   │◄──────────────────────►│  Debug Server│
│  (Bubble Tea)   │      Streaming         │  (MCP + DAP) │
└─────────────────┘                        └──────────────┘
        │                                           │
        │                                           │
    ┌───▼────┐                              ┌──────▼──────┐
    │ gRPC   │                              │ gRPC Server │
    │ Client │                              │   + Event   │
    └────────┘                              │     Bus     │
                                            └─────────────┘
                                                    │
                                            ┌───────▼────────┐
                                            │ Actor System   │
                                            │   (Sessions,   │
                                            │   Debuggers)   │
                                            └────────────────┘
```

## Detailed Design

### 1. Protocol Buffer Service Definition

```protobuf
syntax = "proto3";

package debugmonitor;
option go_package = "github.com/roasbeef/mcp-debug/proto/debugmonitor";

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";
import "google/protobuf/empty.proto";

// Main monitoring service
service DebugMonitor {
  // Server status and metrics
  rpc GetServerStatus(google.protobuf.Empty) returns (ServerStatus);
  rpc StreamMetrics(google.protobuf.Empty) returns (stream MetricsUpdate);
  
  // Session management
  rpc ListSessions(google.protobuf.Empty) returns (SessionList);
  rpc GetSession(SessionRequest) returns (SessionDetails);
  rpc StreamSessionEvents(google.protobuf.Empty) returns (stream SessionEvent);
  rpc TerminateSession(SessionRequest) returns (OperationResult);
  
  // Client connection monitoring
  rpc ListClients(google.protobuf.Empty) returns (ClientList);
  rpc StreamClientEvents(google.protobuf.Empty) returns (stream ClientEvent);
  
  // Command execution (for interactive debugging)
  rpc ExecuteCommand(CommandRequest) returns (CommandResponse);
  rpc StreamCommandHistory(HistoryRequest) returns (stream CommandEntry);
  
  // Log streaming
  rpc StreamLogs(LogFilter) returns (stream LogEntry);
  rpc GetRecentLogs(LogQuery) returns (LogBatch);
  
  // Debug events (breakpoints, steps, etc.)
  rpc StreamDebugEvents(DebugEventFilter) returns (stream DebugEvent);
}

// Core message types
message ServerStatus {
  enum State {
    UNKNOWN = 0;
    STARTING = 1;
    RUNNING = 2;
    ERROR = 3;
    STOPPING = 4;
    STOPPED = 5;
  }
  
  State state = 1;
  google.protobuf.Timestamp started_at = 2;
  string version = 3;
  string actor_system_status = 4;
  int32 port = 5;
}

message MetricsUpdate {
  google.protobuf.Timestamp timestamp = 1;
  int64 uptime_seconds = 2;
  int32 active_sessions = 3;
  int32 connected_clients = 4;
  int64 total_requests = 5;
  int64 error_count = 6;
  double requests_per_second = 7;
  int64 memory_bytes = 8;
  int32 goroutine_count = 9;
}

message SessionDetails {
  string session_id = 1;
  string client_id = 2;
  string program_path = 3;
  int32 process_id = 4;
  
  enum Status {
    UNKNOWN = 0;
    CREATED = 1;
    INITIALIZED = 2;
    LAUNCHING = 3;
    RUNNING = 4;
    PAUSED = 5;
    TERMINATED = 6;
    ERROR = 7;
  }
  Status status = 5;
  
  repeated Breakpoint breakpoints = 6;
  repeated Thread threads = 7;
  Thread current_thread = 8;
  
  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp last_activity = 10;
  
  map<string, string> metadata = 11;
}

message SessionEvent {
  enum Type {
    UNKNOWN = 0;
    SESSION_CREATED = 1;
    SESSION_INITIALIZED = 2;
    PROGRAM_LAUNCHED = 3;
    BREAKPOINT_SET = 4;
    BREAKPOINT_HIT = 5;
    EXECUTION_CONTINUED = 6;
    EXECUTION_STEPPED = 7;
    EXECUTION_PAUSED = 8;
    SESSION_TERMINATED = 9;
    ERROR_OCCURRED = 10;
  }
  
  Type type = 1;
  string session_id = 2;
  google.protobuf.Timestamp timestamp = 3;
  google.protobuf.Any details = 4;  // Type-specific details
  string description = 5;
}

message Breakpoint {
  int32 id = 1;
  string file = 2;
  int32 line = 3;
  bool verified = 4;
  string condition = 5;
  int32 hit_count = 6;
}

message Thread {
  int32 id = 1;
  string name = 2;
  enum State {
    UNKNOWN = 0;
    RUNNING = 1;
    STOPPED = 2;
    WAITING = 3;
  }
  State state = 3;
  string stop_reason = 4;
  repeated StackFrame frames = 5;
}

message StackFrame {
  int32 id = 1;
  string function = 2;
  string file = 3;
  int32 line = 4;
  int32 column = 5;
}
```

### 2. Server-Side Implementation

#### Event Bus Architecture

The event bus serves as the central nervous system for event distribution:

```go
// rpc/event_bus.go
type EventBus struct {
    mu sync.RWMutex
    
    // Subscribers organized by event type
    sessionSubscribers []chan *SessionEvent
    clientSubscribers  []chan *ClientEvent
    logSubscribers     []chan *LogEntry
    debugSubscribers   []chan *DebugEvent
    metricsSubscribers []chan *MetricsUpdate
    
    // Registries for state tracking
    sessionRegistry *SessionRegistry
    clientRegistry  *ClientRegistry
    
    // Actor system reference
    actorSystem *actor.ActorSystem
}

// Subscribe to session events
func (eb *EventBus) SubscribeToSessions() (<-chan *SessionEvent, func()) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    ch := make(chan *SessionEvent, 100)
    eb.sessionSubscribers = append(eb.sessionSubscribers, ch)
    
    // Return channel and unsubscribe function
    return ch, func() {
        eb.unsubscribeSession(ch)
    }
}

// Publish session event to all subscribers
func (eb *EventBus) PublishSessionEvent(event *SessionEvent) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    
    for _, ch := range eb.sessionSubscribers {
        select {
        case ch <- event:
        default:
            // Channel full, skip slow consumers
        }
    }
}
```

#### Actor System Integration

Hook into the actor system to capture events:

```go
// rpc/actor_monitor.go
type ActorMonitor struct {
    eventBus *EventBus
    system   *actor.ActorSystem
}

// Monitor debugger commands and responses
func (am *ActorMonitor) MonitorDebugger(debugger actor.ActorRef[*debugger.DebuggerCmd, *debugger.DebuggerResp]) {
    // Create monitoring actor
    monitor := &DebuggerMonitor{
        eventBus: am.eventBus,
        target:   debugger,
    }
    
    // Register with actor system
    monitorRef := actor.RegisterWithSystem(am.system, "debugger-monitor", monitor)
    
    // Intercept messages (using actor system hooks or proxy pattern)
    // This would require extending the actor system with monitoring capabilities
}

type DebuggerMonitor struct {
    eventBus *EventBus
    target   actor.ActorRef[*debugger.DebuggerCmd, *debugger.DebuggerResp]
}

func (dm *DebuggerMonitor) Receive(ctx context.Context, cmd *debugger.DebuggerCmd) fn.Result[*debugger.DebuggerResp] {
    // Forward to actual debugger
    result := dm.target.Ask(ctx, cmd)
    
    // Extract and publish events based on command type
    switch c := cmd.Cmd.(type) {
    case *debugger.CreateSessionCmd:
        dm.eventBus.PublishSessionEvent(&SessionEvent{
            Type:      SessionEvent_SESSION_CREATED,
            Timestamp: timestamppb.Now(),
        })
    }
    
    return result
}
```

#### Session Registry

Track all debug sessions with metadata:

```go
// rpc/session_registry.go
type SessionRegistry struct {
    mu       sync.RWMutex
    sessions map[string]*SessionInfo
}

type SessionInfo struct {
    SessionID   string
    ClientID    string
    Program     string
    ProcessID   int32
    Status      SessionStatus
    Breakpoints []*Breakpoint
    Threads     []*Thread
    CreatedAt   time.Time
    LastActivity time.Time
    Metadata    map[string]string
}

func (sr *SessionRegistry) RegisterSession(sessionID string, info *SessionInfo) {
    sr.mu.Lock()
    defer sr.mu.Unlock()
    sr.sessions[sessionID] = info
}

func (sr *SessionRegistry) UpdateSessionStatus(sessionID string, status SessionStatus) {
    sr.mu.Lock()
    defer sr.mu.Unlock()
    if session, ok := sr.sessions[sessionID]; ok {
        session.Status = status
        session.LastActivity = time.Now()
    }
}
```

### 3. TUI Client Implementation

#### gRPC Client Manager

Handles connection and reconnection:

```go
// tui/grpc_client.go
type GRPCClient struct {
    conn   *grpc.ClientConn
    client debugmonitor.DebugMonitorClient
    
    // Stream contexts and cancelers
    metricsCancel  context.CancelFunc
    sessionsCancel context.CancelFunc
    logsCancel     context.CancelFunc
    
    // Reconnection
    reconnectCh chan struct{}
    connected   bool
    mu          sync.RWMutex
}

func NewGRPCClient(address string) (*GRPCClient, error) {
    // Connection with retry interceptor
    conn, err := grpc.Dial(address,
        grpc.WithInsecure(),
        grpc.WithUnaryInterceptor(retryInterceptor),
        grpc.WithStreamInterceptor(streamRetryInterceptor),
    )
    if err != nil {
        return nil, err
    }
    
    return &GRPCClient{
        conn:        conn,
        client:      debugmonitor.NewDebugMonitorClient(conn),
        reconnectCh: make(chan struct{}, 1),
    }, nil
}

// Subscribe to session events
func (gc *GRPCClient) SubscribeToSessions(ctx context.Context) (<-chan *SessionEvent, error) {
    stream, err := gc.client.StreamSessionEvents(ctx, &empty.Empty{})
    if err != nil {
        return nil, err
    }
    
    events := make(chan *SessionEvent, 10)
    
    go func() {
        defer close(events)
        for {
            event, err := stream.Recv()
            if err != nil {
                // Handle reconnection
                if err == io.EOF || status.Code(err) == codes.Unavailable {
                    gc.triggerReconnect()
                }
                return
            }
            
            select {
            case events <- event:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return events, nil
}
```

#### Bubble Tea Integration

Convert gRPC streams to Bubble Tea messages:

```go
// tui/stream_commands.go

// Command to subscribe to session events
func subscribeToSessions(client *GRPCClient) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        events, err := client.SubscribeToSessions(ctx)
        if err != nil {
            return ErrorMsg{Err: err}
        }
        
        // Return a command that waits for events
        return waitForSessionEvent(events)()
    }
}

// Command that waits for a session event
func waitForSessionEvent(events <-chan *SessionEvent) tea.Cmd {
    return func() tea.Msg {
        event, ok := <-events
        if !ok {
            // Stream closed, trigger reconnection
            return ReconnectMsg{}
        }
        return SessionUpdateMsg{Event: event}
    }
}

// Updated TUI model
type ImprovedTUIModel struct {
    // ... existing fields ...
    
    grpcClient    *GRPCClient
    sessions      map[string]*SessionInfo
    lastMetrics   *MetricsUpdate
    logBuffer     []LogEntry
    
    // Subscription management
    subscriptions struct {
        sessions bool
        metrics  bool
        logs     bool
    }
}

// Handle messages in Update method
func (m *ImprovedTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case SessionUpdateMsg:
        // Update session data
        m.updateSessionFromEvent(msg.Event)
        m.refreshSessionsTable()
        
        // Continue listening for events
        cmds = append(cmds, waitForSessionEvent(m.sessionEvents))
        
    case MetricsUpdateMsg:
        // Update metrics display
        m.lastMetrics = msg.Metrics
        m.updateDashboard()
        
        // Continue listening
        cmds = append(cmds, waitForMetricsUpdate(m.metricsEvents))
        
    case LogEntryMsg:
        // Add to log buffer
        m.logBuffer = append(m.logBuffer, msg.Entry)
        if len(m.logBuffer) > 1000 {
            m.logBuffer = m.logBuffer[100:] // Keep last 900
        }
        m.updateLogsViewport()
        
        // Continue listening
        cmds = append(cmds, waitForLogEntry(m.logEvents))
        
    case ReconnectMsg:
        // Handle reconnection
        m.serverStatus = ServerReconnecting
        cmds = append(cmds, m.reconnectToServer())
    }
    
    // ... handle other messages ...
    
    return m, tea.Batch(cmds...)
}

// Initialize subscriptions
func (m *ImprovedTUIModel) Init() tea.Cmd {
    return tea.Batch(
        m.connectToServer(),
        textinput.Blink,
    )
}

func (m *ImprovedTUIModel) connectToServer() tea.Cmd {
    return func() tea.Msg {
        client, err := NewGRPCClient("localhost:50051")
        if err != nil {
            return ErrorMsg{Err: err}
        }
        
        m.grpcClient = client
        
        // Start subscriptions
        return tea.Batch(
            subscribeToSessions(client),
            subscribeToMetrics(client),
            subscribeToLogs(client),
        )
    }
}
```

### 4. Implementation Phases

#### Phase 1: Foundation (Week 1)
1. **Day 1-2**: Create protobuf definitions and generate code
2. **Day 3-4**: Implement basic gRPC server and event bus
3. **Day 5**: Create session and client registries

#### Phase 2: Server Integration (Week 2)  
1. **Day 1-2**: Hook event bus into actor system
2. **Day 3-4**: Implement all streaming endpoints
3. **Day 5**: Add metrics collection

#### Phase 3: TUI Client (Week 3)
1. **Day 1-2**: Add gRPC client with reconnection
2. **Day 3-4**: Convert data fetching to streaming
3. **Day 5**: Update Bubble Tea commands

#### Phase 4: Testing & Polish (Week 4)
1. **Day 1-2**: Integration testing
2. **Day 3-4**: Performance optimization
3. **Day 5**: Documentation and examples

## Benefits

### Immediate Benefits
1. **Remote Monitoring**: Monitor debug server from any machine
2. **Multiple Clients**: Multiple TUIs can connect simultaneously
3. **Real-time Updates**: Push-based updates instead of polling
4. **Proper Separation**: Clean architectural boundaries

### Long-term Benefits
1. **Extensibility**: Easy to add new monitoring features
2. **Alternative Clients**: Web UI, mobile apps, CLI tools
3. **Metrics Integration**: Export to Prometheus, Grafana
4. **Collaboration**: Multiple developers debugging together

## Migration Strategy

### Backward Compatibility
- Keep existing embedded TUI mode as fallback
- Add `--grpc` flag to enable new architecture
- Gradual migration of features

### Configuration
```yaml
# Server configuration
server:
  grpc:
    enabled: true
    port: 50051
    max_connections: 10
    tls:
      enabled: false
      cert_file: ""
      key_file: ""

# TUI configuration  
tui:
  server:
    address: "localhost:50051"
    reconnect:
      enabled: true
      interval: 5s
      max_attempts: 10
```

## Testing Strategy

### Unit Tests
- Test each gRPC service method
- Test event bus publish/subscribe
- Test registry operations

### Integration Tests
- End-to-end TUI ↔ Server communication
- Stream reliability under load
- Reconnection scenarios

### Performance Tests
- Multiple concurrent TUI clients
- High-frequency event streaming
- Memory and CPU usage

## Security Considerations

### Authentication
- TLS for encrypted communication
- Token-based authentication
- Role-based access control (future)

### Rate Limiting
- Limit events per second per client
- Circuit breaker for misbehaving clients
- Resource quotas

## Future Enhancements

### Near-term (3-6 months)
1. **Web UI**: gRPC-Web for browser monitoring
2. **Record/Replay**: Record debug sessions
3. **Filtering**: Advanced event filtering

### Long-term (6-12 months)
1. **AI Integration**: Debugging assistance
2. **Collaborative Debugging**: Shared sessions
3. **Cloud Deployment**: SaaS offering

## Conclusion

This gRPC-based architecture transforms the TUI from a limited embedded component to a powerful, flexible monitoring system. The streaming capabilities of gRPC combined with Bubble Tea's reactive model create a responsive, real-time debugging experience that can scale to support multiple users and remote debugging scenarios.

The implementation is ambitious but achievable, with clear phases and testing strategies. The benefits far outweigh the complexity, positioning the MCP Debug Server as a professional-grade debugging solution.