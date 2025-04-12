package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	// gorm.Model
	// DB will use gorm ID for PK. PulbicId is good for logs and urls and future friend reqs?
	ID         string `gorm:"primaryKey"`
	registered bool   `gorm:"default:false"` // i think this is redundant https://gorm.io/docs/create.html#Default-Values
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	// Fields for registered users (pointers allow null values)
	Username *string `gorm:"uniqueIndex"`
	Email    *string `gorm:"uniqueIndex"`
	Password *string
}

func (u *User) getID() string {
	return u.ID
}

// register associates an unregistered user with a username and password and updates the
func (u *User) isRegistered() bool {
	return u.registered
}
