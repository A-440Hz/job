package user

type User interface {
	getUserId() string
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
