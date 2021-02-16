package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/msanatan/go-chatroom/app/models"
	"github.com/msanatan/go-chatroom/rabbitmq"
	"github.com/msanatan/go-chatroom/utils"
)

// GetLastMessages pulls the latest messages from the DB
// Typically used before a client connects to populate the chat
func (s *Server) GetLastMessages(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("method", "GetLastMessages")

	var messages []models.Message
	tx := s.chatroomDB.DB.Order("created_at desc").Limit(50).Preload("User").Find(&messages)
	if tx.Error != nil {
		logger.Errorf("could not pull latest messages: %s", tx.Error.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest,
			errors.New("could not pull the most recent messages from this chat"))
		return
	}

	var messagesPayload []MessagePayload
	for _, message := range messages {
		messagesPayload = append(messagesPayload, MessagePayload{
			Message:  message.Text,
			Type:     message.Type,
			Username: message.User.Username,
			Created:  message.CreatedAt.Format(time.RFC1123Z),
		})
	}

	responsePayload := MessagesPayload{
		Messages: messagesPayload,
		Size:     len(messagesPayload),
	}

	resp, _ := json.Marshal(&responsePayload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// CreateMessage adds a new message to the DB and websocket server
// It also publishes to RabbitMQ in case it's a bot
func (s *Server) CreateMessage(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("method", "CreateMessage")

	var newMessage MessagePayload
	err := json.NewDecoder(r.Body).Decode(&newMessage)
	if err != nil {
		logger.Errorf("could not unmarshal request body: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var message models.Message
	message.Text = newMessage.Message
	message.Type = newMessage.Type
	message.UserID = r.Context().Value("userId").(string)
	message.Init()

	err = message.Validate()
	if err != nil {
		logger.Errorf("message is not valid: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	tx := s.chatroomDB.DB.Create(message)
	if tx.Error != nil {
		logger.Errorf("message is not valid: %s", tx.Error.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest,
			errors.New("could not create a message at this time, please review your request and try again"))
		return
	}

	responsePayload := MessagePayload{
		Message:  message.Text,
		Type:     message.Type,
		Username: r.Context().Value("username").(string),
		Created:  message.CreatedAt.Format(time.RFC1123Z),
	}

	// Now that message is in DB, let's write it to the websocket server
	s.broadcast <- responsePayload

	// Check if message should be handled by a bot
	if s.IsValidBotCommand(responsePayload.Message) {
		if s.rabbitMQClient != nil {
			botCommand, argument := s.ExtractCommandAndArgs(responsePayload.Message)
			botPayload := rabbitmq.BotMessagePayload{
				Command:  botCommand,
				Argument: argument,
			}

			botPayloadJSON, err := json.Marshal(botPayload)
			if err != nil {
				logger.Errorf("strangely enough, could not convert the bot error response to JSON: %s", err.Error())
				s.broadcast <- MessagePayload{
					Message: "Could not send a valid request to the bot. Please review your command",
					Type:    "error",
				}
			}

			s.rabbitMQClient.Publish(botPayloadJSON)
		} else {
			s.broadcast <- MessagePayload{
				Message: "This chatroom isn't configured to work with bots",
				Type:    "error",
			}
		}
	}

	resp, _ := json.Marshal(&responsePayload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}
