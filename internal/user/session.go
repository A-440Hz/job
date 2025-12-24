package user

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

var tokenByteSize = 20

type Session struct {
	ID        string `gorm:"primaryKey"`
	UserID    string `gorm:"foreignKey"`
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func generateSessionID() (string, error) {
	b := make([]byte, tokenByteSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
