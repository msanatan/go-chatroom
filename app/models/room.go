package models

import (
	"errors"
	"time"

	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

// Room represents one dedicated channel to chat in
type Room struct {
	gorm.Model
	ID   string `gorm:"primary_key"`
	Name string `gorm:"not null;" json:"name"`
}

// Init prepares a room object to be saved
func (r *Room) Init() {
	r.ID = ksuid.New().String()
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
