package user

import (
	"job/internal/scheduler"
	"time"

	"gorm.io/gorm"
)

const (
	usernameField   = "Username"
	emailField      = "Email"
	passwordField   = "Password"
	registeredField = "Registered"
	timezoneField   = "Timezone"
)

type User struct {
	// gorm.Model
	// DB will use gorm ID for PK. PulbicId is good for logs and urls and future friend reqs?
	ID         string `gorm:"primaryKey"`
	Registered bool   `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	Timezone   *scheduler.Timezone
	// Fields for registered users (pointers allow null values)
	Username *string `gorm:"uniqueIndex"`
	Email    *string `gorm:"uniqueIndex"`
	Password *[]byte `gorm:"size:60"` // bcrypt hash is 60 bytes]
}

func (u *User) GetID() string {
	return u.ID
}

// register associates an unregistered user with a username and password and updates the
func (u *User) IsRegistered() bool {
	return u.Registered
}
