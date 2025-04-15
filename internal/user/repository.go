package user

import (
	"job/internal/db"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(*User) error
	GetByID(uint) (*User, error)
	GetByPublicID(string) (*User, error)
	Update(*User) error
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(d *gorm.DB) *Repository {
	return &Repository{db: d}
}

// this is a gorm hook function https://gorm.io/docs/create.html#Create-Hooks
// do i do it like this or have a default tag?
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = db.NewPublicID(db.UserIdPrefix)
	return nil
}

func (r *Repository) CreateBaseUser(u *User) (*User, error) {
	// maybe some validation here
	r.db.Create(u)
	if r.db.Error != nil {
		return nil, r.db.Error
	}
	return u, nil
}

func (r *Repository) LookupBaseUser(id string) (*User, error) {
	u := &User{ID: id}
	if err := r.db.First(u); err.Error != nil {
		return nil, err.Error
	}
	return u, nil
}

func (r *Repository) UpdateBaseUser(u *User) (*User, error) {
	// needs validation here? dunno
	r.db.Save(u)
	return u, nil
}

// as is this does a soft delete https://gorm.io/docs/delete.html#Soft-Delete
func (r *Repository) DeleteBaseUser(u *User) error {
	r.db.Delete(u)
	return nil
}
