package server

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// SetupRoutes initializes all routes for the signaling server.
func SetupRoutes(log *logrus.Logger) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		HandleConnection(w, r, log) 
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("this is hello endpoint"))
	})
}
