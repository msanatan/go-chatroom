package utils

import (
	log "github.com/sirupsen/logrus"
)

// InitLogger creates a new logger object
func InitLogger(logLevel, appName string) *log.Entry {
	parentLogger := log.New()
	var logrusLevel log.Level

	switch logLevel {
	case "trace":
		logrusLevel = log.TraceLevel
	case "debug":
		logrusLevel = log.DebugLevel
	case "info":
		logrusLevel = log.InfoLevel
	case "warn":
		logrusLevel = log.WarnLevel
	case "error":
		logrusLevel = log.ErrorLevel
	default:
		logrusLevel = log.DebugLevel
	}

	parentLogger.SetLevel(logrusLevel)
	parentLogger.SetFormatter(&log.JSONFormatter{})
	return parentLogger.WithField("application", appName)
}
