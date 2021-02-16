package service

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/msanatan/go-chatroom/app/models"
	"github.com/msanatan/go-chatroom/rabbitmq"
	log "github.com/sirupsen/logrus"
)

// Server is our hub for all WS clients
type Server struct {
	clients        map[*WSClient]bool
	register       chan *WSClient
	deregister     chan *WSClient
	broadcast      chan MessagePayload
	rabbitMQClient *rabbitmq.Client
	chatroomDB     *models.ChatroomDB
	botSymbol      string
	logger         *log.Entry
}

// NewServer instantiates a new server struct
func NewServer(rabbitMQClient *rabbitmq.Client, chatroomDB *models.ChatroomDB,
	botSymbol string, logger *log.Entry) *Server {
	if botSymbol == "" {
		botSymbol = "/"
	}

	return &Server{
		clients:        make(map[*WSClient]bool),
		register:       make(chan *WSClient),
		deregister:     make(chan *WSClient),
		broadcast:      make(chan MessagePayload),
		rabbitMQClient: rabbitMQClient,
		chatroomDB:     chatroomDB,
		botSymbol:      botSymbol,
		logger:         logger,
	}
}

func (s *Server) registerClient(client *WSClient) {
	s.clients[client] = true
}

func (s *Server) deregisterClient(client *WSClient) {
	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
	}
}

func (s *Server) broadcastToClients(message MessagePayload) {
	for client := range s.clients {
		client.send <- message
	}
}

// Run executes our websocket server to accpet its various requests
func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.registerClient(client)
		case client := <-s.deregister:
			s.deregisterClient(client)
		case message := <-s.broadcast:
			s.broadcastToClients(message)
		}
	}
}

// ClientCount returns the number of connected clients
func (s *Server) ClientCount() int {
	return len(s.clients)
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
func (s *Server) ConsumeRMQ() {
	logger := s.logger.WithField("method", "ConsumeRMQ")
	msgs, err := s.rabbitMQClient.Consume()
	if err != nil {
		logger.Errorf("could not consume response_queue messages: %s", err.Error())
		return
	}

	logger.Debug("waiting on messages from RabbitMQ")
	for msg := range msgs {
		logger.Debugf("received message: %s", string(msg.Body))
		var message MessagePayload
		err = json.Unmarshal(msg.Body, &message)
		if err != nil {
			logger.Errorf("message not in correct format: %s", err.Error())
			continue
		}

		s.broadcast <- message
	}
}

// IsValidBotCommand verifies if a message should be treated as a bot command
func (s *Server) IsValidBotCommand(message string) bool {
	return len(message) > 0 &&
		strings.HasPrefix(message, s.botSymbol) &&
		!strings.HasPrefix(message, s.botSymbol+s.botSymbol)
}

// ExtractCommandAndArgs parses a bot command and any arguments it may have
func (s *Server) ExtractCommandAndArgs(message string) (string, string) {
	if strings.Contains(message, "=") {
		commandString := message[strings.Index(message, s.botSymbol)+1 : strings.Index(message, "=")]
		args := strings.SplitN(message, "=", 2)[1]
		return commandString, args
	}

	return strings.SplitN(message, s.botSymbol, 2)[1], ""
}
