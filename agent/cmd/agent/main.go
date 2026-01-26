// Package main is the entry point for the Smart Monitor Agent
package main

import (
	"log"
	"os"

	"smart-agent/internal/agent"
	"smart-agent/internal/config"
)

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg := config.DefaultConfig()

	// Create agent
	agentInstance, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
		os.Exit(1)
	}

	// Start agent
	if err := agentInstance.Start(); err != nil {
		log.Fatalf("Agent error: %v", err)
		os.Exit(1)
	}
}
