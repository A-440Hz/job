package user

import (
	"errors"
	"job/internal/db"
	"job/internal/scheduler"
	"time"

	"gorm.io/gorm"
)

const ()

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

func (r *Repository) createUser(u *User) (*User, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) lookupUser(id string) (*User, error) {
	u := &User{ID: id}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	// opportunity to return gorm.ErrRecordNotFound differently
	return u, nil
}

func (r *Repository) lookupUserByUsername(uname *string) (*User, error) {
	u := &User{Username: uname}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) updateUserFields(u *User, fields []string) (*User, error) {
	// https://www.codingexplorations.com/blog/gorm-save-vs-update-explained
	// db.Save updates all fields, which is okay for user registration, but not efficient for operations like changing username or password individually
	res := r.db.Model(u).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

// as is this does a soft delete https://gorm.io/docs/delete.html#Soft-Delete
// TODO: look into OnDelete:Cascade and AfterDelete https://stackoverflow.com/questions/76762629/how-to-cascade-a-delete-in-gorm
func (r *Repository) deleteUser(u *User) error {
	// no lookups here because i dont need to log the user struct
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return res.Error
}

func (r *Repository) createSession(userID string) (*Session, error) {
	sID, err := generateSessionID()
	if err != nil {
		return nil, err
	}
	e := time.Now().AddDate(0, 0, 30)
	s := &Session{
		ID:        sID,
		UserID:    userID,
		ExpiresAt: e,
	}
	res := r.db.Create(s)
	if res.Error != nil {
		return nil, res.Error
	}
	return s, nil
}

func (r *Repository) lookupSession(sID string) (*Session, error) {
	s := &Session{ID: sID}
	res := r.db.First(s)
	if res.Error != nil {
		return nil, res.Error
	}
	return s, nil
}

// updateSession performs a gorm.Save operation so it should always be used on lookuped sessions that have all fields
func (r *Repository) updateSession(sn *Session) (*Session, error) {
	res := r.db.Save(sn)
	if res.Error != nil {
		return nil, res.Error
	}
	return sn, nil
}

func (r *Repository) deleteSession(sID string) error {
	if sID == "" {
		return nil
	}
	s := &Session{ID: sID}
	res := r.db.Delete(s)
	return res.Error
}

func (r *Repository) selectAll() ([]User, error) {
	users := []User{}
	res := r.db.Find(&users)
	if res.Error != nil {
		return nil, res.Error
	}
	return users, nil
}
