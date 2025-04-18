package user

import (
	"database/sql/driver"
	"errors"
	"fmt"
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
	Timezone   Timezone
	// Fields for registered users (pointers allow null values)
	Username *string `gorm:"uniqueIndex"`
	Email    *string `gorm:"uniqueIndex"`
	Password *[]byte `gorm:"size:60"` // bcrypt hash is 60 bytes]
}

func (u *User) GetID() string {
	return u.ID
}

// register associates an unregistered user with a username and password and updates the
func (u *User) isRegistered() bool {
	return u.registered
}

// Timezone is a Valuer/Scanner interface for gorm to store time.Location as a basic type
// https://gorm.io/docs/data_types.html#Custom-Data-Types
type Timezone struct {
	*time.Location
}

func (t *Timezone) Value() (driver.Value, error) {
	if t.Location == nil {
		return nil, nil
	}
	return t.Location.String(), nil
}

func (t *Timezone) Scan(value any) error {
	if value == nil {
		t.Location = nil
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("invalid timezone value")
	}
	loc, err := time.LoadLocation(str)
	if err != nil {
		return fmt.Errorf("failed to load location: %w", err)
	}
	t.Location = loc
	return nil
}
