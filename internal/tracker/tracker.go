package tracker

import (
	"errors"
	"fmt"
	"time"
)

type Frequency string

const (
	FreqDaily  Frequency = "daily"
	FreqWeekly Frequency = "weekly"
)

type Tracker interface {
	// getGoal()
	EditGoal() error
	AddItem() error
	EditItem() error
}

type JobAppTracker struct {
	TrackerId         string `gorm:"primaryKey"`
	GoalDeadline      time.Time
	GoalFrequency     Frequency
	GoalQuantity      int
	numBoxesAwarded   int
	numItemsCompleted int
}

func (t *JobAppTracker) EditFrequency(f Frequency) error {
	if f == t.GoalFrequency {
		return nil
	}
	t.GoalFrequency = f
	// db logic here
	// scheduler logic here
	return nil
}

func (t *JobAppTracker) ResetProgress() {
	t.numBoxesAwarded = 0
	t.numItemsCompleted = 0
}

func (t *JobAppTracker) getNextDeadline() (time.Time, error) {
	cd := t.GoalDeadline
	switch t.GoalFrequency {
	case FreqDaily:
		return cd.Add(24 * time.Hour), nil
	case FreqWeekly:
		return cd.Add(24 * time.Hour * 7), nil
	default:
		return cd, errors.New(fmt.Sprintf("Invalid goal frequency: %q", t.GoalFrequency))
	}
}
