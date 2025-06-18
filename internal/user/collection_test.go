package user

import (
	"job/internal/collection"
	"job/internal/db"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	db.ChdirGoTests()
}

func Test_AwardRandomCollectable(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, &User{}, &collection.UserInventory{}, &collection.UserCollectable{}, &collection.Collectable{})
	cSvc := collection.NewService(collection.NewRepository(dBase))
	uSvc := NewService(NewRepository(dBase), cSvc)

	tests := []struct {
		name                string
		wantNumCollectables int
		ops                 func(*Service) ([]collection.UserCollectable, error)
	}{
		{
			name:                "award-20-boxes",
			wantNumCollectables: 20,
			ops: func(uSvc *Service) ([]collection.UserCollectable, error) {
				user, err := uSvc.CreateNewUser(nil)
				require.NoError(t, err)

				uSvc.collection.AssignBoxes(user.GetID(), 20)
				require.NoError(t, err)
				for _ = range 20 {
					col, err := uSvc.collection.AwardOneRandomCollectable(user.GetID())
					assert.NotNil(t, col)
					assert.NoError(t, err)
				}
				return uSvc.collection.GetAllCollectablesForUser(user.GetID())
			},
		},
		{
			name:                "no-boxes-to-open",
			wantNumCollectables: 0,
			ops: func(uSvc *Service) ([]collection.UserCollectable, error) {
				user, err := uSvc.CreateNewUser(nil)
				require.NoError(t, err)

				col, err := uSvc.collection.AwardOneRandomCollectable(user.GetID())
				assert.Nil(t, col)
				assert.ErrorContains(t, err, "no lootboxes to open")
				return uSvc.collection.GetAllCollectablesForUser(user.GetID())
			},
		},
		{
			name:                "run-out-of-boxes",
			wantNumCollectables: 3,
			ops: func(uSvc *Service) ([]collection.UserCollectable, error) {
				user, err := uSvc.CreateNewUser(nil)
				require.NoError(t, err)

				uSvc.collection.AssignBoxes(user.GetID(), 3)
				require.NoError(t, err)
				for _ = range 3 {
					col, err := uSvc.collection.AwardOneRandomCollectable(user.GetID())
					assert.NotNil(t, col)
					assert.NoError(t, err)
				}
				for _ = range 2 {
					col, err := uSvc.collection.AwardOneRandomCollectable(user.GetID())
					assert.Nil(t, col)
					assert.ErrorContains(t, err, "no lootboxes to open")
				}
				return uSvc.collection.GetAllCollectablesForUser(user.GetID())
			},
		},
	}
	for _, tt := range tests {
		userCollectables, _ := tt.ops(uSvc)

		total := 0
		for _, c := range userCollectables {
			total += c.Quantity
		}
		assert.Equal(t, total, tt.wantNumCollectables)
	}
	db.CleanDB(*dBase, &User{}, &collection.UserInventory{}, &collection.UserCollectable{})
}
