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
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = db.NewPublicID(db.UserIdPrefix)
	if u.Timezone == nil {
		u.Timezone = scheduler.GetDefaultTimezone()
	}
	return nil
}

func (r *Repository) CreateUser(u *User) (*User, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) LookupUser(id string) (*User, error) {
	u := &User{ID: id}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	// opportunity to return gorm.ErrRecordNotFound differently
	return u, nil
}

func (r *Repository) UpdateUserFields(u *User, fields []string) (*User, error) {
	// https://www.codingexplorations.com/blog/gorm-save-vs-update-explained
	// db.Save updates all fields, which is okay for user registration, but not efficient for operations like changing username or password individually
	res := r.db.Model(u).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

// as is this does a soft delete https://gorm.io/docs/delete.html#Soft-Delete
func (r *Repository) DeleteUser(u *User) error {
	// no lookups here because i dont need to log the user struct
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) selectAll() ([]User, error) {
	users := []User{}
	res := r.db.Find(&users)
	if res.Error != nil {
		return nil, res.Error
	}
	return users, nil
}
