package server

import (
	"net/http"
	"signaling_server/pkg/config"

	"github.com/sirupsen/logrus"
)

// Start starts the signaling server.

func Start(cfg *config.Config, log *logrus.Logger) {
	// Initialize routes
	SetupRoutes(log)

	// Bind to all interfaces (0.0.0.0) and use the configured port
	log.Infof("Server running on port %s", cfg.ServerPort)
	if err := http.ListenAndServe("0.0.0.0:"+cfg.ServerPort, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

