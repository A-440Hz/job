package scheduler

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"time"
)

// DefaultTimezone is the default timezone for the app. It is set to UTC-7 (PST) by default.
const (
	FreqDaily   Frequency = "daily"
	FreqWeekly  Frequency = "weekly"
	DefaultFreq           = FreqWeekly

	defaultTimezoneOffset = -7 * 60 * 60 // UTC-7 (seconds east of UTC)
)

// Frequency denotes the number of days between deadlines
type Frequency string

var freqToDays = map[Frequency]int{FreqDaily: 1, FreqWeekly: 7}

func (f *Frequency) NumDays() int {
	return freqToDays[*f]
}

func IsValidFrequency(f Frequency) bool {
	if _, exists := freqToDays[f]; !exists {
		return false
	}
	return true
}

func GetDefaultGoalFrequency() Frequency {
	return DefaultFreq
}

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

// Timezone is a Valuer/Scanner interface for gorm to store time.Location as a basic type
// https://gorm.io/docs/data_types.html#Custom-Data-Types
// offset represents seconds east of UTC. UTC-7 is -7 * 60 * 60 = -25200
type Timezone struct {
	Location *time.Location
	offset   int64
}

func NewTimezoneWithOffset(o int64) *Timezone {
	return &Timezone{
		Location: time.FixedZone("", int(o)),
		offset:   o,
	}
}

func (t *Timezone) Value() (driver.Value, error) {
	return t.offset, nil
}

func (t *Timezone) Scan(value any) error {
	if value == nil {
		return errors.New("cannot scan nil timezone value")
	}
	offset, ok := value.(int64)
	if !ok {
		return fmt.Errorf("invalid timezone type scanned: %T", value)
	}
	t.Location = time.FixedZone("", int(offset))
	return nil
}

func (t *Timezone) Equal(other *Timezone) bool {
	if t == nil && other == nil {
		return true
	}
	if t == nil || other == nil {
		return false
	}
	return t.offset == other.offset
}
