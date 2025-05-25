package user

import (
	"errors"
	"fmt"
	"job/internal/scheduler"

	"gorm.io/gorm"
)

// TODO: figure out the business logic for updating password/changing email; should require verification

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

// CreateNewUser creates a new user with the given timezone and returns it. Nil input defaults to the default timezone.
func (s *Service) CreateNewUser(t *scheduler.Timezone) (*User, error) {
	u, err := s.repo.CreateUser(&User{Timezone: t})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// should this refactor to a updateUserFields function which is called in the handler layer?
func (s *Service) RegisterBaseUser(id string, uf *UserUpdateFields) (*User, error) {
	u, err := s.repo.LookupUser(id)
	if err != nil {
		return nil, err
	}
	badFields := map[string]string{}
	if u.IsRegistered() {
		badFields[registeredField] = "user already registered"
	}
	if uf.Username == nil {
		badFields[usernameField] = "missing username for registration request"
	}
	if uf.Email == nil {
		badFields[emailField] = "missing email for registration request"
	}
	if uf.Password == nil {
		badFields[passwordField] = "missing password for registration request"
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for id, v := range badFields {
			err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
		}
		return nil, err
	}
	t := true
	uf.Registered = &t
	return s.UpdateUserFields(id, uf)
}

func (s *Service) validateUpdateUserFields(u User) error {
	badFields := map[string]string{}
	if u.Username != nil {
		res := s.repo.db.Where("username = ?", *u.Username).First(&User{})
		if res.Error == nil {
			badFields[usernameField] = "username already taken"
		} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			badFields[usernameField] = res.Error.Error()
		}
	}
	if u.Email != nil {
		res := s.repo.db.Where("email = ?", *u.Email).First(&User{})
		if res.Error == nil {
			badFields[emailField] = "email already registered"
		} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			badFields[emailField] = res.Error.Error()
		}
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for id, v := range badFields {
			err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
		}
		return err
	}
	return nil
}

// UpdateUserFields updates the valid user fields specified in the fields map.
// TODO: rethink business logic for changing email/username and enable changing password
func (s *Service) UpdateUserFields(id string, fields *UserUpdateFields) (*User, error) {
	_, err := s.repo.LookupUser(id)
	if err != nil {
		return nil, err
	}
	// updateUser has trimmed and hashed fields, so I validate it below
	updateUser, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}
	if err := s.validateUpdateUserFields(*updateUser); err != nil {
		return nil, err
	}

	// i could check for differences here but i don't think it's that important
	updateUser.ID = id
	_, err = s.repo.UpdateUserFields(updateUser, updateFields)
	if err != nil {
		return nil, err
	}
	// it is best practice to return the updated user object
	return s.LookupUser(id)
}

func (s *Service) LookupUser(id string) (*User, error) {
	return s.repo.LookupUser(id)
}

func (s *Service) DeleteUser(id string) error {
	return s.repo.DeleteUser(&User{ID: id})
}

func (s *Service) SelectAll() ([]User, error) {
	return s.repo.selectAll()
}
