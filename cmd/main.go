package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Configure logger to write to stderr
	log.SetOutput(os.Stderr)
	log.Println("Starting Temporal MCP...")

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// TODO: Initialize configuration
	// TODO: Setup Temporal client
	// TODO: Initialize services
	// TODO: Start API server

	log.Println("Temporal MCP is running. Press Ctrl+C to stop.")

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down...", sig)

	// TODO: Perform cleanup and graceful shutdown

	log.Println("Temporal MCP has been stopped.")
}
