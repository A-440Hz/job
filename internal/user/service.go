package user

import (
	"errors"
	"job/internal/scheduler"

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

// should this refactor to a updateUserFields function which is called in the handler layer?
func (s *Service) RegisterBaseUser(id string, uf *UserUpdateFields) (*User, error) {
	badFields := map[string]string{}
	u, err := s.repo.LookupUser(id)
	if err != nil {
		badFields["lookup"] = err.Error()
	}
	if u.IsRegistered() {
		badFields["registered"] = "user already registered"
	}
	if uf.Username == nil {
		badFields["username"] = "missing username for registration request"
	}
	if uf.Email == nil {
		badFields["email"] = "missing email for registration request"
	}
	if uf.Password == nil {
		badFields["password"] = "missing password for registration request"
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for _, v := range badFields {
			err = errors.Join(err, errors.New(v))
		}
		return nil, err
	}
	u, err = s.UpdateUserFields(id, uf)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) validateUpdateUserFields(u User) error {
	badFields := map[string]string{}
	if u.Username != nil {
		res := s.repo.db.Where("username = ?", *u.Username).First(&User{})
		if res.Error == nil {
			badFields["username"] = "username already taken"
		} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			badFields["username_query"] = res.Error.Error()
		}
	}
	if u.Email != nil {
		res := s.repo.db.Where("email = ?", *u.Email).First(&User{})
		if res.Error == nil {
			badFields["email"] = "email already registered"
		} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			badFields["email_query"] = res.Error.Error()
		}
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

// UpdateUserFields updates the valid user fields specified in the fields map.
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
	updateUser.ID = id
	_, err = s.repo.UpdateUserFields(updateUser, updateFields)
	if err != nil {
		return nil, err
	}
	// it is best practice to return the updated user object
	return s.LookupUser(id)
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
