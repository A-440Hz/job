package user

import "gorm.io/gorm"

// baseUser exists so I can allow users to plug and play without having to register an account
// this design assumes I will have to clean up expired user entries in the database
// and I will be able to transition UnregisteredUsers into RegisteredUsers
type baseUser interface {
	getUserId() string
	isRegistered() bool
}

type BUser struct {
	gorm.Model
	UserId     string `gorm:"uniqueIndex"` // DB will use gorm ID for PK. this is good for logs a
	registered bool   `gorm:"default:false"`
}

// I assume the server has some secretToken field for encryption
type UserCredentials struct {
	Email    string
	Password string
}

type RegisteredUser struct {
	BUser
	UserCredentials
	Username string
}

func (u *BUser) getUserId() string {
	return u.UserId
}

// register associates an unregistered user with a username and password and updates the
func (u *BUser) isRegistered() bool {
	return u.registered
}
