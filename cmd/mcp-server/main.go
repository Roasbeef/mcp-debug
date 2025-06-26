package main

import (
	"log"
	"os"

	mcpdebug "github.com/roasbeef/mcp-debug"
)

func main() {
	// Create MCP server with service management
	mcpServer, service := mcpdebug.NewMCPServer()
	defer service.Stop()

	// Start serving
	if err := mcpServer.Serve(); err != nil {
		log.Fatalf("MCP server error: %v", err)
		os.Exit(1)
	}
}