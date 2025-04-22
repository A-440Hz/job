package user

import (
	"errors"
	"job/internal/db"
	"job/internal/scheduler"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// business logic goes here
type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

// CreateNewUser creates a new user with the given timezone and returns it. Nil input defaults to the default timezone.
func (s *Service) CreateNewUser(t *scheduler.Timezone) (*User, error) {
	u := &User{Timezone: t}
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

func (s *Service) RegisterBaseUser(id string, username string, password string, email string) (*User, error) {
	u, err := s.repo.LookupUser(id)
	if err != nil {
		return nil, err
	}

	if u.IsRegistered() {
		return nil, errors.New("current user id already registered")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if err = s.validateRegisterBaseUser(username, email); err != nil {
		return nil, err
	}
	passwordHash, err := db.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u.Username = &username
	u.Email = &email
	u.Password = &passwordHash
	u.Registered = true

	fields := []string{usernameField, emailField, passwordField, registeredField}

	u, err = s.repo.UpdateUserFields(u, fields)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUserFields updates the valid user fields specified in the fields map.
func (s *Service) UpdateUserFields(id string, fields map[string]string) (*User, error) {
	u, err := s.repo.LookupUser(id)
	if err != nil {
		return nil, err
	}
	updateFields := []string{}
	for k, v := range fields {
		switch {
		case usernameField == k && u.Username == nil || *u.Username != v:
			u.Username = &v
			updateFields = append(updateFields, usernameField)
		case emailField == k && u.Email == nil || *u.Email != v:
			v = strings.ToLower(strings.TrimSpace(v))
			u.Email = &v
			updateFields = append(updateFields, emailField)
		case passwordField == k && u.Password == nil || db.CheckPasswordHash(v, *u.Password) == false:
			hash, err := db.HashPassword(v)
			if err != nil {
				return nil, err
			}
			u.Password = &hash
			updateFields = append(updateFields, passwordField)
		case registeredField == k && u.Registered == false:
			u.Registered = true
			updateFields = append(updateFields, registeredField)
		case timezoneField == k:
			v, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			tz := scheduler.NewTimezone(v)
			if u.Timezone == nil || u.Timezone != tz {
				u.Timezone = tz
				updateFields = append(updateFields, timezoneField)
			}
		}
	}
	u, err = s.repo.UpdateUserFields(u, updateFields)
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
