package collection

import (
	"errors"
	"fmt"
	"time"
)

const (
	ValueCommon    CollectableValue = "C"
	ValueRare      CollectableValue = "B"
	ValueUltraRare CollectableValue = "A"
	ValueLegendary CollectableValue = "S"

	FormatJpg CollectableFormat = "image"
	TypeVideo CollectableFormat = "video"
)

type CollectableValue string

var values map[CollectableValue]int = map[CollectableValue]int{ValueCommon: 5, ValueRare: 15, ValueUltraRare: 35, ValueLegendary: 135}

func (c CollectableValue) ToPoints() int {
	return values[c]
}

type CollectableFormat string

// update fields
const (
	quantityField        = "quantity"
	isNewField           = "is_new"
	earnedAtField        = "earned_at"
	numLootboxesField    = "num_lootboxes"
	employmentCoinsField = "employment_coins"
)

type Collectable struct {
	ID       int `gorm:"primaryKey"`
	Name     string
	Value    CollectableValue
	Type     CollectableFormat
	Filename string
}

type UserCollectable struct {
	UserID        string `gorm:"primaryKey"`
	CollectableID int    `gorm:"foreignKey"`
	Quantity      int    // represents how many collectables this user owns. Expected collecable size: 20-300
	IsNew         bool   `gorm:"default:false"`
	EarnedAt      time.Time
	// Variant/UpgradeLevel
}

type UserCollectableUpdateFields struct {
	Quantity *int       `json:"quantity,omitempty"`
	IsNew    *bool      `json:"isNew,omitempty"`
	EarnedAt *time.Time `json:"earnedAt,omitempty"`
}

type UserInventory struct {
	UserID          string `gorm:"primaryKey"`
	NumLootboxes    int    `gorm:"default:0"`
	EmploymentCoins int    `gorm:"default:0"`
}

type UserInventoryUpdateFields struct {
	NumLootboxes    *int `json:"numLootboxes,omitempty"`
	EmploymentCoins *int `json:"employmentCoins,omitempty"`
}

func (uf *UserInventoryUpdateFields) formatForRepo() (*UserInventory, []string, error) {
	i := &UserInventory{}
	fields := []string{}
	badFields := map[string]string{}
	if uf.NumLootboxes != nil {
		if *uf.NumLootboxes < 0 {
			badFields[numLootboxesField] = fmt.Sprintf("cannot be less than 0: %v", *uf.NumLootboxes)
		}
		i.NumLootboxes = *uf.NumLootboxes
	}
	if uf.EmploymentCoins != nil {
		if *uf.EmploymentCoins < 0 {
			badFields[employmentCoinsField] = fmt.Sprintf("cannot be less than 0: %v", *uf.EmploymentCoins)
		}
		i.EmploymentCoins = *uf.EmploymentCoins
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for id, v := range badFields {
			err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
		}
		return nil, nil, err
	}
	return i, fields, nil
}

func (uf *UserCollectableUpdateFields) formatForRepo() (*UserCollectable, []string, error) {
	c := &UserCollectable{}
	fields := []string{}
	if uf.Quantity != nil {
		if *uf.Quantity < 0 {
			return nil, nil, fmt.Errorf("invalid collectable quantity: %d", *uf.Quantity)
		}
		c.Quantity = *uf.Quantity
		fields = append(fields, quantityField)
	}
	if uf.IsNew != nil {
		c.IsNew = *uf.IsNew
		fields = append(fields, isNewField)
	}
	if uf.EarnedAt != nil {
		c.EarnedAt = *uf.EarnedAt
		fields = append(fields, earnedAtField)
	}
	return c, fields, nil
}
