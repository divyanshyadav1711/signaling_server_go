package main

import (
	"signaling_server/pkg/config"
	"signaling_server/pkg/logger"
	"signaling_server/pkg/server"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	log := logger.NewLogger(cfg)

	// Start the server
	log.Info("Starting signaling server...")
	server.Start(cfg, log)
}
