package auth

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

// MessageResponse is the objects in the message payload
// that a user gets upon startup
type MessageResponse struct {
	Message  string `json:"message"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Created  string `json:"created"`
}

// MessagesResponse wrapper around list of messages
// This envelope is useful for future metadata
type MessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Size     int               `json:"size"`
}
