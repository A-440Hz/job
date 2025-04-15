package tracker

import (
	"time"

	"gorm.io/gorm"
)

type ItemStatus string

const (
	StatusComplete   ItemStatus = "complete"
	StatusInProgress ItemStatus = "in progress"
)

type Item interface {
	IsComplete() bool
	EditStatus(s ItemStatus) error
}

type JobAppItem struct {
	ID         string `gorm:"primaryKey"`
	TrackerID  string `gorm:"foreignKey:ID"`
	Title      string
	Body       string
	status     ItemStatus
	attributed bool // this flips when a tracker progress is assigned from this item
	// probably sync shenanigans to iron out? attempt sync on each item creation?
	// also i can refactor this later to award more lootboxes when the status (rejected, etc) changes
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (i *JobAppItem) IsComplete() bool {
	return i.status == StatusComplete
}

func (i *JobAppItem) EditStatus(s ItemStatus) error {
	// noop (job apps are always complete in current design)
	return nil
}
