package user

import "gorm.io/gorm"

// baseUser exists so I can allow users to plug and play without having to register an account
// this design assumes I will have to clean up expired user entries in the database
// and I will be able to transition UnregisteredUsers into RegisteredUsers
// type baseUser interface {
// 	getUserId() string
// 	isRegistered() bool
// }

type BaseUser struct {
	gorm.Model
	// DB will use gorm ID for PK. UserId is good for logs and urls and future friend reqs?
	UserId     string `gorm:"uniqueIndex"`
	registered bool   `gorm:"default:false"`
}

// I assume the server has some secretToken field for encryption
type UserCredentials struct {
	Email    string
	Password string
}

type RegisteredUser struct {
	BaseUser
	UserCredentials
	Username string
}

func (u *BaseUser) getUserId() string {
	return u.UserId
}

// register associates an unregistered user with a username and password and updates the
func (u *BaseUser) isRegistered() bool {
	return u.registered
}
