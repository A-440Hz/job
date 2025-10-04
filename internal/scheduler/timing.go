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

func (f *Frequency) StrPtr() *string {
	s := string(*f)
	return &s
}

func IsValidFrequency(f Frequency) bool {
	if _, exists := freqToDays[f]; !exists {
		return false
	}
	return true
}

func GetDefaultCycleFrequency() Frequency {
	return DefaultFreq
}

// GetDefaultTimezone returns a Timezone object corresponding to UTC-7 (PST)
func GetDefaultTimezone() *Timezone {
	return NewTimezoneWithOffset(defaultTimezoneOffset)
}

// GetDefaultCycleDeadline returns 1AM in 6 days at location l, defaulting to DefaultTimezone
// 6 days because it is an easy value to get consistent test results from the scheduler
func GetDefaultCycleDeadline(z *Timezone) time.Time {
	var l *time.Location
	if z == nil {
		log.Println("user timezone was nil -- mutating value to default timezone location")
		l = GetDefaultTimezone().Location
	} else {
		l = z.Location
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+6, 1, 0, 0, 0, l)
}

func GetCurrentServerDay() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}

// OneDayApart is intended for use with truncated times from GetCurrentServerDay
func OneDayApart(new, old time.Time) bool {
	return new.Sub(old) < (time.Hour*25) && !new.Equal(old)
}

// Timezone is a Valuer/Scanner interface for gorm to store time.Location as a basic type
// https://gorm.io/docs/data_types.html#Custom-Data-Types
// offsetSeconds represents seconds east of UTC, used for time.FixedZone syntax. UTC-7 is -7 * 60 * 60 = -25200
type Timezone struct {
	Location      *time.Location
	OffsetSeconds int64
}

func NewTimezoneWithOffset(o int64) *Timezone {
	return &Timezone{
		Location:      time.FixedZone("", int(o)),
		OffsetSeconds: o,
	}
}

// GetOffset is used for testing only
func (t *Timezone) GetOffset() int64 {
	return t.OffsetSeconds
}

func (t *Timezone) Value() (driver.Value, error) {
	return t.OffsetSeconds, nil
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
	t.OffsetSeconds = offset
	return nil
}

func (t *Timezone) Equal(other *Timezone) bool {
	if t == nil && other == nil {
		return true
	}
	if t == nil || other == nil {
		return false
	}
	return t.OffsetSeconds == other.OffsetSeconds
}
