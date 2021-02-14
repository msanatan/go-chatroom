package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/msanatan/go-chatroom/app/service"
	"github.com/msanatan/go-chatroom/utils"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	logger := utils.InitLogger(logLevel, "chatroom")
	wsServer := service.NewServer(nil, "/", logger)
	go wsServer.Run()

	// Setup default WS Client config
	defaultClientConfig := &service.ClientConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     (60 * time.Second * 9) / 10,
		MaxMessageSize: 10000,
	}

	r := mux.NewRouter()
	r.HandleFunc("/ws", service.ServeWs(wsServer, defaultClientConfig, logger))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Debugf("Running at http://localhost:%s", port)
	logger.Fatal(http.ListenAndServe(":"+port, r))
}
