package service

// MessagePayload is the envelope for messages sent to and from the chat participants
type MessagePayload struct {
	Message  string `json:"message"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Created  string `json:"created"`
}

// MessagesPayload wrapper around list of messages
// This envelope is useful for future metadata
type MessagesPayload struct {
	Messages []MessagePayload `json:"messages"`
	Size     int              `json:"size"`
}

// CreateUserResponse is the payload for a successful created user
// We don't want to send password details in the response
type CreateUserResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// LoginPayload is the payload for a login request
type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the payload for a successful login response
type LoginResponse struct {
	Token string `json:"token"`
}
