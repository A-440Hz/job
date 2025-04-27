package scheduler

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

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
