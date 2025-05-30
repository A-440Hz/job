package collection

import (
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(d *gorm.DB) *Repository {
	return &Repository{db: d}
}

func (r *Repository) CreateUserInventory(u *UserInventory) (*UserInventory, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) LookupUserInventory(id string) (*UserInventory, error) {
	u := &UserInventory{UserID: id}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) UpdateUserInventoryFields(u *UserInventory, fields []string) (*UserInventory, error) {
	res := r.db.Model(u).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) DeleteUserInventory(u *UserInventory) error {
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user inventory not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) CreateUserCollectable(u *UserCollectable) (*UserCollectable, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) UpdateUserCollectableFields(u *UserCollectable, fields []string) (*UserCollectable, error) {
	res := r.db.Model(u).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) LookupUserCollectable(userId string, collectableID int) (*UserCollectable, error) {
	u := &UserCollectable{UserID: userId, CollectableID: collectableID}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) DeleteUserCollectable(u *UserCollectable) error {
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user collectable not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}
