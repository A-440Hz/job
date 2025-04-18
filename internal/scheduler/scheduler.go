package scheduler

import (
	"time"
)

// https://www.iana.org/time-zones
var DefaultTimezone *time.Location

// this is a bad design and I should use dependency injection in production code
func init() {
	var err error
	DefaultTimezone, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		DefaultTimezone = time.UTC
	}
}

// GetDefaultGoalDeadline attempts to return 1am at the user's local timezone, otherwise defaulting to DefaultTimezone
func GetDefaultGoalDeadline(l *time.Location) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 1, 0, 0, 0, l)
}
