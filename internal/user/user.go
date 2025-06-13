package user

import (
	"job/internal/collection"
	"job/internal/db"
	"job/internal/scheduler"
	"strings"
	"time"

	"gorm.io/gorm"
)

// can look into having gorm autofill json tags and read struct tags with reflection
// (so i can replace this pattern later). as it is I just need to make sure everything is snake case
// of the struct attributes
const (
	usernameField   = "username"
	emailField      = "email"
	passwordField   = "password"
	registeredField = "registered"
	timezoneField   = "timezone"
)

type User struct {
	ID         string `gorm:"primaryKey"`
	Registered bool   `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	Timezone   *scheduler.Timezone
	// Fields for registered users (pointers allow null values)
	Username *string `gorm:"uniqueIndex" json:"username,omitempty"`
	Email    *string `gorm:"uniqueIndex" json:"email,omitempty"`
	Password *[]byte `gorm:"size:60" json:"-"` // bcrypt hash is 60 bytes

	// potentially add every tracker type as foreign keys
	// JobAppTracker tracker.JobAppTracker `gorm:"foreignKey:UserID"`
	Inventory *collection.UserInventory `gorm:"foreignKey:UserID;references:ID" json:"inventory,omitempty"`
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

// sanitizeFields trims leading and trailing whitespaces in the username and email, and lowercases the email
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

// formatForRepo extracts and sanitizes the fields from the json struct, and hashes the password
func (uf *UserUpdateFields) formatForRepo() (*User, []string, error) {
	u := &User{}
	fields := []string{}
	uf.sanitizeFields()
	if uf.Registered != nil {
		u.Registered = *uf.Registered
		fields = append(fields, registeredField)
	}
	if uf.Username != nil {
		u.Username = uf.Username
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
