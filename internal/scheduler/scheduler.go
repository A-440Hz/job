package scheduler

import (
	"time"
)

// https://www.iana.org/time-zones
// yeah I can't find anything from that link.
// see /usr/share/zoneinfo on linux
const defaultTimezoneOffset = -7 * 60 * 60 // UTC-7
// DefaultTimezone is the default timezone for the app. It is set to UTC-7 (PST) by default.

// GetDefaultTimezone returns a Timezone object corresponding to UTC-7 (PST)
func GetDefaultTimezone() *Timezone {
	return NewTimezone(defaultTimezoneOffset)
}

// GetDefaultGoalDeadline returns 1AM at location l, defaulting to DefaultTimezone
func GetDefaultGoalDeadline(l *time.Location) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 1, 0, 0, 0, l)
}
