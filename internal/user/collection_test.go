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
				for range 20 {
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
				for range 3 {
					col, err := uSvc.collection.AwardOneRandomCollectable(user.GetID())
					assert.NotNil(t, col)
					assert.NoError(t, err)
				}
				for range 2 {
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

func Test_AwardTenRandomCollectables(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, &User{}, &collection.UserInventory{}, &collection.UserCollectable{}, &collection.Collectable{})
	cSvc := collection.NewService(collection.NewRepository(dBase))
	uSvc := NewService(NewRepository(dBase), cSvc)

	t.Run("award-10-boxes", func(t *testing.T) {
		user, err := uSvc.CreateNewUser(nil)
		require.NoError(t, err)

		_, err = uSvc.collection.AssignBoxes(user.GetID(), 10)
		require.NoError(t, err)

		cols, err := uSvc.collection.AwardTenRandomCollectables(user.GetID())
		require.NoError(t, err)
		require.NotNil(t, cols)
		require.Len(t, cols, 10)

		userCollectables, err := uSvc.collection.GetAllCollectablesForUser(user.GetID())
		require.NoError(t, err)

		total := 0
		for _, c := range userCollectables {
			total += c.Quantity
		}
		assert.Equal(t, 10, total)
	})

	t.Run("not-enough-boxes", func(t *testing.T) {
		user, err := uSvc.CreateNewUser(nil)
		require.NoError(t, err)

		cols, err := uSvc.collection.AwardTenRandomCollectables(user.GetID())
		assert.Nil(t, cols)
		assert.ErrorContains(t, err, "not enough lootboxes")
	})

	db.CleanDB(*dBase, &User{}, &collection.UserInventory{}, &collection.UserCollectable{})
}
