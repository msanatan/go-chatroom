package models

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ChatroomDB allows the app to neatly interface with GORM
type ChatroomDB struct {
	DB     *gorm.DB
	logger *log.Entry
}

// NewChatroomDB instantiates a new ChatroomDB object
func NewChatroomDB(db *sql.DB, logger *log.Entry) (*ChatroomDB, error) {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), nil)

	if err != nil {
		return nil, err
	}

	return &ChatroomDB{
		DB:     gormDB,
		logger: logger,
	}, nil
}

// Migrate runs the migrations on the GORM models
func (c *ChatroomDB) Migrate() error {
	err := c.DB.AutoMigrate(&Room{})
	if err != nil {
		return err
	}

	err = c.DB.AutoMigrate(&Message{})
	if err != nil {
		return err
	}

	err = c.DB.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	return nil
}
