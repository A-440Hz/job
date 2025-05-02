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

func (r *Repository) lookupJobAppTrackerFromTrackerID(id string) (*JobAppTracker, error) {
	t := &JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}}
	res := r.db.First(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

func (r *Repository) LookupJobAppTrackerFromUserID(uuid string) (*JobAppTracker, error) {
	t := &JobAppTracker{UnderlyingTracker: UnderlyingTracker{UserID: uuid}}
	res := r.db.First(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

func (r *Repository) GetJobAppTrackerWithItemsFromTrackerID(id string) (*JobAppTracker, error) {
	t := &JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}}
	res := r.db.Preload("Items").First(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

func (r *Repository) GetJobAppTrackerWithItemsFromUserID(uuid string) (*JobAppTracker, error) {
	t := &JobAppTracker{UnderlyingTracker: UnderlyingTracker{UserID: uuid}}
	res := r.db.Preload("Items").First(t)
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

// GetScorableJobAppItems returns tracker items within the goal timeframe which are StatusComplete and not yet attributed
// returning list instead of pointers because of small expected return size and faster field access
// might need to switch because idk how well it will work with gorm
func (r *Repository) GetScorableJobAppItems(t *JobAppTracker) ([]JobAppItem, error) {
	items := []JobAppItem{}
	// is there a case for validation or do i assume tracker timeframes are always set correctly?
	begin, end := t.GetValidTimeframe()
	if begin.IsZero() || end.IsZero() {
		return nil, errors.New("invalid timeframe")
	}
	res := r.db.Where("tracker_id = ?", t.GetID()).
		Where("created_at BETWEEN ? AND ?", begin, end).
		Where("status = ? AND NOT is_attributed", StatusComplete).
		Order("created_at ASC").
		Find(&items)
	if res.Error != nil {
		return nil, res.Error
	}
	return items, nil
}

func (i *JobAppItem) BeforeCreate(tx *gorm.DB) error {
	i.ID = db.NewPublicID(db.ItemIdPrefix)
	return nil
}

func (r *Repository) CreateJobAppTrackerItem(i *JobAppItem) (*JobAppItem, error) {
	res := r.db.Create(i)
	if res.Error != nil {
		return nil, res.Error
	}
	return i, nil
}

func (r *Repository) LookupJobAppItems(trackerID string) ([]*JobAppItem, error) {
	var items []*JobAppItem
	res := r.db.Where("tracker_id = ?", trackerID).
		Order("created_at ASC"). // ordering options pattern? https://gorm.io/docs/query.html#Order
		Find(&items)
	if res.Error != nil {
		return nil, res.Error
	}
	return items, nil
}

func (r *Repository) LookupJobAppItem(trackerID, itemID string) (*JobAppItem, error) {
	item := &JobAppItem{ID: itemID}
	res := r.db.Where("tracker_id = ?", trackerID).First(&item)
	if res.Error != nil {
		return nil, res.Error
	}
	return item, nil
}

func (r *Repository) UpdateJobAppTrackerItemFields(i *JobAppItem, fields []string) (*JobAppItem, error) {
	res := r.db.Model(i).Select(fields).Updates(i)
	if res.Error != nil {
		return nil, res.Error
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
