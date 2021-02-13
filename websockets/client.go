package websockets

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Client is the websocket client users will connect to
type Client struct {
	conn   *websocket.Conn
	server *Server
	logger *log.Entry
	room   string
}

// NewClient instantiates a new client
func NewClient(conn *websocket.Conn, server *Server, logger *log.Entry, room string) *Client {
	return &Client{
		conn:   conn,
		server: server,
		logger: logger,
		room:   room,
	}
}
