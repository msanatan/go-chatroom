package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/msanatan/go-chatroom/app/rabbitmq"
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
	room   string
}

// NewWSClient instantiates a new websocket client
func NewWSClient(conn *websocket.Conn, server *Server, config *ClientConfig, logger *log.Entry, room string) *WSClient {
	return &WSClient{
		conn:   conn,
		server: server,
		config: config,
		send:   make(chan MessagePayload),
		logger: logger,
		room:   room,
	}
}

func (client *WSClient) disconnect() {
	logger := client.logger.WithField("method", "disconnect")
	client.server.deregister <- client
	close(client.send)
	client.conn.Close()
	logger.Debug("disconnecting client")
}

func (client *WSClient) readMessages() {
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

		// Check if message should be handled by a bot
		if client.IsValidBotCommand(message.Message) {
			if client.server.rabbitMQClient != nil {
				botCommand, argument := client.ExtractCommandAndArgs(message.Message)
				botPayload := rabbitmq.BotMessagePayload{
					Command:  botCommand,
					Argument: argument,
				}

				botPayloadJSON, err := json.Marshal(botPayload)
				if err != nil {
					logger.Errorf("strangely enough, could not convert the bot error response to JSON: %s", err.Error())
					client.server.broadcast <- MessagePayload{
						Message: "Could not send a valid request to the bot. Please review your command",
						Type:    "error",
					}
					continue
				}

				client.server.rabbitMQClient.Publish(botPayloadJSON)
			} else {
				client.server.broadcast <- MessagePayload{
					Message: "This chatroom isn't configured to work with bots",
					Type:    "error",
				}
			}
		}
	}

}

func (client *WSClient) writeMessages() {
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

// IsValidBotCommand verifies if a message should be treated as a bot command
func (client *WSClient) IsValidBotCommand(message string) bool {
	return len(message) > 0 &&
		strings.HasPrefix(message, client.server.botSymbol) &&
		!strings.HasPrefix(message, client.server.botSymbol+client.server.botSymbol)
}

// ExtractCommandAndArgs parses a bot command and any arguments it may have
func (client *WSClient) ExtractCommandAndArgs(message string) (string, string) {
	if strings.Contains(message, "=") {
		commandString := message[strings.Index(message, client.server.botSymbol)+1 : strings.Index(message, "=")]
		args := strings.SplitN(message, "=", 2)[1]
		return commandString, args
	}

	return strings.SplitN(message, client.server.botSymbol, 2)[1], ""
}
