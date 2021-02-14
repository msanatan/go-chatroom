package service

// MessagePayload is the envelope for messages sent to and from the chat participants
type MessagePayload struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}
