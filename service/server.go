package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Server is our hub for all WS clients
type Server struct {
	clients    map[*WSClient]bool
	register   chan *WSClient
	deregister chan *WSClient
	broadcast  chan MessagePayload
	bots       map[string]Bot
	botSymbol  string
	logger     *log.Entry
}

// NewServer instantiates a new server struct
func NewServer(bots map[string]Bot, botSymbol string, logger *log.Entry) *Server {
	if bots == nil {
		bots = make(map[string]Bot)
	}

	if botSymbol == "" {
		botSymbol = "/"
	}

	return &Server{
		clients:    make(map[*WSClient]bool),
		register:   make(chan *WSClient),
		deregister: make(chan *WSClient),
		broadcast:  make(chan MessagePayload),
		bots:       bots,
		botSymbol:  botSymbol,
		logger:     logger,
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
	logger := server.logger.WithField("method", "Run")
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.deregister:
			server.deregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)

			// Check if message should be handled by a bot
			if server.IsValidBotCommand(message.Message) {
				botCommand, argument := server.ExtractCommandAndArgs(message.Message)
				if bot, ok := server.bots[botCommand]; ok {
					response, err := bot.ProcessCommand(argument)
					if err != nil {
						logger.Errorf("error processing bot command: %s", err.Error())
						server.broadcastToClients(MessagePayload{
							Message: err.Error(),
							Type:    "error",
						})
					} else {
						logger.Debugf("%s bot succesffuly responding to command", botCommand)
						server.broadcastToClients(MessagePayload{
							Message: response,
							Type:    "botMessage",
						})
					}
				} else {
					server.broadcastToClients(MessagePayload{
						Message: fmt.Sprintf("%s is not a recognized bot command", botCommand),
						Type:    "error",
					})
				}
			}
		}
	}
}

// IsValidBotCommand verifies if a message should be treated as a bot command
func (server *Server) IsValidBotCommand(message string) bool {
	return len(message) > 0 &&
		strings.HasPrefix(message, server.botSymbol) &&
		!strings.HasPrefix(message, server.botSymbol+server.botSymbol)
}

// ExtractCommandAndArgs parses a bot command and any arguments it may have
func (server *Server) ExtractCommandAndArgs(message string) (string, string) {
	if strings.Contains(message, "=") {
		commandString := message[strings.Index(message, server.botSymbol)+1 : strings.Index(message, "=")]
		args := strings.SplitN(message, "=", 2)[1]
		return commandString, args
	}

	return strings.SplitN(message, server.botSymbol, 2)[1], ""
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
