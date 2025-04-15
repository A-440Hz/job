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

type Tracker interface {
	EditGoal() error
	AddItem() error
	EditItem() error
	RemoveItem() error
}

type JobAppTracker struct {
	ID                string `gorm:"primaryKey"`
	GoalDeadline      time.Time
	GoalFrequency     Frequency `gorm:"default:weekly"`
	GoalQuantity      int       `gorm:"default:5"`
	numBoxesAwarded   int       `gorm:"default:0"`
	numItemsCompleted int       `gorm:"default:0"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}
