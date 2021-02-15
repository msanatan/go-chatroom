package models

import (
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

// User is an entity that can log in our system
type User struct {
	gorm.Model
	ID       string `gorm:"primary_key"`
	Username string `gorm:"not null;unique" json:"username"`
	Email    string `gorm:"not null;unique" json:"email"`
	Password string `gorm:"not null;" json:"password"`
}

// HashPassword encrypts a password so it can be stored safely
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// VerifyPassword decrypts and checks if a hashed password is the same as the given string
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// BeforeSave is a GORM hook that encrypts the password before saving it
func (u *User) BeforeSave() error {
	hashedPassword, err := HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// Init prepares a user object to be saved
func (u *User) Init() {
	u.ID = ksuid.New().String()
	u.Username = html.EscapeString(strings.TrimSpace(u.Username))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}
