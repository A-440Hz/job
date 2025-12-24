package db

import (
	"errors"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	maxRetries   = 3
	retryDelay   = 300 * time.Millisecond
	retryBackoff = 3.0
)

// WithRetry wraps a GORM operation with automatic retry logic for connection errors
// Use this for critical operations that must succeed despite transient connection issues
//
// Example:
//
//	var user User
//	err := db.WithRetry(func(db *gorm.DB) error {
//	    return db.First(&user).Error
//	})
func WithRetry(operation func() error) error {
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = operation()

		if err == nil {
			return nil
		}

		if !isRetryableError(err) || attempt == maxRetries {
			return err
		}

		delay := time.Duration(float64(retryDelay) * float64(attempt) * retryBackoff)
		log.Printf("DB operation failed (attempt %d/%d), retrying after %v: %v",
			attempt, maxRetries, delay, err)
		time.Sleep(delay)
	}
	return err
}

// PingDB checks if the database connection is alive and attempts to reconnect if needed
func PingDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return WithRetry(func() error {
		return sqlDB.Ping()
	})
}

// isRetryableError determines if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"no connection to the server",
		"server closed the connection",
		"driver: bad connection",
		"bad connection",
		"unexpected eof",
		"connection timed out",
		"invalid connection",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return errors.Is(err, gorm.ErrInvalidDB)
}
