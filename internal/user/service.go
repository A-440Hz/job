package user

import (
	"errors"
	"strings"
)

// business logic goes here

type UserService interface {
	AssignNewUser()
}

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) AssignNewUser() error {
	s.repo.CreateBaseUser(&User{})
	return nil
}

func (s *Service) validteRegisterBaseUser(username string, email string) error {
	badFields := map[string]string{}
	res := s.repo.db.Where("username = ?", username).Find(&User{})
	if res.RowsAffected > 0 {
		badFields["username"] = "username already registered"
	}
	res = s.repo.db.Where("email = ?", email).Find(&User{})
	if res.RowsAffected > 0 {
		badFields["email"] = "email already registered"
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for _, v := range badFields {
			err = errors.Join(err, errors.New(v))
		}
		return err
	}
	return nil
}

func (s *Service) RegisterBaseUser(id string, username string, password string, email string) error {
	u, err := s.repo.LookupBaseUser(id)
	if err != nil {
		return err
	}

	if u.isRegistered() {
		return errors.New("current user id already registered")
	}
	// validate username unique
	// validate valid password??
	email = strings.ToLower(strings.TrimSpace(email))
	// validate email unique

	return nil
}
