package main

import (
	"log"
	"os"

	mcpdebug "github.com/roasbeef/mcp-debug"
)

func main() {
	// Launch TUI with proper Bubble Tea components and actor router
	log.Println("Starting MCP Debug Server TUI Console...")
	if err := mcpdebug.RunTUI(); err != nil {
		log.Fatalf("TUI failed: %v", err)
		os.Exit(1)
	}
}