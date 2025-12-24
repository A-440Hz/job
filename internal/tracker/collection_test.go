package tracker

import (
	"job/internal/collection"
	"job/internal/db"
	"job/internal/scheduler"
	"job/internal/user"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	db.ChdirGoTests()
}

func Test_AssignBoxes(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, &JobAppTracker{}, &JobAppItem{}, &collection.UserInventory{})
	cSvc := collection.NewService(collection.NewRepository(dBase))
	tSvc := NewService(NewRepository(dBase), scheduler.NewScheduler(), cSvc)
	uSvc := user.NewService(user.NewRepository(dBase), cSvc)

	complete := StatusComplete
	newItem := &JobAppItemUpdateFields{
		Title:        strPtr("title"),
		Body:         strPtr("body"),
		Status:       &complete,
		IsAttributed: boolPtr(false),
	}

	tests := []struct {
		name         string
		wantNumBoxes int
		ops          func(*Service, *user.Service, int) (*collection.UserInventory, error)
		wantErrMsg   []string
	}{
		{
			name:         "add-20-boxes",
			wantNumBoxes: 20,
			ops: func(tSvc *Service, uSvc *user.Service, wantNumBoxes int) (*collection.UserInventory, error) {
				user, err := uSvc.CreateNewUser(nil)
				require.NoError(t, err)
				_, err = tSvc.CreateNewJobAppTracker(user)
				require.NoError(t, err)

				// set item goal to 1 for easy count
				_, err = tSvc.UpdateJobAppTrackerFields(user.GetID(), &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields{GoalQuantity: intPtr(1)}})
				require.NoError(t, err)

				// create items
				for range wantNumBoxes {
					_, err = tSvc.CreateJobAppItem(user.GetID(), newItem)
					require.NoError(t, err)
				}
				return tSvc.collection.LookupUserInventory(user.GetID())
			},
		},
		{
			name:         "add-boxes-via-goal-quantity",
			wantNumBoxes: 5,
			ops: func(tSvc *Service, uSvc *user.Service, wantNumBoxes int) (*collection.UserInventory, error) {
				user, err := uSvc.CreateNewUser(nil)
				require.NoError(t, err)
				_, err = tSvc.CreateNewJobAppTracker(user)
				require.NoError(t, err)

				// explicity set goal quantity to 5
				_, err = tSvc.UpdateJobAppTrackerFields(user.GetID(), &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields{GoalQuantity: intPtr(5)}})
				require.NoError(t, err)

				// create items; expect 1 box awarded and 4 scorable items
				for range 9 {
					_, err = tSvc.CreateJobAppItem(user.GetID(), newItem)
					require.NoError(t, err)
				}
				repoTracker, err := tSvc.LookupJobAppTrackerFromUserID(user.GetID())
				require.NoError(t, err)
				assert.Equal(t, repoTracker.CurBoxesAwarded, 1)
				assert.Equal(t, repoTracker.CurScorableItems, 4)
				ui, err := tSvc.collection.LookupUserInventory(user.GetID())
				require.NoError(t, err)
				assert.Equal(t, ui.NumLootboxes, 1)

				// change goal quantity to 1 and expect 4 more awarded boxes
				repoTracker, err = tSvc.UpdateJobAppTrackerFields(repoTracker.GetUserID(), &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields{GoalQuantity: intPtr(1)}})
				require.NoError(t, err)
				assert.Equal(t, repoTracker.CurBoxesAwarded, 5)
				assert.Equal(t, repoTracker.CurScorableItems, 0)
				ui, err = tSvc.collection.LookupUserInventory(user.GetID())
				require.NoError(t, err)
				assert.Equal(t, ui.NumLootboxes, 5)
				return ui, nil
			},
		},
		{
			name:         "add-then-delete",
			wantNumBoxes: 12,
			ops: func(tSvc *Service, uSvc *user.Service, wantNumBoxes int) (*collection.UserInventory, error) {
				user, err := uSvc.CreateNewUser(nil)
				require.NoError(t, err)
				_, err = tSvc.CreateNewJobAppTracker(user)
				require.NoError(t, err)

				// set item goal to 1 for easy count
				repoTracker, err := tSvc.UpdateJobAppTrackerFields(user.GetID(), &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields{GoalQuantity: intPtr(1)}})
				require.NoError(t, err)

				// create items
				for range wantNumBoxes {
					repoTracker, err = tSvc.CreateJobAppItem(user.GetID(), newItem)
					require.NoError(t, err)
				}
				// delete all items
				for _, i := range repoTracker.Items {
					err = tSvc.DeleteJobAppItem(user.GetID(), i.GetID())
					require.NoError(t, err)
				}
				// numLootboxes should not decrease
				return tSvc.collection.LookupUserInventory(user.GetID())
			},
		},
	}
	for _, tt := range tests {
		ui, err := tt.ops(tSvc, uSvc, tt.wantNumBoxes)
		if len(tt.wantErrMsg) > 0 {
			for _, wantMsg := range tt.wantErrMsg {
				assert.ErrorContains(t, err, wantMsg)
			}
			continue
		}
		assert.Equal(t, ui.NumLootboxes, tt.wantNumBoxes)
	}
	db.CleanDB(*dBase, &JobAppTracker{}, &JobAppItem{}, &collection.UserInventory{})
}
