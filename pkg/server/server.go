package server

import (
	"net/http"
	"signaling_server/pkg/config"

	"github.com/sirupsen/logrus"
)

// Start starts the signaling server without SSL/TLS.
func Start(cfg *config.Config, log *logrus.Logger) {
	// Setup routes (unchanged)
	SetupRoutes(log)

	// Define the server address
	address := "0.0.0.0:" + cfg.ServerPort

	// Log the server start message
	log.Infof("Server running on http://%s", address)

	// Start the HTTP server without SSL/TLS
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
