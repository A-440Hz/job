package collection

import "time"

const (
	ValueCommon    CollectableValue = "C"
	ValueRare      CollectableValue = "B"
	ValueUltraRare CollectableValue = "A"
	ValueLegendary CollectableValue = "S"

	FormatJpg CollectableFormat = "image"
	TypeVideo CollectableFormat = "video"
)

type CollectableValue string

var values map[CollectableValue]int = map[CollectableValue]int{ValueCommon: 5, ValueRare: 15, ValueUltraRare: 50, ValueLegendary: 200}

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
	CollectableID uint   `gorm:"foreignKey"`
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
