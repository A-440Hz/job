package user

import (
	"job/internal/db"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

// this is a gorm hook function https://gorm.io/docs/create.html#Create-Hooks
// do i do it like this or have a default tag?
func (u *BUser) BeforeCreate(tx *gorm.DB) error {
	u.UserId = db.NewPublicID(db.UserIdPrefix)
	return nil
}

func NewRepo(conn *gorm.DB) *Repository {
	return &Repository{db: conn}
}

func (r *Repository) Create(u *BUser)
func (r *Repository) Read(u *BUser)
func (r *Repository) Update(u *BUser)
func (r *Repository) Delete(u *BUser)
