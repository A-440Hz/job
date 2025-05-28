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

func (t *JobAppTracker) BeforeCreate(tx *gorm.DB) error {
	t.ID = db.NewPublicID(db.TrackerIdPrefix)
	return nil
}

// CreateJobAppTracker relies on an UnderlyingTracker being set in the service layer
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
	t := &JobAppTracker{}
	res := r.db.Where("user_id = ?", uuid).First(t)
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
	t := &JobAppTracker{}
	res := r.db.Preload("Items").Where("user_id = ?", uuid).First(t)
	if res.Error != nil {
		return nil, res.Error
	}
	return t, nil
}

// updateUnderlyingTrackerFields is a dispatcher which queries the tracker type and calls the appropriate tracker update method
func (r *Repository) updateUnderlyingTrackerFields(t *UnderlyingTracker, fields []string) error {
	var err error
	switch t.TrackerType {
	case JobAppTrackerType:
		t := &JobAppTracker{UnderlyingTracker: *t}
		_, err = r.UpdateJobAppTrackerFields(t, fields)
	default:
		err = errors.New("invalid tracker type")
	}
	return err
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

// GetScorableJobAppItems returns scorable (StatusComplete && !IsAttributed) tracker items created within the goal timeframe which are
// returning list instead of pointers because of small expected return size and faster field access
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

func (r *Repository) DeleteJobAppTrackerItem(i *JobAppItem) error {
	res := r.db.Delete(i)
	if res.RowsAffected == 0 {
		return errors.New("item not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// getAllUnderlyingTrackers runs on startup, retrieving repo trackers of every type and feeding them into the scheduler
// it currently does not filter "expired" trackers and has no
func (r *Repository) getAllUnderlyingTrackers() ([]UnderlyingTracker, error) {
	allUnderlying := []UnderlyingTracker{}
	jobAppTrackers := []JobAppTracker{}
	res := r.db.Find(&jobAppTrackers)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, jat := range jobAppTrackers {
		allUnderlying = append(allUnderlying, jat.UnderlyingTracker)
	}
	// insert more tracker types here:
	return allUnderlying, nil
}

func (r *Repository) selectAllJobAppTrackers() ([]JobAppTracker, error) {
	trackers := []JobAppTracker{}
	res := r.db.Preload("Items").Find(&trackers)
	if res.Error != nil {
		return nil, res.Error
	}
	return trackers, nil
}
