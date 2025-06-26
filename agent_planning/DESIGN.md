# MCP Debugger Design

This document outlines the design for an MCP (Machine Control Plane) server that exposes a debugging service. This service will allow agents and developers to debug complex systems engineering programs, initially targeting Go applications, using the Delve debugger.

## 1. Introduction

The goal of this project is to create a remote, programmatic debugging environment built on top of the Machine Control Plane (MCP). This will enable automated agents and developers to debug applications running in diverse environments without needing direct access to the machine or a traditional IDE. By exposing a debugger through an MCP interface, we can create powerful new debugging workflows and tools.

We leverage the [Debug Adapter Protocol (DAP)](https://microsoft.github.io/debug-adapter-protocol/) to provide a standardized interface for debugging features, combined with an actor-based architecture for type-safe concurrent session management.

## 2. Architecture

### 2.1 High-Level System Architecture

The system consists of the following components:

```
+-----------------+      +-----------------+      +-----------------+
|   MCP Client    | <--> |   MCP Server    | <--> |   Actor System  |
| (Agent/User)    |      |  (This Project) |      | (Session Mgmt)  |
+-----------------+      +-----------------+      +-----------------+
                                                       |
                                                       v
+-----------------+      +-----------------+      +-----------------+
|      TUI        | <--> |   Debugger      | <--> |    Session      |
|   (Optional)    |      |    Factory      |      |     Actor       |
+-----------------+      +-----------------+      +-----------------+
                                                       |
                                                       v
                                                +-----------------+
                                                | Delve DAP       |
                                                | (via TCP)       |
                                                +-----------------+
                                                       |
                                                       v
                                                +-----------------+
                                                | Target Process  |
                                                +-----------------+
```

### 2.2 Actor-Based Architecture

The system uses an actor-based architecture for managing concurrent debugging sessions:

- **Actor System**: Central coordinator for all actors, provides service discovery via Receptionist
- **Debugger Factory Actor**: Creates and manages debugging sessions
- **Session Actors**: Individual actors that handle DAP communication with Delve processes
- **Message Types**: Type-safe wrappers for DAP requests and responses

### 2.3 Component Details

- **MCP Client:** Any MCP-compliant client that can send messages to the server
- **MCP Server:** The core service exposing debugging functionality through MCP protocol  
- **Debugger Factory Actor:** Creates new debugging sessions and manages their lifecycle
- **Session Actor:** Manages individual DAP debugging sessions with Delve processes
- **Delve DAP Server:** The Go debugger running in DAP mode via TCP connection
- **Target Process:** The Go program being debugged

## 3. Actor-Based Message System

### 3.1 Type-Safe Message Wrappers

The system implements type-safe message wrappers for all DAP communication:

**DAPRequest** - Wraps DAP requests sent to Session actors:
```go
type DAPRequest struct {
    actor.BaseMessage
    Request dap.Message  // Raw DAP request (e.g., dap.InitializeRequest)
}
```

**DAPResponse** - Wraps DAP responses from Session actors:
```go
type DAPResponse struct {
    actor.BaseMessage
    Response dap.Message  // Raw DAP response (e.g., dap.InitializeResponse)
}
```

### 3.2 Helper Functions for Clean API

The system provides type-safe helper functions that encapsulate the actor communication:

```go
func InitializeSession(session ActorRef[*DAPRequest, *DAPResponse], clientID string) (*dap.InitializeResponse, error)
func Launch(session ActorRef[*DAPRequest, *DAPResponse], program string, args []string) (*dap.LaunchResponse, error)
func SetBreakpoints(session ActorRef[*DAPRequest, *DAPResponse], source string, lines []int) (*dap.SetBreakpointsResponse, error)
```

These functions:
- Create properly typed DAP request messages
- Send them to the Session actor using Ask pattern
- Handle response unwrapping and type checking
- Provide clean API for consumers (TUI, MCP server)

### 3.3 Session Actor Implementation

Each Session actor:
- Manages one Delve DAP process via TCP connection
- Receives `DAPRequest` messages containing raw DAP requests
- Forwards requests to Delve DAP server
- Returns `DAPResponse` messages with raw DAP responses
- Handles concurrent request/response correlation
- Manages session lifecycle and cleanup

## 4. Debug Adapter Protocol (DAP)

We use the Debug Adapter Protocol as the communication standard between Session actors and Delve debugger:

- **Standardization:** DAP is widely adopted, enabling future support for different debuggers/languages
- **Rich Feature Set:** Supports breakpoints, stepping, variable inspection, expression evaluation, etc.
- **Library Support:** Uses `github.com/google/go-dap` for message serialization/deserialization
- **TCP Communication:** Direct TCP connection to Delve DAP server for performance

## 5. Session and Process Management

### 5.1 Debugger Factory Actor

The Debugger Factory Actor (`DebuggerKey` service) handles:
- Creating new debugging sessions on demand
- Assigning unique session IDs (`session-1`, `session-2`, etc.)
- Registering Session actors with the Actor System
- Managing session lifecycle

**Message Types:**
```go
type DebuggerMsg struct {
    Command DebuggerCommand  // StartDebuggerCmd, StopDebuggerCmd
}

type DebuggerResp struct {
    Status string  // Returns session ID for new sessions
}
```

### 5.2 Session Management

Each debugging session:
- **Delve Process:** Launches new `dlv dap` process automatically
- **TCP Connection:** Establishes connection to Delve DAP server
- **Concurrent Communication:** Handles multiple simultaneous DAP requests
- **Cleanup:** Terminates Delve process and closes connections on session end

### 5.3 Target Process Handling

Sessions support two debugging modes:
- **Launch Mode:** Start new target process with specified binary and arguments
- **Attach Mode:** Attach to existing running process by PID

Implementation is handled through DAP `LaunchRequest` and `AttachRequest` messages.

## 6. MCP Integration

The MCP server exposes debugging functionality through standard MCP protocol:

### 6.1 Service Discovery
- MCP clients discover available debugging capabilities
- Session management through MCP resource URIs
- Type-safe operation definitions

### 6.2 Operation Mapping
MCP operations map to DAP requests via helper functions:
- `debug.initialize()` → `InitializeSession()` → `dap.InitializeRequest`
- `debug.launch()` → `Launch()` → `dap.LaunchRequest`  
- `debug.setBreakpoints()` → `SetBreakpoints()` → `dap.SetBreakpointsRequest`
- `debug.continue()` → `Continue()` → `dap.ContinueRequest`
- `debug.stackTrace()` → `StackTrace()` → `dap.StackTraceRequest`

### 6.3 Error Handling
- Actor communication failures
- DAP protocol errors
- Session lifecycle errors
- Delve process management errors

## 7. Implementation Benefits

### 7.1 Type Safety
- Compile-time verification of message types
- No runtime message routing errors
- Clear API contracts between components

### 7.2 Concurrency
- Multiple simultaneous debugging sessions
- Per-session actor isolation
- Built-in backpressure and message queuing

### 7.3 Maintainability  
- Clean separation of concerns
- Actor system handles complexity
- Helper functions provide simple APIs
- Testable components in isolation

## 8. Future Work

- **Multi-language support:** Extend to other DAP-compatible debuggers
- **Security:** Authentication and authorization for MCP access
- **Remote File Access:** Source code retrieval through MCP resources
- **Advanced Features:** Conditional breakpoints, watchpoints, memory inspection
- **Performance:** Connection pooling, session persistence
