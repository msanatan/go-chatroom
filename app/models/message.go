package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// Message saves a message sent from the client
type Message struct {
	gorm.Model
	Text   string `gorm:"not null;" json:"text"`
	Type   string `gorm:"not null;" json:"type"`
	UserID uint
	User   *User
	RoomID uint
	Room   *Room
}

// Init prepares a message object to be saved
func (m *Message) Init() {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
}

// Validate checks if a message model is correctly formed
func (m *Message) Validate() error {
	if m.Text == "" {
		return errors.New("message text is missing")
	}

	if m.Type == "" {
		return errors.New("message type is missing")
	}

	if m.UserID == 0 {
		return errors.New("user ID is missing")
	}

	if m.RoomID == 0 {
		return errors.New("room ID is missing")
	}

	return nil
}
