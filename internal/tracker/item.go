package tracker

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"
)

type ItemStatus string

var ValidJobAppItemStatus []ItemStatus = []ItemStatus{StatusComplete, StatusInProgress}

const (
	StatusComplete   ItemStatus = "complete"
	StatusInProgress ItemStatus = "in progress"
	StatusDefault    ItemStatus = StatusComplete

	titleField           = "title"
	bodyField            = "body"
	urlField             = "url"
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
	Title        string // Company: Role
	Body         string // optional description
	Url          string
	Status       ItemStatus
	IsAttributed bool
	// probably sync shenanigans to iron out? attempt sync on each item creation?
	// also i can refactor this later to award more lootboxes when the status (rejected, etc) changes
	AttributionTime *time.Time // might be useful for solving sync issues/debugging later
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (i *JobAppItem) GetID() string {
	return i.ID
}

func (i *JobAppItem) IsScorable() bool {
	return i.Status == StatusComplete && !i.IsAttributed
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

// TODO: IsScorable
type JobAppItemUpdateFields struct {
	Title           *string     `json:"title,omitempty"`
	Body            *string     `json:"body,omitempty"`
	Url             *string     `json:"url,omitempty"`
	Status          *ItemStatus `json:"status,omitempty"`
	IsAttributed    *bool       `json:"isAttributed,omitempty"`
	attributionTime *time.Time  // no json tag here because this is set internally
}

// formatForRepo returns error when Status is not a ValidJobAppItemStatus
func (uf *JobAppItemUpdateFields) formatForRepo() (*JobAppItem, []string, error) {
	j := &JobAppItem{}
	fields := []string{}
	if uf.Title != nil {
		j.Title = *uf.Title
		fields = append(fields, titleField)
	}
	if uf.Body != nil {
		j.Body = *uf.Body
		fields = append(fields, bodyField)
	}
	if uf.Url != nil {
		j.Url = *uf.Url
		fields = append(fields, urlField)
	}
	if uf.Status != nil {
		if !slices.Contains(ValidJobAppItemStatus, *uf.Status) {
			return nil, nil, fmt.Errorf("invalid application status: %q", *uf.Status)
		}
		j.Status = *uf.Status
		fields = append(fields, statusField)
	}
	if uf.IsAttributed != nil {
		j.IsAttributed = *uf.IsAttributed
		fields = append(fields, isAttributedField)
	}
	if uf.attributionTime != nil {
		j.AttributionTime = uf.attributionTime
		fields = append(fields, attributionTimeField)
	}
	return j, fields, nil
}
