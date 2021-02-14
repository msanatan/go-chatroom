package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/msanatan/go-chatroom/app/rabbitmq"
	"github.com/msanatan/go-chatroom/app/service"
	"github.com/msanatan/go-chatroom/utils"
	"github.com/streadway/amqp"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	logger := utils.InitLogger(logLevel, "chatroom")
	var wsServer *service.Server
	var rabbitMQClient *rabbitmq.Client

	rabbitConnection := os.Getenv("RABBITMQ_CONNECTION")
	if rabbitConnection == "" {
		logger.Error("No RabbitMQ connection string provided, will not setup connection")
	} else {
		conn, err := amqp.Dial(rabbitConnection)
		if err != nil {
			logger.Fatalf("could not connect to RabbitMQ instance: %s", err.Error())
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			logger.Fatalf("could not open a channel: %s", err.Error())
		}
		defer ch.Close()

		rabbitMQClient = rabbitmq.NewClient(ch, logger)
		err = rabbitMQClient.QueueDeclare()
		if err != nil {
			logger.Fatalf("could not declare queues: %s", err.Error())
		}
	}

	wsServer = service.NewServer(rabbitMQClient, nil, "/", logger)
	go wsServer.Run()
	if rabbitMQClient != nil {
		go wsServer.ConsumeRMQ()
	}

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
