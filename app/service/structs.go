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
