package user

type baseUser interface {
	getUserId() string
}

type User interface {
	baseUser
}

type UnregisteredUser struct {
	UserId string
}

type RegisteredUser struct {
	UserId string
}

func (u *UnregisteredUser) getUserId() string {
	return u.UserId
}
