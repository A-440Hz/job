package tracker

import (
	"time"

	"gorm.io/gorm"
)

type Frequency string

const (
	FreqDaily  Frequency = "daily"
	FreqWeekly Frequency = "weekly"
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
	ID            uint   `gorm:"primaryKey"`
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
	ID                string `gorm:"primaryKey"`
	Tracker           UnderlyingTracker
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
