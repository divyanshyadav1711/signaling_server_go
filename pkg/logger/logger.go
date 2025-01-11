package logger

import (
	"github.com/sirupsen/logrus"
	"signaling_server/pkg/config"
)

func NewLogger(cfg *config.Config) *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
	return log
}
