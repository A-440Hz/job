package collection

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) AssignBoxes(userID string, lootboxDelta int) (*UserInventory, error) {
	repoInv, err := s.repo.LookupUserInventory(userID)
	if err != nil {
		return nil, err
	}
	repoInv.NumLootboxes = max(0, repoInv.NumLootboxes+lootboxDelta)
	_, err = s.repo.UpdateUserInventoryFields(repoInv, []string{numLootboxesField})
	if err != nil {
		return nil, err
	}
	return s.repo.LookupUserInventory(userID)
}

func (s *Service) CreateNewUserInventory(userID string) (*UserInventory, error) {
	i := &UserInventory{UserID: userID}
	return s.repo.CreateUserInventory(i)
}

func (s *Service) UpdateUserInventoryFields(userID string, fields *UserInventoryUpdateFields) (*UserInventory, error) {
	_, err := s.repo.LookupUserInventory(userID)
	if err != nil {
		return nil, err
	}
	updateInventory, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	_, err = s.repo.UpdateUserInventoryFields(updateInventory, updateFields)
	if err != nil {
		return nil, err
	}
	return s.repo.LookupUserInventory(userID)
}

func (s *Service) LookupUserInventory(userID string) (*UserInventory, error) {
	return s.repo.LookupUserInventory(userID)
}

func (s *Service) DeleteUserInventory(i *UserInventory) error {
	return s.repo.DeleteUserInventory(i)
}

func (s *Service) UpdateUserCollectableFields(userID string, collecatbleID int, fields *UserCollectableUpdateFields) (*UserCollectable, error) {
	updateCol, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	updateCol.UserID, updateCol.CollectableID = userID, collecatbleID
	_, err = s.repo.UpdateUserCollectableFields(updateCol, updateFields)
	if err != nil {
		return nil, err
	}
	return s.repo.LookupUserCollectable(userID, collecatbleID)
}
