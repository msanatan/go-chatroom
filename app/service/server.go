package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/msanatan/go-chatroom/app/models"
	"github.com/msanatan/go-chatroom/rabbitmq"
	log "github.com/sirupsen/logrus"
)

// Server is our hub for all WS clients
type Server struct {
	rooms          map[uint]map[*WSClient]bool
	register       chan *Subscription
	Deregister     chan *Subscription
	broadcast      chan MessagePayload
	rabbitMQClient *rabbitmq.Client
	chatroomDB     *models.ChatroomDB
	jwtSecret      string
	botSymbol      string
	logger         *log.Entry
}

// NewServer instantiates a new server struct
func NewServer(rabbitMQClient *rabbitmq.Client, chatroomDB *models.ChatroomDB,
	jwtSecret, botSymbol string, logger *log.Entry) *Server {
	if botSymbol == "" {
		botSymbol = "/"
	}

	return &Server{
		rooms:          make(map[uint]map[*WSClient]bool),
		register:       make(chan *Subscription),
		Deregister:     make(chan *Subscription),
		broadcast:      make(chan MessagePayload),
		rabbitMQClient: rabbitMQClient,
		chatroomDB:     chatroomDB,
		jwtSecret:      jwtSecret,
		botSymbol:      botSymbol,
		logger:         logger,
	}
}

func (s *Server) registerClient(subscription *Subscription) {
	if s.rooms[subscription.RoomID] == nil {
		s.rooms[subscription.RoomID] = make(map[*WSClient]bool)
	}

	s.rooms[subscription.RoomID][subscription.Client] = true
}

func (s *Server) deregisterClient(subscription *Subscription) {
	if _, ok := s.rooms[subscription.RoomID][subscription.Client]; ok {
		delete(s.rooms[subscription.RoomID], subscription.Client)
		close(subscription.Client.send)

		// Remove room from active connections in memory
		if len(s.rooms[subscription.RoomID]) == 0 {
			delete(s.rooms, subscription.RoomID)
		}
	}
}

func (s *Server) broadcastToClients(message MessagePayload) {
	for client := range s.rooms[message.RoomID] {
		select {
		case client.send <- message:
		default:
			s.deregisterClient(&Subscription{
				Client: client,
				RoomID: message.RoomID,
			})
		}
	}
}

// Run executes our websocket server to accpet its various requests
func (s *Server) Run() {
	for {
		select {
		case subscription := <-s.register:
			s.registerClient(subscription)
		case subscription := <-s.Deregister:
			s.deregisterClient(subscription)
		case message := <-s.broadcast:
			s.broadcastToClients(message)
		}
	}
}

// ClientCount returns the number of connected clients
func (s *Server) ClientCount(roomID uint) int {
	return len(s.rooms[roomID])
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// ServeWs registers a WS client
func ServeWs(server *Server, clientConfig *ClientConfig, logger *log.Entry) http.HandlerFunc {
	logger = logger.WithField("method", "ServeWs")
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID, err := strconv.ParseUint(vars["roomId"], 10, 32)
		if err != nil {
			logger.Errorf("room ID is not valid: %s", err.Error())
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Errorf("error trying to setup websocket connection: %q", err.Error())
			return
		}

		logger.Debug("Creating new websocket client")
		client := NewWSClient(conn, server, clientConfig, logger)
		subscription := &Subscription{
			Client: client,
			RoomID: uint(roomID),
		}

		go subscription.writeMessages()
		go subscription.readMessages()

		server.register <- subscription
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
