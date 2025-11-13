package collection

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	repo            *Repository
	AllCollectables []Collectable // cache the immutable collectables
}

func NewService(r *Repository) *Service {
	s := &Service{repo: r}
	err := s.repo.importCollectables()
	if err != nil {
		panic(err)
	}
	s.AllCollectables, err = s.repo.selectAllCollectables()
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Service) AssignBoxes(userID string, lootboxDelta int) (*UserInventory, error) {
	repoInv, err := s.repo.lookupUserInventory(userID)
	if err != nil {
		return nil, err
	}
	if lootboxDelta == 0 {
		return repoInv, nil
	}
	repoInv.NumLootboxes = max(0, repoInv.NumLootboxes+lootboxDelta)
	_, err = s.repo.updateUserInventoryFields(repoInv, []string{numLootboxesField})
	if err != nil {
		return nil, err
	}
	return s.repo.lookupUserInventory(userID)
}

func (s *Service) CreateNewUserInventory(userID string) (*UserInventory, error) {
	i := &UserInventory{UserID: userID}
	return s.repo.createUserInventory(i)
}

func (s *Service) UpdateUserInventoryFields(userID string, fields *UserInventoryUpdateFields) (*UserInventory, error) {
	_, err := s.repo.lookupUserInventory(userID)
	if err != nil {
		return nil, err
	}
	updateInventory, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	updateInventory.UserID = userID
	_, err = s.repo.updateUserInventoryFields(updateInventory, updateFields)
	if err != nil {
		return nil, err
	}
	return s.repo.lookupUserInventory(userID)
}

func (s *Service) LookupUserInventory(userID string) (*UserInventory, error) {
	return s.repo.lookupUserInventory(userID)
}

func (s *Service) DeleteUserInventory(i *UserInventory) error {
	return s.repo.deleteUserInventory(i)
}

func (s *Service) AwardOneRandomCollectable(userID string) (*UserCollectable, error) {
	ui, err := s.LookupUserInventory(userID)
	if err != nil {
		return nil, err
	}
	if ui.NumLootboxes < 1 {
		return nil, errors.New("user has no lootboxes to open")
	}
	// make them 1-indexed like pokemon
	cID := rand.Intn(s.repo.size) + 1
	repoColl, err := s.repo.lookupUserCollectable(userID, cID)
	if err == gorm.ErrRecordNotFound {
		errs := []error{}
		// try create new collectable in repo
		_, err := s.repo.createUserCollectable(&UserCollectable{
			UserID:        userID,
			CollectableID: cID,
			Quantity:      1,
			IsNew:         true,
			EarnedAt:      time.Now(),
		})
		errs = append(errs, err)
		// try decrement user lootbox count
		oneLess := ui.NumLootboxes - 1
		_, err = s.UpdateUserInventoryFields(userID, &UserInventoryUpdateFields{NumLootboxes: &oneLess})
		errs = append(errs, err)
		// do a lookup so it returns with preloaded collectable
		col, err := s.repo.lookupUserCollectable(userID, cID)
		errs = append(errs, err)
		return col, errors.Join(errs...)
	} else if err != nil {
		return nil, err
	}
	repoColl.IsNew = true
	repoColl.Quantity += 1

	errs := []error{}
	// try update new collectable fields in repo
	_, err = s.repo.updateUserCollectableFields(repoColl, []string{isNewField, quantityField})
	errs = append(errs, err)
	// try decrement user lootbox count
	oneLess := ui.NumLootboxes - 1
	_, err = s.UpdateUserInventoryFields(userID, &UserInventoryUpdateFields{NumLootboxes: &oneLess})
	errs = append(errs, err)
	// try lookup of new collectable from repo
	col, err := s.repo.lookupUserCollectable(userID, cID)
	errs = append(errs, err)
	return col, errors.Join(errs...)
}

func (s *Service) AwardTenRandomCollectables(userID string) ([]*UserCollectable, error) {
	ui, err := s.LookupUserInventory(userID)
	if err != nil {
		return nil, err
	}
	if ui.NumLootboxes < 10 {
		return nil, errors.New("user not enough lootboxes to open")
	}
	// cancel all db transactions if any of them fail
	tx := s.repo.beginCollectablesTransaction()
	defer func() {
		// rollback transactions on panic
		if r := recover(); r != nil {
			log.Print("recovering transaction due to panic")
			tx.Rollback()
		}
	}()
	var collectables []*UserCollectable
	for range 10 {
		coll, err := s.AwardOneRandomCollectable(userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		collectables = append(collectables, coll)
	}
	err = tx.Commit().Error
	if err != nil {
		log.Print("failed to commit transaction")
		return nil, err
	}
	return collectables, nil
}

func (s *Service) UpdateUserCollectableFields(userID string, collecatbleID int, fields *UserCollectableUpdateFields) (*UserCollectable, error) {
	updateCol, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	updateCol.UserID, updateCol.CollectableID = userID, collecatbleID
	_, err = s.repo.updateUserCollectableFields(updateCol, updateFields)
	if err != nil {
		return nil, err
	}
	return s.repo.lookupUserCollectable(userID, collecatbleID)
}

func (s *Service) GetAllCollectablesForUser(userID string) ([]UserCollectable, error) {
	return s.repo.getAllCollectablesForUser(userID)
}

func (s *Service) DeleteUserCollectables(userID string) error {
	return s.repo.deleteAllUserCollectables(userID)
}

func (s *Service) SelectAllCollectables() ([]Collectable, error) {
	return s.AllCollectables, nil
}
