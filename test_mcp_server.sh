#!/bin/bash

# Test script to demonstrate MCP server functionality
# This script shows how to use the MCP server for debugging

echo "=== MCP Debug Server Test ==="
echo "This script demonstrates the MCP debugging server functionality."
echo

echo "1. Building MCP server..."
go build -o mcp-server ./cmd/mcp-server
if [ $? -ne 0 ]; then
    echo "Failed to build MCP server"
    exit 1
fi
echo "✓ MCP server built successfully"
echo

echo "2. Building example program for debugging..."
go build -o examples/simple/simple ./examples/simple
if [ $? -ne 0 ]; then
    echo "Failed to build example program"
    exit 1
fi
echo "✓ Example program built successfully"
echo

echo "3. MCP server is ready to accept debugging requests!"
echo "   Start the server with: ./mcp-server"
echo "   The server communicates via stdio using JSON-RPC 2.0"
echo

echo "Available MCP Tools:"
echo "  • create_debug_session    - Create a new debugging session"
echo "  • initialize_session      - Initialize DAP protocol"
echo "  • launch_program          - Launch a Go program for debugging"
echo "  • configuration_done      - Signal configuration complete"
echo "  • set_breakpoints         - Set source breakpoints"
echo "  • continue_execution      - Continue program execution"
echo "  • step_next              - Step over (next line)"
echo "  • step_in                - Step into functions"
echo "  • step_out               - Step out of functions"
echo "  • pause_execution        - Pause execution"
echo "  • get_threads            - Get thread information"
echo "  • get_stack_frames       - Get call stack"
echo "  • get_variables          - Get variable values"
echo "  • evaluate_expression    - Evaluate expressions"
echo

echo "4. Example usage workflow:"
echo "   a) create_debug_session {'session_id': 'debug1'}"
echo "   b) initialize_session {'session_id': 'debug1', 'client_id': 'client1'}"
echo "   c) launch_program {'session_id': 'debug1', 'program': './examples/simple/simple'}"
echo "   d) configuration_done {'session_id': 'debug1'}"
echo "   e) set_breakpoints {'session_id': 'debug1', 'file': './examples/simple/main.go', 'lines': [15]}"
echo "   f) continue_execution {'session_id': 'debug1', 'thread_id': 1}"
echo "   g) get_threads {'session_id': 'debug1'}"
echo "   h) get_variables {'session_id': 'debug1', 'frame_id': 1}"
echo

echo "5. Run tests to verify functionality:"
go test -v . > /dev/null
if [ $? -eq 0 ]; then
    echo "✓ All tests pass"
else
    echo "✗ Some tests failed"
fi

echo
echo "MCP Debug Server is ready for use!"
echo "Integration with LLM clients enables AI-powered debugging workflows."