package websockets

import (
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Server is our hub for all WS clients
type Server struct {
	clients    map[*Client]bool
	register   chan *Client
	deregister chan *Client
	logger     *log.Entry
}

// NewServer instantiates a new server struct
func NewServer(logger *log.Entry) *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		deregister: make(chan *Client),
		logger:     logger,
	}
}

func (server *Server) registerClient(client *Client) {
	server.clients[client] = true
}

func (server *Server) deregisterClient(client *Client) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
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
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// ServeWs runs a websocket client
func ServeWs(server *Server, logger *log.Entry) http.HandlerFunc {
	logger = logger.WithField("method", "ServeWs")
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Errorf("error trying to setup websocket connection: %q", err.Error())
			return
		}

		logger.Debug("Creating new websocket client")
		client := NewClient(conn, server, logger, "main")
		server.register <- client
	}
}
