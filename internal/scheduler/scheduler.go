package scheduler

import (
	"log"
	"time"
)

// https://www.iana.org/time-zones
// yeah I can't find anything from that link.
// see /usr/share/zoneinfo on linux

// DefaultTimezone is the default timezone for the app. It is set to UTC-7 (PST) by default.
const defaultTimezoneOffset = -7 * 60 * 60 // UTC-7

// GetDefaultTimezone returns a Timezone object corresponding to UTC-7 (PST)
func GetDefaultTimezone() *Timezone {
	return NewTimezoneWithOffset(defaultTimezoneOffset)
}

// GetDefaultGoalDeadline returns 1AM at location l, defaulting to DefaultTimezone
func GetDefaultGoalDeadline(z *Timezone) time.Time {
	var l *time.Location
	if z == nil {
		log.Println("user timezone was nil -- mutating value to default timezone location")
		l = GetDefaultTimezone().Location
	} else {
		l = z.Location
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 1, 0, 0, 0, l)
}
