package server

import (
	"net/http"
	"signaling_server/pkg/config"

	"github.com/sirupsen/logrus"
)

func Start(cfg *config.Config, log *logrus.Logger) {

	SetupRoutes(log)

	
	address := "0.0.0.0:" + cfg.ServerPort

	
	log.Infof("Server running on http://%s", address)

	
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
