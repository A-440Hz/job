package tracker

import (
	"errors"
	"job/internal/db"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(d *gorm.DB) *Repository {
	return &Repository{db: d}
}

// func (t *UnderlyingTracker) BeforeCreate(tx *gorm.DB) error {
// 	t.ID = db.NewPublicID(db.TrackerIdPrefix)
// 	return nil
// }

func (t *JobAppTracker) BeforeCreate(tx *gorm.DB) error {
	t.ID = db.NewPublicID(db.TrackerIdPrefix)
	return nil
}

// user is validated in the service layer
// func (r *Repository) CreateUnderlyingTracker(u *user.User) (*UnderlyingTracker, error) {
// 	t := &UnderlyingTracker{
// 		UserID:       u.GetID(),
// 		GoalDeadline: scheduler.GetDefaultGoalDeadline(u.Timezone.Location),
// 	}
// 	res := r.db.Create(t)
// 	if res.Error != nil {
// 		return nil, res.Error
// 	}
// 	return t, nil
// }

func (r *Repository) CreateJobAppTracker(t *JobAppTracker) (*JobAppTracker, error) {
	res := r.db.Create(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

func (r *Repository) LookupJobAppTracker(id string) (*JobAppTracker, error) {
	t := &JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}}
	res := r.db.First(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

func (r *Repository) GetJobAppTrackerFromUserID(uuid string) (*JobAppTracker, error) {
	t := &JobAppTracker{}
	res := r.db.Where("user_id = ?", uuid).First(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

// TODO: change this to update mask later
func (r *Repository) UpdateJobAppTracker(t *JobAppTracker) (*JobAppTracker, error) {
	res := r.db.Save(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

func (r *Repository) UpdateJobAppTrackerFields(t *JobAppTracker, fields []string) (*JobAppTracker, error) {
	res := r.db.Model(t).Select(fields).Updates(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

// DeleteJobAppTracker soft deletes tracker t https://gorm.io/docs/delete.html#Soft-Delete
func (r *Repository) DeleteJobAppTracker(t *JobAppTracker) error {
	res := r.db.Delete(t)
	if res.RowsAffected == 0 {
		return errors.New("tracker not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (i *JobAppItem) BeforeCreate(tx *gorm.DB) error {
	i.ID = db.NewPublicID(db.ItemIdPrefix)
	return nil
}

func (r *Repository) CreateJobAppTrackerItem(t *JobAppTracker, title, body string) (*JobAppItem, error) {
	i := &JobAppItem{
		TrackerID: t.ID,
		Title:     title,
		Body:      body,
	}
	if err := r.db.Create(i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

func (r *Repository) LookupJobAppTrackerItems(trackerID string) ([]*JobAppItem, error) {
	var items []*JobAppItem
	res := r.db.Where("tracker_id = ?", trackerID).
		Order("created_at ASC"). // ordering options pattern? https://gorm.io/docs/query.html#Order
		Find(&items)
	if res.Error != nil {
		return nil, res.Error
	}
	return items, nil
}

func (r *Repository) UpdateJobAppTrackerItemField(trackerID string, itemID string, fields []string) (*JobAppItem, error) {
	var item *JobAppItem
	res := r.db.Model(item).Where("tracker_id = ? AND id = ?", trackerID, itemID).Select(fields).Updates(item)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("item not found")
	}
	return item, nil
}

func (r *Repository) UpdateJobAppTrackerItem(t *JobAppTracker, i *JobAppItem) (*JobAppItem, error) {
	if err := r.db.Save(i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

func (r *Repository) DeleteJobAppTrackerItem(t *JobAppTracker, i *JobAppItem) error {
	res := r.db.Delete(i)
	if res.RowsAffected == 0 {
		return errors.New("item not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}
