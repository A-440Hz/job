package user

import (
	"job/internal/db"

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
func (u *BaseUser) BeforeCreate(tx *gorm.DB) error {
	u.UserId = db.NewPublicID(db.UserIdPrefix)
	return nil
}

func NewRepo(conn *gorm.DB) *Repository {
	return &Repository{db: conn}
}

func (r *Repository) CreateBaseUser(u *BaseUser) {
	r.db.Create(u)
}
func (r *Repository) Read(u *BaseUser)
func (r *Repository) Update(u *BaseUser)
func (r *Repository) Delete(u *BaseUser)
