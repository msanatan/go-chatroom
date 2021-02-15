package models

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ChatroomDB allows the app to neatly interface with GORM
type ChatroomDB struct {
	db     *gorm.DB
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
		db:     gormDB,
		logger: logger,
	}, nil
}

// Migrate runs the migrations on the GORM models
func (c *ChatroomDB) Migrate(db *gorm.DB) error {
	//run gorm migrations
	err := c.db.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	return nil
}
