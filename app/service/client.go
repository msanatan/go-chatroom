package service

import (
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// ClientConfig contains configuration needed to communicate with the WS server
type ClientConfig struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

// WSClient is the websocket client users will connect to
type WSClient struct {
	conn   *websocket.Conn
	server *Server
	send   chan MessagePayload
	config *ClientConfig
	logger *log.Entry
}

// Subscription is a struct to encapsulates a client connection
// and the room it's connecting to
type Subscription struct {
	Client *WSClient
	RoomID uint
}

// NewWSClient instantiates a new websocket client
func NewWSClient(conn *websocket.Conn, server *Server, config *ClientConfig, logger *log.Entry) *WSClient {
	return &WSClient{
		conn:   conn,
		server: server,
		config: config,
		send:   make(chan MessagePayload),
		logger: logger,
	}
}

func (s *Subscription) disconnect() {
	logger := s.Client.logger.WithField("method", "disconnect")
	s.Client.server.Deregister <- s
	close(s.Client.send)
	s.Client.conn.Close()
	logger.Debug("disconnecting client")
}

func (s *Subscription) readMessages() {
	logger := s.Client.logger.WithField("method", "readMessages")
	defer func() {
		s.disconnect()
	}()

	s.Client.conn.SetReadLimit(s.Client.config.MaxMessageSize)
	s.Client.conn.SetReadDeadline(time.Now().Add(s.Client.config.PongWait))
	s.Client.conn.SetPongHandler(func(string) error {
		s.Client.conn.SetReadDeadline(time.Now().Add(s.Client.config.PongWait))
		return nil
	})

	// Start endless read loop, waiting for messages from client
	for {
		var message MessagePayload
		err := s.Client.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("unexpected close error: %s", err.Error())
			}
			break
		}
	}
}

func (s *Subscription) writeMessages() {
	logger := s.Client.logger.WithField("method", "writeMessages")
	ticker := time.NewTicker(s.Client.config.PingPeriod)
	defer func() {
		ticker.Stop()
		s.Client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-s.Client.send:
			s.Client.conn.SetWriteDeadline(time.Now().Add(s.Client.config.WriteWait))
			if !ok {
				// The WsServer closed the channel.
				s.Client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := s.Client.conn.WriteJSON(message)
			if err != nil {
				logger.Errorf("error sending message: %s", err.Error())
				return
			}
		case <-ticker.C:
			s.Client.conn.SetWriteDeadline(time.Now().Add(s.Client.config.WriteWait))
			if err := s.Client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Errorf("unable to send ping: %s", err.Error())
				return
			}
		}
	}
}
