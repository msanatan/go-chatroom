package service

import (
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var newline = []byte{'\n'}
var space = []byte{' '}

// ClientConfig contains configuration needed to communicate with the WS server
type ClientConfig struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

// Client is the websocket client users will connect to
type Client struct {
	conn   *websocket.Conn
	server *Server
	send   chan MessagePayload
	config *ClientConfig
	logger *log.Entry
	room   string
}

// NewClient instantiates a new client
func NewClient(conn *websocket.Conn, server *Server, config *ClientConfig, logger *log.Entry, room string) *Client {
	return &Client{
		conn:   conn,
		server: server,
		config: config,
		send:   make(chan MessagePayload),
		logger: logger,
		room:   room,
	}
}

func (client *Client) disconnect() {
	logger := client.logger.WithField("method", "disconnect")
	client.server.deregister <- client
	close(client.send)
	client.conn.Close()
	logger.Debug("disconnecting client")
}

func (client *Client) readMessages() {
	logger := client.logger.WithField("method", "readMessages")
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(client.config.MaxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(client.config.PongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(client.config.PongWait))
		return nil
	})

	// Start endless read loop, waiting for messages from client
	for {
		var message MessagePayload
		err := client.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("unexpected close error: %s", err.Error())
			}
			break
		}

		client.server.broadcast <- message
	}

}

func (client *Client) writeMessages() {
	logger := client.logger.WithField("method", "writeMessages")
	ticker := time.NewTicker(client.config.PingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(client.config.WriteWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := client.conn.WriteJSON(message)
			if err != nil {
				logger.Errorf("error sending message: %s", err.Error())
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(client.config.WriteWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Errorf("unable to send ping: %s", err.Error())
				return
			}
		}
	}
}
