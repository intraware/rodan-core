package utils

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func NewLogger(production bool) {
	logger := logrus.New()
	logger.Out = os.Stdout

	if production {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			ForceColors:     true,
		})
		logger.SetLevel(logrus.DebugLevel)
	}
	Logger = logger
}
