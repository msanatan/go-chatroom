package rabbitmq

// BotMessagePayload is the data envelopment published to bots
type BotMessagePayload struct {
	Command  string `json:"command"`
	Argument string `json:"argument"`
	RoomID   uint   `json:"roomId"`
}
