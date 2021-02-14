package service

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/msanatan/go-chatroom/app/rabbitmq"
	log "github.com/sirupsen/logrus"
)

// Server is our hub for all WS clients
type Server struct {
	clients        map[*WSClient]bool
	register       chan *WSClient
	deregister     chan *WSClient
	broadcast      chan MessagePayload
	rabbitMQClient *rabbitmq.Client
	bots           map[string]Bot
	botSymbol      string
	logger         *log.Entry
}

// NewServer instantiates a new server struct
func NewServer(rabbitMQClient *rabbitmq.Client, bots map[string]Bot, botSymbol string, logger *log.Entry) *Server {
	if bots == nil {
		bots = make(map[string]Bot)
	}

	if botSymbol == "" {
		botSymbol = "/"
	}

	return &Server{
		clients:        make(map[*WSClient]bool),
		register:       make(chan *WSClient),
		deregister:     make(chan *WSClient),
		broadcast:      make(chan MessagePayload),
		rabbitMQClient: rabbitMQClient,
		bots:           bots,
		botSymbol:      botSymbol,
		logger:         logger,
	}
}

func (server *Server) registerClient(client *WSClient) {
	server.clients[client] = true
}

func (server *Server) deregisterClient(client *WSClient) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
	}
}

func (server *Server) broadcastToClients(message MessagePayload) {
	for client := range server.clients {
		client.send <- message
	}
}

// Run executes our websocket server to accpet its various requests
func (server *Server) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.deregister:
			server.deregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

// ClientCount returns the number of connected clients
func (server *Server) ClientCount() int {
	return len(server.clients)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// ServeWs registers a WS client
func ServeWs(server *Server, clientConfig *ClientConfig, logger *log.Entry) http.HandlerFunc {
	logger = logger.WithField("method", "ServeWs")
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Errorf("error trying to setup websocket connection: %q", err.Error())
			return
		}

		logger.Debug("Creating new websocket client")
		client := NewWSClient(conn, server, clientConfig, logger, "main")

		go client.writeMessages()
		go client.readMessages()

		server.register <- client
	}
}

// ConsumeRMQ reads the RabbitMQ response queue and broadcasts it to clients
func (server *Server) ConsumeRMQ() {
	logger := server.logger.WithField("method", "ConsumeRMQ")
	msgs, err := server.rabbitMQClient.Consume()
	if err != nil {
		logger.Errorf("could not consume response_queue messages: %s", err.Error())
		return
	}

	logger.Debug("waiting on messages from RabbitMQ")
	for d := range msgs {
		logger.Debugf("received message: %s", string(d.Body))
		message := MessagePayload{
			Message: string(d.Body),
			Type:    d.Type,
		}

		server.broadcast <- message
	}
}
