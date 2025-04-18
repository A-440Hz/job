package user

import (
	"errors"
	"job/internal/db"
	"job/internal/scheduler"

	"gorm.io/gorm"
)

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
	if u.Timezone.Location == nil {
		u.Timezone = Timezone{scheduler.DefaultTimezone}
	}
	return nil
}

func (r *Repository) CreateUser(u *User) (*User, error) {
	// maybe some validation here
	r.db.Create(u)
	if r.db.Error != nil {
		return nil, r.db.Error
	}
	return u, nil
}

func (r *Repository) LookupUser(id string) (*User, error) {
	u := &User{ID: id}
	if err := r.db.First(u); err.Error != nil {
		return nil, err.Error
	}
	return u, nil
}

func (r *Repository) UpdateUser(u *User) (*User, error) {
	r.db.Save(u)
	if r.db.Error != nil {
		return nil, r.db.Error
	}
	return u, nil
}

// as is this does a soft delete https://gorm.io/docs/delete.html#Soft-Delete
func (r *Repository) DeleteUser(u *User) error {
	// no lookups here because i dont need to log the user struct
	result := r.db.Delete(u)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	if result.Error != nil {
		return result.Error
	}
	return nil
}
