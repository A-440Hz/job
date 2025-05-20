package tracker

import (
	"fmt"
	"job/internal/scheduler"
	"reflect"
	"time"

	"gorm.io/gorm"
)

const (
	// underlying tracker fields
	goalDeadlineField           = "goal_deadline"
	goalFrequencyField          = "goal_frequency"
	goalQuantityField           = "goal_quantity"
	curItemsCompletedField      = "cur_items_completed"
	curBoxesAwardedField        = "cur_boxes_awarded"
	curGoalStreakField          = "cur_goal_streak"
	maxGoalStreakField          = "max_goal_streak"
	maxItemsCompletedDailyField = "max_items_completed_daily"
	totalItemsCompletedField    = "total_items_completed"
	totalBoxesAwardedField      = "total_boxes_awarded"
	firstCompletedField         = "first_completed"
)

// chatgpt says:
// Polymorphism / Inheritance is Not Native to GORM
// Go doesn’t have inheritance, and GORM doesn't natively support polymorphic models in the way that, say, Django does.
// You'll need to simulate it via composition and good relationship structuring.
// initially I wanted a StatsTracker struct as a layer between JobAppTracker and UnderlyingTracker, but it seems simpler to
// move the stats mechanism into the UnderlyingTracker struct itself.
// This is fine as long as different tracker types can share the same type of stats.

type TrackerType string

func (tt TrackerType) StringPtr() *string {
	s := string(tt)
	return &s
}

const (
	JobAppTrackerType TrackerType = "job_app_tracker"
)

// The underlying tracker is a base struct that contains the common fields for all trackers.
// There is no good reason for it to be a separate entity in the database, so I will create it in memory and store the two trackers together in gorm.
// The separation is primarily to fulfill the composite pattern and hold specific types of Items.
type UnderlyingTracker struct {
	// I shouldn't need to embed a User. A foreign key is sufficient. The User attributes will displayed separately on the user page
	ID                string              `gorm:"primaryKey"`
	UserID            string              `gorm:"index"` // the index tag improves query performance for common lookup fields
	GoalDeadline      time.Time           // when this time is reached, CurItemsCompleted will be reset and the deadline is pushed forward by GoalFrequency
	GoalFrequency     scheduler.Frequency `gorm:"default:weekly"` // ideally this default should be overriden in the create hooks, per tracker type
	GoalQuantity      int                 `gorm:"default:5"`      // the target number of items to complete within the deadline to reward a box
	CurItemsCompleted int                 `gorm:"default:0"`      // the number of scorable items within this deadline that have not yet been converted
	CurBoxesAwarded   int                 `gorm:"default:0"`      // the number of boxes awarded
	TrackerType       TrackerType         // TrackerType helps link the UnderlyingTracker with its respective Update method

	// stats
	CurGoalStreak          int `gorm:"default:0"`
	MaxGoalStreak          int `gorm:"default:0"`
	MaxItemsCompletedDaily int `gorm:"default:0"`
	TotalItemsCompleted    int `gorm:"default:0"`
	TotalBoxesAwarded      int `gorm:"default:0"` // idk about this one.. it sounds like somthing for Collection to track
	FirstCompleted         *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              gorm.DeletedAt `gorm:"index"`
}

// this is a composite which holds JobAppItems and associates them to the UnderlyingTracker
type JobAppTracker struct {
	UnderlyingTracker
	// gorm does not automatically fetch foreign key fields unless explicitly Preloaded
	Items []JobAppItem `gorm:"foreignKey:TrackerID;references:ID"`
}

// TrackerStats is a json object for UnderlyingTracker to return
type TrackerStats struct {
	CurGoalStreak          int     `json:"curGoalStreak"`
	MaxGoalStreak          int     `json:"maxGoalStreak"`
	MaxItemsCompletedDaily int     `json:"maxItemsCompletedDaily"`
	TotalItemsCompleted    int     `json:"totalItemsCompleted"`
	TotalBoxesAwarded      int     `json:"totalBoxesAwarded"`
	AvgDailyCompleted      float64 `json:"avgDailyCompleted"`
}

// GetID returns the "TrackerID" primary key of the UnderlyingTracker
func (t *JobAppTracker) GetID() string {
	return t.ID
}

func (t *JobAppTracker) GetUserID() string {
	return t.UserID
}
func (t *UnderlyingTracker) GetValidTimeframe() (time.Time, time.Time) {
	days := -1 * t.GoalFrequency.NumDays()
	begin := t.GoalDeadline.AddDate(0, 0, days)
	return begin, t.GoalDeadline
}

func (t *UnderlyingTracker) ToTrackerGoal() *scheduler.TrackerGoal {
	return &scheduler.TrackerGoal{
		TrackerID:     t.ID,
		GoalDeadline:  t.GoalDeadline,
		GoalFrequency: t.GoalFrequency,
		TrackerType:   *t.TrackerType.StringPtr(),
	}
}

type UnderlyingTrackerUpdateFields struct {
	GoalDeadline      *time.Time `json:"goalDeadline,omitempty"`
	GoalFrequency     *string    `json:"goalFrequency,omitempty"`
	GoalQuantity      *int       `json:"goalQuantity,omitempty"`
	CurItemsCompleted *int       `json:"curItemsCompleted,omitempty"`
	CurBoxesAwarded   *int       `json:"curBoxesAwarded,omitempty"`

	//stats
	CurGoalStreak          *int       `json:"curGoalStreak,omitempty"`
	MaxGoalStreak          *int       `json:"maxGoalStreak,omitempty"`
	MaxItemsCompletedDaily *int       `json:"maxItemsCompletedDaily,omitempty"`
	TotalItemsCompleted    *int       `json:"totalItemsCompleted,omitempty"`
	TotalBoxesAwarded      *int       `json:"totalBoxesAwarded,omitempty"`
	FirstCompleted         *time.Time `json:"firstCompleted,omitempty"`
}

func (uf *UnderlyingTrackerUpdateFields) formatForRepo() (*UnderlyingTracker, []string, error) {
	t := &UnderlyingTracker{}
	fields := []string{}
	if uf.GoalDeadline != nil {
		t.GoalDeadline = *uf.GoalDeadline
		fields = append(fields, goalDeadlineField)
	}
	if uf.GoalFrequency != nil {
		f := scheduler.Frequency(*uf.GoalFrequency)
		if !scheduler.IsValidFrequency(f) {
			return nil, nil, fmt.Errorf("invalid goal frequency: %q", f)
		}
		t.GoalFrequency = f
		fields = append(fields, goalFrequencyField)
	}
	if uf.GoalQuantity != nil {
		t.GoalQuantity = *uf.GoalQuantity
		fields = append(fields, goalQuantityField)
	}
	if uf.CurItemsCompleted != nil {
		t.CurItemsCompleted = *uf.CurItemsCompleted
		fields = append(fields, curItemsCompletedField)
	}
	if uf.CurBoxesAwarded != nil {
		t.CurBoxesAwarded = *uf.CurBoxesAwarded
		fields = append(fields, curBoxesAwardedField)
	}
	if uf.CurGoalStreak != nil {
		t.CurGoalStreak = *uf.CurGoalStreak
		fields = append(fields, curGoalStreakField)
	}
	if uf.MaxGoalStreak != nil {
		t.MaxGoalStreak = *uf.MaxGoalStreak
		fields = append(fields, maxGoalStreakField)
	}
	if uf.MaxItemsCompletedDaily != nil {
		t.MaxItemsCompletedDaily = *uf.MaxItemsCompletedDaily
		fields = append(fields, maxItemsCompletedDailyField)
	}
	if uf.TotalItemsCompleted != nil {
		t.TotalItemsCompleted = *uf.TotalItemsCompleted
		fields = append(fields, totalItemsCompletedField)
	}
	if uf.TotalBoxesAwarded != nil {
		t.TotalBoxesAwarded = *uf.TotalBoxesAwarded
		fields = append(fields, totalBoxesAwardedField)
	}
	if uf.FirstCompleted != nil {
		t.FirstCompleted = uf.FirstCompleted
		fields = append(fields, firstCompletedField)
	}
	return t, fields, nil
}

func (uf *UnderlyingTrackerUpdateFields) IsNil() bool {
	return uf == nil || reflect.DeepEqual(uf, &UnderlyingTrackerUpdateFields{})
}

type JobAppTrackerUpdateFields struct {
	UnderlyingTrackerUpdateFields
}

func (uf *JobAppTrackerUpdateFields) formatForRepo() (*JobAppTracker, []string, error) {
	t := &JobAppTracker{}
	fields := []string{}
	if !uf.UnderlyingTrackerUpdateFields.IsNil() {
		ut, f, err := uf.UnderlyingTrackerUpdateFields.formatForRepo()
		if err != nil {
			return nil, nil, err
		}
		t.UnderlyingTracker = *ut
		fields = append(fields, f...)
	}
	return t, fields, nil
}
