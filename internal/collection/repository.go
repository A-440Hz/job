package collection

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"gorm.io/gorm"
)

const seedFile = "data/collectables.json"

type Repository struct {
	db   *gorm.DB
	size int
}

func NewRepository(d *gorm.DB) *Repository {
	return &Repository{db: d}
}

// importCollectables deletes the collectables table and re-imports it from the json seedFile
func (r *Repository) importCollectables() error {
	r.db.Exec("TRUNCATE TABLE collectables RESTART IDENTITY CASCADE")
	f, err := os.Open(seedFile)
	defer f.Close()
	if err != nil {
		return err
	}
	var collectables []Collectable
	err = json.NewDecoder(f).Decode(&collectables)
	if err != nil {
		return err
	}
	r.size = len(collectables)
	res := r.db.Create(&collectables)
	return res.Error
}

func (u *UserInventory) BeforeCreate(tx *gorm.DB) error {
	u.LastCompleted = time.Time{} // set to zero value
	return nil
}

func (r *Repository) createUserInventory(u *UserInventory) (*UserInventory, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) lookupUserInventory(id string) (*UserInventory, error) {
	u := &UserInventory{UserID: id}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) updateUserInventoryFields(u *UserInventory, fields []string) (*UserInventory, error) {
	res := r.db.Model(u).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) deleteUserInventory(u *UserInventory) error {
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user inventory not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) createUserCollectable(u *UserCollectable) (*UserCollectable, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) updateUserCollectableFields(u *UserCollectable, fields []string) (*UserCollectable, error) {
	res := r.db.Model(u).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) lookupUserCollectable(userId string, collectableID int) (*UserCollectable, error) {
	u := &UserCollectable{}
	res := r.db.Where("user_id = ?", userId).
		Where("collectable_id = ?", collectableID).
		Preload("Collectable").
		First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) deleteUserCollectable(u *UserCollectable) error {
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user collectable not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) selectAllCollectables() ([]Collectable, error) {
	collectables := []Collectable{}
	res := r.db.Find(&collectables)
	if res.Error != nil {
		return nil, res.Error
	}
	return collectables, nil
}
