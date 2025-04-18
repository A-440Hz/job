package user

import (
	"errors"
	"job/internal/db"
	"strings"
	"time"

	"gorm.io/gorm"
)

// business logic goes here
type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateNewUser(l *time.Location) (*User, error) {
	u := &User{Timezone: Timezone{Location: l}}
	u, err := s.repo.CreateUser(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) validateRegisterBaseUser(username string, email string) error {
	badFields := map[string]string{}
	res := s.repo.db.Where("username = ?", username).First(&User{})
	if res.Error == nil {
		badFields["username"] = "username already registered"
	} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		badFields["username_query"] = res.Error.Error()
	}
	res = s.repo.db.Where("email = ?", email).First(&User{})
	if res.Error == nil {
		badFields["email"] = "email already registered"
	} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		badFields["email_query"] = res.Error.Error()
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
	u, err := s.repo.LookupUser(id)
	if err != nil {
		return err
	}

	if u.isRegistered() {
		return errors.New("current user id already registered")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if err = s.validateRegisterBaseUser(username, email); err != nil {
		return err
	}
	passwordHash, err := db.HashPassword(password)
	if err != nil {
		return err
	}

	u.Username = &username
	u.Email = &email
	u.Password = &passwordHash
	u.registered = true

	s.repo.UpdateUser(u)
	return nil
}

// needs a path to change username/email/password after entering the correct password
// not sure how to lay out the function if it's a multiple step process
func (s *Service) updateUser(u *User) (*User, error) {
	u, err := s.repo.UpdateUser(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) LookupUser(id string) (*User, error) {
	u, err := s.repo.LookupUser(id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) DeleteUser(id string) error {
	u := &User{ID: id}
	if err := s.repo.DeleteUser(u); err != nil {
		return err
	}
	return nil
}
