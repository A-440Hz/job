package user

import (
	"job/internal/db"
	"job/internal/scheduler"
	"strings"
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
	ID         string `gorm:"primaryKey"`
	Registered bool   `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt      `gorm:"index"`
	Timezone   *scheduler.Timezone ``
	// Fields for registered users (pointers allow null values)
	Username *string `gorm:"uniqueIndex" json:"username,omitempty"`
	Email    *string `gorm:"uniqueIndex" json:"email,omitempty"`
	Password *[]byte `gorm:"size:60" json:"-"` // bcrypt hash is 60 bytes
}

func (u *User) GetID() string {
	return u.ID
}

// register associates an unregistered user with a username and password and updates the
func (u *User) IsRegistered() bool {
	return u.Registered
}

type UserUpdateFields struct {
	Registered *bool   `json:"registered,omitempty"`
	Username   *string `json:"username,omitempty"`
	Email      *string `json:"email,omitempty"`
	Password   *string `json:"password,omitempty"`
	Timezone   *int64  `json:"timezone,omitempty"`
}

func (uf *UserUpdateFields) sanitizeFields() {
	if uf.Username != nil {
		// trim username
		uname := strings.TrimSpace(*uf.Username)
		uf.Username = &uname
	}
	if uf.Email != nil {
		// trim and lowercase email
		umail := strings.ToLower(strings.TrimSpace(*uf.Email))
		uf.Email = &umail
	}
}

func (uf *UserUpdateFields) formatForRepo() (*User, []string, error) {
	u := &User{}
	fields := []string{}
	uf.sanitizeFields()
	if uf.Registered != nil {
		u.Registered = *uf.Registered
		fields = append(fields, registeredField)
	}
	if uf.Username != nil {
		// trim username
		uname := strings.TrimSpace(*uf.Username)
		u.Username = &uname
		fields = append(fields, usernameField)
	}
	if uf.Email != nil {
		// trim and lowercase email
		umail := strings.ToLower(strings.TrimSpace(*uf.Email))
		u.Email = &umail
		fields = append(fields, emailField)
	}
	if uf.Password != nil {
		// hash password
		p, err := db.HashPassword(*uf.Password)
		if err != nil {
			return nil, nil, err
		}
		u.Password = &p
		fields = append(fields, passwordField)
	}
	if uf.Timezone != nil {
		u.Timezone = scheduler.NewTimezoneWithOffset(*uf.Timezone)
		fields = append(fields, timezoneField)
	}

	return u, fields, nil
}
