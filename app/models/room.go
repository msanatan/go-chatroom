package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Room represents one dedicated channel to chat in
type Room struct {
	gorm.Model
	Name     string `gorm:"not null;" json:"name"`
	Messages []Message
}

// Init prepares a room object to be saved
func (r *Room) Init() {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
}

// Validate checks if a message model is correctly formed
func (r *Room) Validate() error {
	if r.Name == "" {
		return errors.New("room name is missing")
	}

	return nil
}
