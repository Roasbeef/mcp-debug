package main

import (
	"log"
	"os"

	mcpdebug "github.com/roasbeef/mcp-debug"
	"github.com/roasbeef/mcp-debug/internal/logging"
)

func main() {
	// Initialize file logging
	logFile, err := logging.InitFileLogger()
	if err != nil {
		log.Printf("Warning: Failed to initialize file logging: %v", err)
		// Continue without file logging
	} else {
		defer logFile.Close()
	}

	log.Println("Starting Go DAP MCP Server...")

	// Create MCP server with service management
	mcpServer, service := mcpdebug.NewMCPServer()
	defer service.Stop()

	// Start serving
	if err := mcpServer.Serve(); err != nil {
		log.Fatalf("MCP server error: %v", err)
		os.Exit(1)
	}
}