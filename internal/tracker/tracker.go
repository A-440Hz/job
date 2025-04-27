package tracker

import (
	"reflect"
	"time"

	"gorm.io/gorm"
)

type Frequency string

var frequencies = []Frequency{FreqDaily, FreqWeekly}

const (
	FreqDaily   Frequency = "daily"
	FreqWeekly  Frequency = "weekly"
	DefaultFreq           = FreqWeekly

	// underlying tracker fields
	goalDeadlineField           = "goal_deadline"
	goalFrequencyField          = "goal_frequency"
	goalQuantityField           = "goal_quantity"
	curGoalStreakField          = "cur_goal_streak"
	maxGoalStreakField          = "max_goal_streak"
	maxItemsCompletedDailyField = "max_items_completed_daily"
	totalItemsCompletedField    = "total_items_completed"
	totalBoxesAwardedField      = "total_boxes_awarded"
	firstCompletedField         = "first_completed"

	// job app tracker fields
	numBoxesAwardedField   = "num_boxes_awarded"
	numItemsCompletedField = "num_items_completed"
)

// chatgpt says:
// Polymorphism / Inheritance is Not Native to GORM
// Go doesn’t have inheritance, and GORM doesn't natively support polymorphic models in the way that, say, Django does.
// You'll need to simulate it via composition and good relationship structuring.
// initially I wanted a StatsTracker struct as a layer between JobAppTracker and UnderlyingTracker, but it seems simpler to
// move the stats mechanism into the UnderlyingTracker struct itself.
// This is fine as long as different tracker types can share the same type of stats.

type TrackerInterface interface {
	AddItem(Item) error
	RemoveItem(Item) error
	RefreshStatus() error
	ResetProgress() error
	EditGoal(f Frequency, quantity int) error
	GetStats() (*TrackerStats, error)
}

type UnderlyingTracker struct {
	// I shouldn't need to embed a User. A foreign key is sufficient.
	ID            string `gorm:"primaryKey"`
	UserID        string `gorm:"index"` // the index tag improves query performance for common lookup fields
	GoalDeadline  time.Time
	GoalFrequency Frequency `gorm:"default:weekly"`
	GoalQuantity  int       `gorm:"default:5"`

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

// essentially this is a item factory? it creates JobAppItems and assigns them to the UnderlyingTracker
type JobAppTracker struct {
	UnderlyingTracker
	numBoxesAwarded   int `gorm:"default:0"`
	numItemsCompleted int `gorm:"default:0"`
}

// TrackerStats is a json object for UnderlyingTracker to return
type TrackerStats struct {
	GoalStreak          int     `json:"goalStreak"`
	MaxGoalStreak       int     `json:"maxGoalStreak"`
	TotalItemsCompleted int     `json:"totalItemsCompleted"`
	TotalBoxesAwarded   int     `json:"totalBoxesAwarded"`
	AvgDailyCompleted   float64 `json:"avgDailyCompleted"`
}

func (t *JobAppTracker) GetID() string {
	return t.ID
}

type UnderlyingTrackerUpdateFields struct {
	GoalDeadline  *time.Time `json:"goalDeadline,omitempty"`
	GoalFrequency *string    `json:"goalFrequency,omitempty"`
	GoalQuantity  *int       `json:"goalQuantity,omitempty"`

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
		t.GoalFrequency = Frequency(*uf.GoalFrequency)
		fields = append(fields, goalFrequencyField)
	}
	if uf.GoalQuantity != nil {
		t.GoalQuantity = *uf.GoalQuantity
		fields = append(fields, goalQuantityField)
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
	NumBoxesAwarded   *int `json:"numBoxesAwarded,omitempty"`
	NumItemsCompleted *int `json:"numItemsCompleted,omitempty"`
	UnderlyingTrackerUpdateFields
}

func (uf *JobAppTrackerUpdateFields) formatForRepo() (*JobAppTracker, []string, error) {
	t := &JobAppTracker{}
	fields := []string{}
	if uf.NumBoxesAwarded != nil {
		t.numBoxesAwarded = *uf.NumBoxesAwarded
		fields = append(fields, "numBoxesAwarded")
	}
	if uf.NumItemsCompleted != nil {
		t.numItemsCompleted = *uf.NumItemsCompleted
		fields = append(fields, "numItemsCompleted")
	}
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
