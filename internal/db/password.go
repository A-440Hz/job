package db

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var GenericLoginError = errors.New("username and password do not match")

func HashPassword(password string) ([]byte, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func PasswordMatchesHash(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}
