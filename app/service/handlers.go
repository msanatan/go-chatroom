package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/msanatan/go-chatroom/app/models"
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
