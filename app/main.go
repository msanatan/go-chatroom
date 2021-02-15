package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/msanatan/go-chatroom/app/auth"
	"github.com/msanatan/go-chatroom/app/models"
	"github.com/msanatan/go-chatroom/app/service"
	"github.com/msanatan/go-chatroom/rabbitmq"
	"github.com/msanatan/go-chatroom/utils"
	"github.com/streadway/amqp"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	logger := utils.InitLogger(logLevel, "chatroom")
	var wsServer *service.Server
	var rabbitMQClient *rabbitmq.Client

	// Setup Postgres database
	var dbPort int
	dbPortString := os.Getenv("POSTGRES_PORT")
	if dbPortString == "" {
		dbPort = 5432
	} else {
		var convErr error
		dbPort, convErr = strconv.Atoi(dbPortString)
		if convErr != nil {
			logger.Fatalf("%s is not a valid Postgres port number", dbPortString)
		}
	}
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"), dbPort, os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Fatalf("could not connect to Postgres DB: %s", err.Error())
	}

	dbClient, err := models.NewChatroomDB(db, logger)
	if err != nil {
		logger.Fatalf("could not connect to Postgres DB: %s", err.Error())
	}

	err = dbClient.Migrate()
	if err != nil {
		logger.Fatalf("could not complete DB migrations: %s", err.Error())
	}

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

		rabbitMQClient = rabbitmq.NewClient(ch, "command_queue", "response_queue", logger)
		err = rabbitMQClient.QueueDeclare()
		if err != nil {
			logger.Fatalf("could not declare queues: %s", err.Error())
		}
	}

	wsServer = service.NewServer(rabbitMQClient, "/", logger)
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

	staticFiles := os.Getenv("STATIC_FILES")
	if staticFiles == "" {
		staticFiles = "./public/"
	}

	// Create auth client
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Fatalf("Missing JWT_SECRET env var")
	}
	authClient := auth.NewClient(dbClient, jwtSecret, logger)

	r := mux.NewRouter()
	r.HandleFunc("/login", authClient.Login).Methods("POST")
	r.HandleFunc("/register", authClient.CreateUser).Methods("POST")

	protected := r.PathPrefix("/api").Subrouter()
	protected.HandleFunc("/ws", service.ServeWs(wsServer, defaultClientConfig, logger))
	protected.Use(authClient.IsAuthenticated)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticFiles)))

	// Keep a log of all incoming requests
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug(r.RequestURI)
			next.ServeHTTP(w, r)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Debugf("Running at http://localhost:%s", port)
	logger.Fatal(http.ListenAndServe(":"+port, r))
}
