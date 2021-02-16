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
