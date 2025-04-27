package tracker

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ItemStatus string

const (
	StatusComplete   ItemStatus = "complete"
	StatusInProgress ItemStatus = "in progress"

	titleField           = "title"
	bodyField            = "body"
	statusField          = "status"
	isAttributedField    = "is_attributed"
	attributionTimeField = "attribution_time"
)

type Item interface {
	IsComplete() bool
	EditStatus(s ItemStatus) error
}

type JobAppItem struct {
	ID           string `gorm:"primaryKey"`
	TrackerID    string `gorm:"index"`
	Title        string // maybe separate this into Company and Position
	Body         string
	Status       ItemStatus
	IsAttributed bool // this flips when a tracker progress is assigned from this item
	// probably sync shenanigans to iron out? attempt sync on each item creation?
	// also i can refactor this later to award more lootboxes when the status (rejected, etc) changes
	AttributionTime *time.Time // might be useful for solving sync issues/debugging later
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (i *JobAppItem) IsComplete() bool {
	return i.Status == StatusComplete
}

func (i *JobAppItem) EditStatus(s ItemStatus) error {
	// noop (job apps are always complete in current design)
	return nil
}

func (s *ItemStatus) Scan(value any) error {
	if value == nil {
		return nil
	}
	status, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid status type scanned: %T", value)
	}
	*s = ItemStatus(status)
	return nil
}

func (s *ItemStatus) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return string(*s), nil
}

type JobAppItemUpdateFields struct {
	Title *string `json:"title,omitempty"`
}
