package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/msanatan/go-chatroom/app/models"
	log "github.com/sirupsen/logrus"
)

// Client has all the handlers for user related activities
type Client struct {
	chatroomDB *models.ChatroomDB
	jwtSecret  string
	logger     *log.Entry
}

// NewClient instantiates a new auth client
func NewClient(chatroomDB *models.ChatroomDB, jwtSecret string, logger *log.Entry) *Client {
	return &Client{
		chatroomDB: chatroomDB,
		jwtSecret:  jwtSecret,
		logger:     logger,
	}
}

// WriteErrorResponse is a helper function that returns JSON response for errors
func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encodeError := json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
	if encodeError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Something bad happened, contact the system admin"}`))
	}
}

// CreateUser is a handler that creates a new user
func (c *Client) CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := c.logger.WithField("method", "CreateUser")
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logger.Errorf("could not unmarshal request body: %s", err.Error())
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user.Init()
	err = user.Validate("create")
	if err != nil {
		logger.Errorf("user is not valid: %s", err.Error())
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	createdUser := c.chatroomDB.DB.Create(user)
	if createdUser.Error != nil {
		logger.Errorf("user is not valid: %s", createdUser.Error.Error())
		WriteErrorResponse(w, http.StatusBadRequest,
			errors.New("could not create user at this time, please review your details and try again"))
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
func (c *Client) Login(w http.ResponseWriter, r *http.Request) {
	logger := c.logger.WithField("method", "Login")
	var loginRequest LoginPayload
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		logger.Errorf("could not unmarshal request body: %s", err.Error())
		WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var user models.User
	tx := c.chatroomDB.DB.Where("username = ?", loginRequest.Username).First(&user)
	if tx.Error != nil {
		logger.Errorf("could not find user: %s", err.Error())
		WriteErrorResponse(w, http.StatusNotFound, errors.New("no user found with username: "+loginRequest.Username))
		return
	}

	if loginRequest.Password == user.Password {
		logger.Debug("login successful, returning token")
		token, err := GenerateJWT(user.ID, c.jwtSecret, 3600)
		if err != nil {
			logger.Errorf("could not generate a JWT: %s", err.Error())
			WriteErrorResponse(w, http.StatusInternalServerError,
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
func (c *Client) IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := c.logger.WithField("method", "IsAuthenticated")
		token := GetTokenFromRequest(r)
		if token == "" {
			logger.Error("no token found in request")
			WriteErrorResponse(w, http.StatusUnauthorized, errors.New("you need to be authenticated first"))
			return
		}

		userID, err := VerifyJWT(token, c.jwtSecret)
		if err != nil {
			logger.Errorf("could not verify JWT: %s", err.Error())
			WriteErrorResponse(w, http.StatusForbidden, err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userId", userID))
		next.ServeHTTP(w, r)
	})
}
