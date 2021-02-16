package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
)

// Message saves a message sent from the client
type Message struct {
	gorm.Model
	ID     string `gorm:"primary_key"`
	Text   string `gorm:"not null;" json:"text"`
	Type   string `gorm:"not null;" json:"type"`
	UserID string
	User   *User
}

// Init prepares a message object to be saved
func (m *Message) Init() {
	m.ID = ksuid.New().String()
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
}

// Validate checks if a message model is correctly formed
func (m *Message) Validate(context string) error {
	if m.Text == "" {
		return errors.New("message text is missing")
	}

	if m.Type == "" {
		return errors.New("message type is missing")
	}

	return nil
}
