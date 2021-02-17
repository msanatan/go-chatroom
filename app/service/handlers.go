package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/msanatan/go-chatroom/app/auth"
	"github.com/msanatan/go-chatroom/app/models"
	"github.com/msanatan/go-chatroom/rabbitmq"
	"github.com/msanatan/go-chatroom/utils"
)

// GetLastMessages pulls the latest messages from the DB
// Typically used before a client connects to populate the chat
func (s *Server) GetLastMessages(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("method", "GetLastMessages")

	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["roomId"])
	if err != nil {
		logger.Errorf("room ID is not valid: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest,
			errors.New("you can only get messages from a valid room ID"))
		return
	}

	var messages []models.Message
	tx := s.chatroomDB.DB.Where("room_id = ?", roomID).Order("created_at asc").Limit(50).Preload("User").Find(&messages)
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
			RoomID:   message.RoomID,
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
	message.UserID = r.Context().Value("userId").(uint)
	message.RoomID = newMessage.RoomID
	message.Init()

	err = message.Validate()
	if err != nil {
		logger.Errorf("message is not valid: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	tx := s.chatroomDB.DB.Create(message)
	if tx.Error != nil {
		logger.Errorf("failed to create message: %s", tx.Error.Error())
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

// CreateRoom is a handler that creates a new room
func (s *Server) CreateRoom(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("method", "CreateRoom")
	var room models.Room
	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		logger.Errorf("could not unmarshal request body: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	room.Init()
	err = room.Validate()
	if err != nil {
		logger.Errorf("room is not valid: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	createdRoom := s.chatroomDB.DB.Create(room)
	if createdRoom.Error != nil {
		logger.Errorf("failed to create room: %s", createdRoom.Error.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest,
			errors.New("could not create a room at this time, please review your details and try again"))
		return
	}

	responsePayload := RoomPayload{
		ID:   room.ID,
		Name: room.Name,
	}

	resp, _ := json.Marshal(&responsePayload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// GetRooms returns a list of all rooms
func (s *Server) GetRooms(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("method", "GetRooms")

	var rooms []models.Room
	tx := s.chatroomDB.DB.Order("created_at asc").Find(&rooms)
	if tx.Error != nil {
		logger.Errorf("could not pull list of rooms: %s", tx.Error.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest,
			errors.New("could not pull the list of rooms"))
		return
	}

	var roomsPayload []RoomPayload
	for _, room := range rooms {
		roomsPayload = append(roomsPayload, RoomPayload{
			ID:   room.ID,
			Name: room.Name,
		})
	}

	responsePayload := RoomsPayload{
		Rooms: roomsPayload,
		Size:  len(roomsPayload),
	}

	resp, _ := json.Marshal(&responsePayload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// CreateUser is a handler that creates a new user
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("method", "CreateUser")
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logger.Errorf("could not unmarshal request body: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user.Init()
	err = user.Validate("create")
	if err != nil {
		logger.Errorf("user is not valid: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	createdUser := s.chatroomDB.DB.Create(user)
	if createdUser.Error != nil {
		logger.Errorf("failed to create user: %s", createdUser.Error.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest,
			errors.New("could not create a user at this time, please review your details and try again"))
		return
	}

	responsePayload := CreateUserResponse{
		Username: user.Username,
		Email:    user.Email,
	}

	resp, _ := json.Marshal(&responsePayload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// Login validates a user, returning a JWT if login was successful
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("method", "Login")
	var loginRequest LoginPayload
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		logger.Errorf("could not unmarshal request body: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var user models.User
	tx := s.chatroomDB.DB.Where("username = ?", loginRequest.Username).First(&user)
	if tx.Error != nil {
		logger.Errorf("could not find user: %s", err.Error())
		utils.WriteErrorResponse(w, http.StatusNotFound, errors.New("no user found with username: "+loginRequest.Username))
		return
	}

	if loginRequest.Password == user.Password {
		logger.Debug("login successful, returning token")
		token, err := auth.GenerateJWT(int(user.ID), user.Username, s.jwtSecret, 3600)
		if err != nil {
			logger.Errorf("could not generate a JWT: %s", err.Error())
			utils.WriteErrorResponse(w, http.StatusInternalServerError,
				errors.New("we're experiencing difficulty completing your login request, please try again at a later time"))
			return
		}

		responsePayload := LoginResponse{
			Token: token,
		}

		resp, _ := json.Marshal(&responsePayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// IsAuthenticated is a middleware that checks if a requester is authenticated or not
func (s *Server) IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithField("method", "IsAuthenticated")
		token := auth.GetTokenFromRequest(r)
		if token == "" {
			logger.Error("no token found in request")
			utils.WriteErrorResponse(w, http.StatusUnauthorized, errors.New("you need to be authenticated first"))
			return
		}

		userID, username, err := auth.VerifyJWT(token, s.jwtSecret)
		if err != nil {
			logger.Errorf("could not verify JWT: %s", err.Error())
			utils.WriteErrorResponse(w, http.StatusForbidden, err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userId", userID))
		r = r.WithContext(context.WithValue(r.Context(), "username", username))
		next.ServeHTTP(w, r)
	})
}
