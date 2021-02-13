package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/msanatan/go-chatroom/websockets"
	log "github.com/sirupsen/logrus"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	logger := initLogger(logLevel)
	wsServer := websockets.NewServer(logger)
	go wsServer.Run()

	// Setup default WS Client config
	defaultClientConfig := &websockets.ClientConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     (60 * time.Second * 9) / 10,
		MaxMessageSize: 10000,
	}

	r := mux.NewRouter()
	r.HandleFunc("/ws", websockets.ServeWs(wsServer, defaultClientConfig, logger))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Debugf("Running at http://localhost:%s", port)
	logger.Fatal(http.ListenAndServe(":"+port, r))
}

func initLogger(logLevel string) *log.Entry {
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
	return parentLogger.WithField("application", "go-chatroom")
}
