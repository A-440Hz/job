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
