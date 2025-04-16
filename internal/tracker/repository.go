package tracker

import (
	"errors"
	"job/internal/db"
	"job/internal/scheduler"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(d *gorm.DB) *Repository {
	return &Repository{db: d}
}

func (t *UnderlyingTracker) BeforeCreate(tx *gorm.DB) error {
	t.ID = db.NewPublicID(db.TrackerIdPrefix)
	return nil
}

func getDefaultGoalDeadline() time.Time {
	return scheduler.GetDefaultGoalDeadline()
}

func (t *JobAppTracker) BeforeCreate(tx *gorm.DB) error {
	t.ID = db.NewPublicID(db.TrackerIdPrefix)
	return nil
}

func (r *Repository) CreateJobAppTracker(t *JobAppTracker) (*JobAppTracker, error) {
	r.db.Create(t)
	if r.db.Error != nil {
		return nil, r.db.Error
	}
	return t, nil
}

func (r *Repository) LookupJobAppTracker(id string) (*JobAppTracker, error) {
	t := &JobAppTracker{ID: id}
	if err := r.db.First(t); err.Error != nil {
		return nil, err.Error
	}
	return t, nil
}

func (r *Repository) UpdateJobAppTracker(t *JobAppTracker) (*JobAppTracker, error) {
	r.db.Save(t)
	return t, nil
}

// as is this does a soft delete https://gorm.io/docs/delete.html#Soft-Delete
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

func (r *Repository) GetJobAppTrackerItems(trackerID string) ([]*JobAppItem, error) {
	var items []*JobAppItem
	if err := r.db.Where("tracker_id = ?", trackerID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
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

func (r *Repository) LookupJobAppTrackerItem(t *JobAppTracker, id string) (*JobAppItem, error) {
	i := &JobAppItem{ID: id}
	if err := r.db.Where("tracker_id = ?", t.ID).First(i).Error; err != nil {
		return nil, err
	}
	return i, nil
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
