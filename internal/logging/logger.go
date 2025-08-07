package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// InitFileLogger initializes a logger that writes to a file in ~/.dlv-mcp-server
func InitFileLogger() (*os.File, error) {
	// Create log directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".dlv-mcp-server")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("debug_%s.log", timestamp))
	
	// Also create a symlink to latest log
	latestLink := filepath.Join(logDir, "latest.log")
	os.Remove(latestLink) // Remove old symlink if exists
	
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create symlink to latest log (ignore errors as it's not critical)
	os.Symlink(logFile, latestLink)

	// Set default logger to write to both file and stdout
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	
	log.Printf("========================================")
	log.Printf("MCP Debug Server started at %s", time.Now().Format(time.RFC3339))
	log.Printf("Log file: %s", logFile)
	log.Printf("========================================")
	
	fmt.Printf("Logging to: %s\n", logFile)
	
	return file, nil
}