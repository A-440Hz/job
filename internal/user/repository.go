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

// func (u *User) AfterDelete(tx *gorm.DB) error {
// 	tx.Delete("job_app_trackers")
// 	return nil
// }

func (r *Repository) createUser(u *User) (*User, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) getUserAndUserInventory(id string) (*User, error) {
	u := &User{ID: id}
	res := r.db.Preload("Inventory").First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}
func (r *Repository) lookupUser(id string) (*User, error) {
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	u := &User{ID: id}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	// opportunity to return gorm.ErrRecordNotFound differently
	return u, nil
}

func (r *Repository) lookupUserByUsername(uname string) (*User, error) {
	if uname == "" {
		return nil, gorm.ErrRecordNotFound
	}
	u := &User{}
	res := r.db.Where("username = ?", uname).First(u)
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
	if u == nil || u.ID == "" {
		return gorm.ErrRecordNotFound
	}
	badFields := map[string]string{}
	res := r.db.Table("job_app_trackers").Where("user_id = ?", u.ID).Delete(nil)
	if res.Error != nil {
		badFields["job_app_trackers"] = res.Error.Error()
	}
	res = r.db.Delete(u)
	if res.Error != nil {
		badFields["users"] = res.Error.Error()
	}
	var err error
	if len(badFields) > 0 {
		err = errors.New("delete error:")
		for id, v := range badFields {
			err = errors.Join(err, errors.New(id+": "+v))
		}
	}
	return err
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

func (r *Repository) deleteUserSessions(uID string) error {
	if uID == "" {
		return nil
	}
	sessions := []Session{}
	res := r.db.Where("user_id = ?", uID).Delete(sessions)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) getExpiredSessions() ([]Session, error) {
	sessions := []Session{}
	today := scheduler.GetCurrentServerDay()
	res := r.db.Where("expires_at < ?", today).Find(&sessions)
	if res.Error != nil {
		return nil, res.Error
	}
	return sessions, nil
}

// getExpiredDemoUsers selects unregistered users with 0 created items created before the grace period of (2) days
func (r *Repository) getExpiredDemoUsers() ([]User, error) {
	cutoff := scheduler.GetCurrentServerDay().AddDate(0, 0, -2)
	users := []User{}
	res := r.db.Joins("JOIN user_inventories ON user_inventories.user_id = users.id").
		Where("registered = ?", false).
		Where("user_inventories.last_completed = ?", time.Time{}).
		Where("users.created_at < ?", cutoff).
		Find(&users)

	if res.Error != nil {
		return nil, res.Error
	}
	return users, nil
}

func (r *Repository) selectAllUsers() ([]User, error) {
	users := []User{}
	res := r.db.Preload("Inventory").Find(&users)
	if res.Error != nil {
		return nil, res.Error
	}
	return users, nil
}

func (r *Repository) selectAllSessions() ([]Session, error) {
	sessions := []Session{}
	res := r.db.Find(&sessions)
	if res.Error != nil {
		return nil, res.Error
	}
	return sessions, nil
}
