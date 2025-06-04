package collection

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	s := &Service{repo: r}
	err := s.repo.importCollectables()
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

func (s *Service) SelectAllCollectables() ([]Collectable, error) {
	return s.repo.selectAllCollectables()
}
