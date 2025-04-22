package scheduler

import (
	"database/sql/driver"
	"errors"
	"time"
)

// Timezone is a Valuer/Scanner interface for gorm to store time.Location as a basic type
// https://gorm.io/docs/data_types.html#Custom-Data-Types
// offset represents seconds east of UTC. UTC-7 is -7 * 60 * 60 = -25200
type Timezone struct {
	Location *time.Location
	offset   int
}

func NewTimezone(o int) *Timezone {
	return &Timezone{
		Location: time.FixedZone("", o),
		offset:   o,
	}
}

func (t *Timezone) Value() (driver.Value, error) {
	return t.offset, nil
}

func (t *Timezone) Scan(value any) error {
	if value == nil {
		t.Location = nil
		return nil
	}
	offset, ok := value.(int)
	if !ok {
		return errors.New("invalid timezone value")
	}
	t.Location = time.FixedZone("", offset)
	return nil
}
