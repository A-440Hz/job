package tracker

import (
	"job/internal/collection"
	"job/internal/db"
	"job/internal/scheduler"
	"job/internal/user"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func init() {
	// https://intellij-support.jetbrains.com/hc/en-us/community/posts/360009685279-Go-test-working-directory-keeps-changing-to-dir-of-the-test-file-instead-of-value-in-template
	// it's pretty cool that you get to do this in golang to get releative paths to work
	db.ChdirGoTests()
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func itemStatusPtr(s string) *ItemStatus {
	status := ItemStatus(s)
	return &status
}

func Test_formatForRepo(t *testing.T) {
	n := time.Now().Unix()
	tests := []struct {
		name             string
		CycleDeadline    *int64
		CycleFrequency   *string
		GoalQuantity     *int
		CurScorableItems *int
		CurBoxesAwarded  *int
		wantErrMsg       []string
	}{
		{
			name:             "valid-fields",
			CycleDeadline:    &n,
			GoalQuantity:     intPtr(1),
			CycleFrequency:   strPtr(string(scheduler.DefaultFreq)),
			CurScorableItems: intPtr(2),
			CurBoxesAwarded:  intPtr(3),
			wantErrMsg:       nil,
		},
		{
			name:           "invalid-frequency",
			CycleFrequency: strPtr("1234671"),
			wantErrMsg:     []string{"invalid goal frequency"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateFields := &UnderlyingTrackerUpdateFields{
				CycleDeadline:    tt.CycleDeadline,
				CycleFrequency:   tt.CycleFrequency,
				GoalQuantity:     tt.GoalQuantity,
				CurScorableItems: tt.CurScorableItems,
				CurBoxesAwarded:  tt.CurBoxesAwarded,
			}

			underlyingTracker, fields, err := updateFields.formatForRepo()
			if len(tt.wantErrMsg) > 0 {
				for _, msg := range tt.wantErrMsg {
					assert.Contains(t, err.Error(), msg)
				}
				assert.Empty(t, underlyingTracker)
				assert.Empty(t, fields)
				return
			} else {
				require.NoError(t, err)
			}
			jobAppFields := &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: *updateFields}
			jaT, jaFields, err := jobAppFields.formatForRepo()
			assert.NoError(t, err)
			assert.ElementsMatch(t, jaFields, fields)
			assert.Equal(t, jaT.UnderlyingTracker, *underlyingTracker)

			// validate fields
			assert.Equal(t, *tt.CycleDeadline, underlyingTracker.CycleDeadline.Unix())
			assert.Equal(t, *tt.CycleFrequency, string(underlyingTracker.CycleFrequency))
			assert.Equal(t, *tt.GoalQuantity, underlyingTracker.GoalQuantity)
			assert.Equal(t, *tt.CurScorableItems, underlyingTracker.CurScorableItems)
			assert.Equal(t, *tt.CurBoxesAwarded, underlyingTracker.CurBoxesAwarded)
		})
	}
}

func Test_UpdateJobAppTrackerFields(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, &JobAppTracker{})
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler(), collection.NewService(collection.NewRepository(dBase)))

	n := time.Now().Round(time.Hour).Add(time.Hour).Unix()
	tests := []struct {
		name             string
		CycleDeadline    *int64
		CycleFrequency   *string
		GoalQuantity     *int
		CurScorableItems *int
		CurBoxesAwarded  *int
		wantErrMsg       []string
	}{
		{
			name:             "valid-fields",
			CycleDeadline:    &n,
			GoalQuantity:     intPtr(1),
			CycleFrequency:   strPtr(string(scheduler.DefaultFreq)),
			CurScorableItems: intPtr(0),
			CurBoxesAwarded:  intPtr(3),
			wantErrMsg:       nil,
		},
		{
			name:       "no-fields",
			wantErrMsg: []string{"no fields to update"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// a constant tracker to test against
			dummyUser := user.User{ID: "usr_12345678"}
			t1, err := svc.CreateNewJobAppTracker(&dummyUser)
			defer db.CleanDB(*dBase, &JobAppTracker{})

			require.NoError(t, err)
			assert.NotNil(t, t1)

			updateFields := &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
				CycleDeadline:    tt.CycleDeadline,
				CycleFrequency:   tt.CycleFrequency,
				GoalQuantity:     tt.GoalQuantity,
				CurScorableItems: tt.CurScorableItems,
				CurBoxesAwarded:  tt.CurBoxesAwarded,
			}}
			updateTracker, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(), updateFields)
			if len(tt.wantErrMsg) > 0 {
				for _, msg := range tt.wantErrMsg {
					assert.Contains(t, err.Error(), msg)
				}
				assert.Empty(t, updateTracker)
				return
			} else {
				require.NoError(t, err)
				require.NotNil(t, updateTracker)
			}
			assert.Equal(t, *tt.CycleDeadline, updateTracker.CycleDeadline.Unix())
			assert.Equal(t, *tt.CycleFrequency, string(updateTracker.CycleFrequency))
			assert.Equal(t, *tt.GoalQuantity, updateTracker.GoalQuantity)
			assert.Equal(t, *tt.CurScorableItems, updateTracker.CurScorableItems)
			assert.Equal(t, *tt.CurBoxesAwarded, updateTracker.CurBoxesAwarded)

			// test subsequent update -- happy path
			newQuantity := *tt.GoalQuantity + 50
			updateTracker2, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(),
				&JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
					GoalQuantity:   &newQuantity,
					CycleFrequency: strPtr("daily"),
				}})
			require.NoError(t, err)
			require.NotNil(t, updateTracker2)
			assert.Equal(t, newQuantity, updateTracker2.GoalQuantity)
			assert.Equal(t, scheduler.FreqDaily, updateTracker2.CycleFrequency)

			// test scheduler reassignment
			newDeadline := time.Unix(*tt.CycleDeadline, 0).Add(time.Hour)
			moreItems := *tt.CurScorableItems + 10
			ndu := newDeadline.Unix()
			updateTracker2_5, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(),
				&JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
					CycleDeadline:    &ndu,
					CycleFrequency:   strPtr("weekly"),
					CurScorableItems: &moreItems,
					GoalQuantity:     &newQuantity,
				}})
			require.NotNil(t, updateTracker2_5)
			require.NoError(t, err)
			assert.Equal(t, newDeadline, updateTracker2_5.CycleDeadline)
			assert.Equal(t, scheduler.FreqWeekly, updateTracker2_5.CycleFrequency)
			assert.Equal(t, moreItems, updateTracker2_5.CurScorableItems)
			assert.Equal(t, newQuantity, updateTracker2_5.GoalQuantity)

		})
	}
}

func Test_CreateJobAppItem(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
	repo := NewRepository(dBase)
	cs := collection.NewService(collection.NewRepository(dBase))
	svc := NewService(repo, scheduler.NewScheduler(), cs)
	uSvc := user.NewService(user.NewRepository(dBase), cs)

	defaultStatus := StatusComplete
	n := time.Now().Round(time.Hour)
	tests := []struct {
		name            string
		Title           *string
		Body            *string
		Status          *ItemStatus
		IsAttributed    *bool
		AttributionTime *time.Time
		wantErrMsg      []string
	}{
		{
			name:         "valid-fields",
			Title:        strPtr("Title"),
			Body:         strPtr("Body"),
			Status:       &defaultStatus,
			IsAttributed: boolPtr(false),
		},
		{
			name:       "missing-fields",
			wantErrMsg: []string{"missing item name", "missing item status", "missing isAttributed"},
		},
		{
			name:         "invalid-status",
			Title:        strPtr("Title"),
			Body:         strPtr("Body"),
			Status:       itemStatusPtr("something-invalid"),
			IsAttributed: boolPtr(false),
			wantErrMsg:   []string{"invalid application status"},
		},
		{
			name:            "with-attr-time",
			Title:           strPtr("Title"),
			Body:            strPtr("Body"),
			Status:          &defaultStatus,
			IsAttributed:    boolPtr(true),
			AttributionTime: &n,
			wantErrMsg:      []string{"cannot create an attributed item"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyUser, err := uSvc.CreateNewUser(scheduler.GetDefaultTimezone())
			require.NoError(t, err)
			assert.NotNil(t, dummyUser)
			t1, err := svc.CreateNewJobAppTracker(dummyUser)
			require.NoError(t, err)
			assert.NotNil(t, t1)

			t1, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
				Title:           tt.Title,
				Body:            tt.Body,
				Status:          tt.Status,
				IsAttributed:    tt.IsAttributed,
				attributionTime: tt.AttributionTime,
			})

			if len(tt.wantErrMsg) > 0 {
				for _, msg := range tt.wantErrMsg {
					assert.ErrorContains(t, err, msg)
				}
				assert.Nil(t, t1)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, t1)

			t1, err = svc.GetJobAppTrackerWithItemsFromUserID(dummyUser.GetID())
			require.NoError(t, err)

			assert.Equal(t, *tt.Title, t1.Items[0].Title)
			assert.Equal(t, *tt.Body, t1.Items[0].Body)
			assert.Equal(t, *tt.Status, t1.Items[0].Status)
			assert.Equal(t, *tt.IsAttributed, t1.Items[0].IsAttributed)
			assert.Equal(t, tt.AttributionTime, t1.Items[0].AttributionTime)
			dBase.Exec("TRUNCATE TABLE job_app_trackers, job_app_items, users, user_inventories RESTART IDENTITY CASCADE")
		})
	}

}

func Test_UpdateJobAppItemFields(t *testing.T) {
	// test updating the status complete field and seeing if tracker curItemsComplete decrements by 1
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, &user.User{})
	cSvc := collection.NewService(collection.NewRepository(dBase))
	svc := NewService(NewRepository(dBase), scheduler.NewScheduler(), cSvc)
	uSvc := user.NewService(user.NewRepository(dBase), cSvc)

	dummyUser, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	tracker, err := svc.CreateNewJobAppTracker(dummyUser)
	require.NoError(t, err)
	require.NotNil(t, tracker)

	// Set GoalQuantity to 2 for easier testing
	tracker, err = svc.UpdateJobAppTrackerFields(dummyUser.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			GoalQuantity: intPtr(2),
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, tracker.GoalQuantity)

	statusComplete := StatusComplete
	statusInProgress := StatusInProgress

	// Create one completed item and ensure CurScorableItems is 1
	tracker, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Item 1"),
		Body:         strPtr("Body 1"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	item1 := tracker.Items[0]

	tracker, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Item 2"),
		Body:         strPtr("Body 2"),
		Status:       &statusInProgress,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)

	// At this point, CurScorableItems should be 0, CurBoxesAwarded should be 1
	assert.Equal(t, 1, tracker.CurScorableItems)
	assert.Equal(t, 0, tracker.CurBoxesAwarded)

	var item2 JobAppItem
	for _, i := range tracker.Items {
		if i == item1 {
			continue
		}
		item2 = i
	}

	// Update item1 status from complete to in-progress and check that CurScorableItems has decremented
	tracker, err = svc.UpdateJobAppItemFields(dummyUser.GetID(), item1.ID, &JobAppItemUpdateFields{
		Status: &statusInProgress,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, tracker.CurScorableItems)
	assert.Equal(t, 0, tracker.CurBoxesAwarded)

	// Update both items to complete and check for CurBoxesAwarded
	tracker, err = svc.UpdateJobAppItemFields(dummyUser.GetID(), item1.ID, &JobAppItemUpdateFields{
		Status: &statusComplete,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, tracker.CurScorableItems)
	assert.Equal(t, 0, tracker.CurBoxesAwarded)
	tracker, err = svc.UpdateJobAppItemFields(dummyUser.GetID(), item2.ID, &JobAppItemUpdateFields{
		Status: &statusComplete,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, tracker.CurScorableItems)
	assert.Equal(t, 1, tracker.CurBoxesAwarded)

	dBase.Exec("TRUNCATE TABLE job_app_trackers, job_app_items, users RESTART IDENTITY CASCADE")
}

// Test_Scheduler_AdvanceDeadline creates 4 tracker deadlines from now+1s and verifies the scheduler correctly cycles them in the brackground
func Test_Scheduler_AdvanceDeadline(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, &JobAppTracker{}, &JobAppItem{})
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler(), collection.NewService(collection.NewRepository(dBase)))

	t0 := time.Now().Round(time.Second).Add(time.Second)
	t1 := t0.Add(time.Second * 1)
	t2 := t1.Add(time.Second * 1)
	t3 := t2.Add(time.Second * 1)

	u0 := user.User{ID: "usr_00000000"}
	u1 := user.User{ID: "usr_11111111"}
	u2 := user.User{ID: "usr_22222222"}
	u3 := user.User{ID: "usr_33333333"}

	svc.Start()

	jt0, err := svc.CreateNewJobAppTracker(&u0)
	require.NoError(t, err)
	jt1, err := svc.CreateNewJobAppTracker(&u1)
	require.NoError(t, err)
	jt2, err := svc.CreateNewJobAppTracker(&u2)
	require.NoError(t, err)
	jt3, err := svc.CreateNewJobAppTracker(&u3)
	require.NoError(t, err)
	jt00, err := svc.UpdateJobAppTrackerFields(u0.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CycleDeadline:  int64Ptr(t0.Unix()),
			CycleFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt0, jt00)
	jt11, err := svc.UpdateJobAppTrackerFields(u1.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CycleDeadline:  int64Ptr(t1.Unix()),
			CycleFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt1, jt11)
	jt22, err := svc.UpdateJobAppTrackerFields(u2.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CycleDeadline:  int64Ptr(t2.Unix()),
			CycleFrequency: strPtr("weekly"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt2, jt22)
	jt33, err := svc.UpdateJobAppTrackerFields(u3.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CycleDeadline:  int64Ptr(t3.Unix()),
			CycleFrequency: strPtr("weekly"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt3, jt33)
	time.Sleep(time.Second * 6)

	jt000, err := svc.LookupJobAppTrackerFromUserID(u0.GetID())
	require.NoError(t, err)
	jt111, err := svc.LookupJobAppTrackerFromUserID(u1.GetID())
	require.NoError(t, err)
	jt222, err := svc.LookupJobAppTrackerFromUserID(u2.GetID())
	require.NoError(t, err)
	jt333, err := svc.LookupJobAppTrackerFromUserID(u3.GetID())
	require.NoError(t, err)
	assert.Equal(t, t0.AddDate(0, 0, 1), jt000.CycleDeadline)
	assert.Equal(t, t1.AddDate(0, 0, 1), jt111.CycleDeadline)
	assert.Equal(t, t2.AddDate(0, 0, 7), jt222.CycleDeadline)
	assert.Equal(t, t3.AddDate(0, 0, 7), jt333.CycleDeadline)
	// assert.NotEqual(t, jt000.CycleDeadline, jt00.CycleDeadline)
	// assert.NotEqual(t, jt111.CycleDeadline, jt11.CycleDeadline)
	// assert.NotEqual(t, jt222.CycleDeadline, jt22.CycleDeadline)
	// assert.NotEqual(t, jt333.CycleDeadline, jt33.CycleDeadline)
	log.Printf("after: %v, before: %v", jt000.CycleDeadline, jt00.CycleDeadline)
	log.Printf("after: %v, before: %v", jt111.CycleDeadline, jt11.CycleDeadline)
	log.Printf("after: %v, before: %v", jt222.CycleDeadline, jt22.CycleDeadline)
	log.Printf("after: %v, before: %v", jt333.CycleDeadline, jt33.CycleDeadline)
	svc.scheduler.Stop()
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
}

func Test_JobAppTrackerItemScoring(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, &user.User{}, &collection.UserInventory{})
	cSvc := collection.NewService(collection.NewRepository(dBase))
	svc := NewService(NewRepository(dBase), scheduler.NewScheduler(), cSvc)
	uSvc := user.NewService(user.NewRepository(dBase), cSvc)

	dummyUser, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	t1, err := svc.CreateNewJobAppTracker(dummyUser)
	require.NoError(t, err)
	require.NotNil(t, t1)

	t1, err = svc.UpdateJobAppTrackerFields(dummyUser.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields{GoalQuantity: intPtr(1)}})
	require.NoError(t, err)
	require.Equal(t, 1, t1.GoalQuantity)

	statusComplete := StatusComplete
	statusInProgress := StatusInProgress
	t1, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Title"),
		Body:         strPtr("Body"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	assert.Equal(t, 0, t1.CurScorableItems)
	assert.Equal(t, 1, t1.CurBoxesAwarded)
	// assert.Equal(t, 1, t1.MaxGoalStreak)

	t1, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Title"),
		Body:         strPtr("Body"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	assert.Equal(t, 0, t1.CurScorableItems)
	assert.Equal(t, 2, t1.CurBoxesAwarded)

	// change the GoalQuantity to 2 and add more items
	// note how it will not retroactively affect the boxes already awarded
	t1, err = svc.UpdateJobAppTrackerFields(dummyUser.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields{GoalQuantity: intPtr(2)}})
	require.NoError(t, err)
	require.Equal(t, 2, t1.GoalQuantity)
	assert.Equal(t, 2, t1.CurBoxesAwarded)
	assert.Equal(t, 0, t1.CurScorableItems)

	t1, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Title"),
		Body:         strPtr("Body"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, t1.CurScorableItems)
	assert.Equal(t, 2, t1.CurBoxesAwarded)

	// test counters stay consistent when CreateJobAppItem fails &..
	// test counters do not increment when item is in progress
	t1, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Title"),
		Body:         strPtr("Body"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(true),
	})
	require.Error(t, err)
	require.Nil(t, t1)
	t1, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Title"),
		Body:         strPtr("Body"),
		Status:       &statusInProgress,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, t1.CurScorableItems)
	assert.Equal(t, 2, t1.CurBoxesAwarded)

	// test counters work as expected despite previous failed calls
	t1, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Title"),
		Body:         strPtr("Body"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	assert.Equal(t, 0, t1.CurScorableItems)
	assert.Equal(t, 3, t1.CurBoxesAwarded)

	// test behavior when decreasing GoalQuantity to a lower number than current items
	dummyUser2, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	t2, err := svc.CreateNewJobAppTracker(dummyUser2)
	require.NoError(t, err)
	require.NotNil(t, t2)
	for i := 0; i < 4; i++ {
		t2, err = svc.CreateJobAppItem(dummyUser2.GetID(), &JobAppItemUpdateFields{
			Title:        strPtr("Title"),
			Body:         strPtr("Body"),
			Status:       &statusComplete,
			IsAttributed: boolPtr(false),
		})
		require.NoError(t, err)
	}
	assert.Equal(t, 4, t2.CurScorableItems)
	assert.Equal(t, 0, t2.CurBoxesAwarded)

	t2, err = svc.UpdateJobAppTrackerFields(dummyUser2.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields{GoalQuantity: intPtr(2)}})
	require.NoError(t, err)
	require.Equal(t, 2, t2.GoalQuantity)
	assert.Equal(t, 0, t2.CurScorableItems)
	assert.Equal(t, 2, t2.CurBoxesAwarded)

	db.CleanDB(*dBase, &JobAppTracker{}, &JobAppItem{}, &user.User{}, &collection.UserInventory{})
}

func Test_RestoreJobAppTrackerItem_Repository(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
	repo := NewRepository(dBase)
	cs := collection.NewService(collection.NewRepository(dBase))
	svc := NewService(repo, scheduler.NewScheduler(), cs)
	uSvc := user.NewService(user.NewRepository(dBase), cs)

	// Create a user and tracker
	dummyUser, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	tracker, err := svc.CreateNewJobAppTracker(dummyUser)
	require.NoError(t, err)

	statusComplete := StatusComplete

	// Create an item
	tracker, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Test Item"),
		Body:         strPtr("Test Body"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	require.Len(t, tracker.Items, 1)
	itemID := tracker.Items[0].ID

	t.Run("restore-item-that-is-not-deleted", func(t *testing.T) {
		// Attempt to restore an item that is not deleted
		_, err := repo.restoreJobAppTrackerItem(itemID, tracker.GetID())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "item is not deleted")
	})

	// Delete the item
	err = svc.DeleteJobAppItem(dummyUser.GetID(), itemID)
	require.NoError(t, err)

	// Verify item is soft-deleted
	var deletedItem JobAppItem
	err = dBase.Unscoped().Where("id = ?", itemID).First(&deletedItem).Error
	require.NoError(t, err)
	assert.True(t, deletedItem.DeletedAt.Valid, "Item should be soft-deleted")

	t.Run("restore-deleted-item-successfully", func(t *testing.T) {
		// Restore the deleted item
		restoredItem, err := repo.restoreJobAppTrackerItem(itemID, tracker.GetID())
		require.NoError(t, err)
		require.NotNil(t, restoredItem)
		assert.Equal(t, itemID, restoredItem.ID)
		assert.False(t, restoredItem.DeletedAt.Valid, "Item should no longer be deleted")

		// Verify item is now visible in normal queries
		var activeItem JobAppItem
		err = dBase.Where("id = ?", itemID).First(&activeItem).Error
		require.NoError(t, err)
		assert.Equal(t, itemID, activeItem.ID)
	})

	// Delete the item again for next test
	err = repo.deleteJobAppTrackerItem(&JobAppItem{ID: itemID})
	require.NoError(t, err)

	t.Run("restore-nonexistent-item", func(t *testing.T) {
		// Attempt to restore an item that doesn't exist
		_, err := repo.restoreJobAppTrackerItem("item_nonexistent", tracker.GetID())
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
}

func Test_RestoreJobAppItem_Service(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
	repo := NewRepository(dBase)
	cs := collection.NewService(collection.NewRepository(dBase))
	svc := NewService(repo, scheduler.NewScheduler(), cs)
	uSvc := user.NewService(user.NewRepository(dBase), cs)

	// Create two users and trackers
	user1, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	tracker1, err := svc.CreateNewJobAppTracker(user1)
	require.NoError(t, err)

	user2, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	tracker2, err := svc.CreateNewJobAppTracker(user2)
	require.NoError(t, err)

	statusComplete := StatusComplete

	// Create item for user1
	tracker1, err = svc.CreateJobAppItem(user1.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("User1 Item"),
		Body:         strPtr("Body1"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	require.Len(t, tracker1.Items, 1)
	item1ID := tracker1.Items[0].ID

	// Create item for user2
	tracker2, err = svc.CreateJobAppItem(user2.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("User2 Item"),
		Body:         strPtr("Body2"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	require.Len(t, tracker2.Items, 1)
	// item2ID := tracker2.Items[0].ID

	t.Run("restore-item-belonging-to-different-tracker", func(t *testing.T) {
		// Delete item1
		err := svc.DeleteJobAppItem(user1.GetID(), item1ID)
		require.NoError(t, err)

		// Attempt to restore item1 using user2's tracker
		_, err = svc.RestoreJobAppItem(user2.GetID(), item1ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "item does not belong to this tracker")

		// Restore using correct user
		restoredTracker, err := svc.RestoreJobAppItem(user1.GetID(), item1ID)
		require.NoError(t, err)
		require.NotNil(t, restoredTracker)
		assert.Len(t, restoredTracker.Items, 1)
		assert.Equal(t, item1ID, restoredTracker.Items[0].ID)
	})

	t.Run("restore-updates-scorable-items-count", func(t *testing.T) {
		// Set goal quantity to 3 to prevent auto-scoring
		tracker1, err := svc.UpdateJobAppTrackerFields(user1.GetID(), &JobAppTrackerUpdateFields{
			UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
				GoalQuantity: intPtr(3),
			},
		})
		require.NoError(t, err)
		require.Equal(t, 1, tracker1.CurScorableItems)

		// Create another item
		tracker1, err = svc.CreateJobAppItem(user1.GetID(), &JobAppItemUpdateFields{
			Title:        strPtr("Second Item"),
			Body:         strPtr("Body"),
			Status:       &statusComplete,
			IsAttributed: boolPtr(false),
		})
		require.NoError(t, err)
		require.Equal(t, 2, tracker1.CurScorableItems)
		anItemID := tracker1.Items[0].ID

		// Delete the second item
		err = svc.DeleteJobAppItem(user1.GetID(), anItemID)
		require.NoError(t, err)

		// Check that CurScorableItems decremented
		tracker1, err = svc.GetJobAppTrackerWithItemsFromUserID(user1.GetID())
		require.NoError(t, err)
		assert.Equal(t, 1, tracker1.CurScorableItems)

		// Restore the item
		tracker1, err = svc.RestoreJobAppItem(user1.GetID(), anItemID)
		require.NoError(t, err)
		assert.Equal(t, 2, tracker1.CurScorableItems, "CurScorableItems should increment on restore")
		assert.Len(t, tracker1.Items, 2)

		// Verify the restored item is visible
		var restoredItem *JobAppItem
		for i := range tracker1.Items {
			if tracker1.Items[i].ID == anItemID {
				restoredItem = &tracker1.Items[i]
				break
			}
		}
		require.NotNil(t, restoredItem, "Restored item should be in tracker items")
		assert.Equal(t, "Second Item", restoredItem.Title)
	})

	t.Run("restore-non-scorable-item", func(t *testing.T) {
		statusInProgress := StatusInProgress

		// Create an in-progress (non-scorable) item
		tracker2, err := svc.CreateJobAppItem(user2.GetID(), &JobAppItemUpdateFields{
			Title:        strPtr("In Progress Item"),
			Body:         strPtr("Body"),
			Status:       &statusInProgress,
			IsAttributed: boolPtr(false),
		})
		require.NoError(t, err)
		inProgressItemID := tracker2.Items[0].ID
		initialScorableCount := tracker2.CurScorableItems

		// Delete the in-progress item
		err = svc.DeleteJobAppItem(user2.GetID(), inProgressItemID)
		require.NoError(t, err)

		// Verify scorable count didn't change (item wasn't scorable)
		tracker2, err = svc.GetJobAppTrackerWithItemsFromUserID(user2.GetID())
		require.NoError(t, err)
		assert.Equal(t, initialScorableCount, tracker2.CurScorableItems)

		// Restore the item
		tracker2, err = svc.RestoreJobAppItem(user2.GetID(), inProgressItemID)
		require.NoError(t, err)
		assert.Equal(t, initialScorableCount, tracker2.CurScorableItems, "Non-scorable item restore should not change count")
	})

	t.Run("restore-attributed-item", func(t *testing.T) {
		// Set goal quantity to 1 so all items get attributed
		tracker1, err := svc.UpdateJobAppTrackerFields(user1.GetID(), &JobAppTrackerUpdateFields{
			UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
				GoalQuantity: intPtr(1),
			},
		})
		require.NoError(t, err)

		// Current scorable count should be 0
		assert.Equal(t, 0, tracker1.CurScorableItems)

		// Create item - should be auto-attributed since we already have enough items
		tracker1, err = svc.CreateJobAppItem(user1.GetID(), &JobAppItemUpdateFields{
			Title:        strPtr("Auto Attributed Item"),
			Body:         strPtr("Body"),
			Status:       &statusComplete,
			IsAttributed: boolPtr(false),
		})
		require.NoError(t, err)

		// Find the attributed item
		var attributedItem *JobAppItem
		for _, item := range tracker1.Items {
			if item.Title == "Auto Attributed Item" {
				attributedItem = &item
				break
			}
		}
		require.NotNil(t, attributedItem)
		assert.True(t, attributedItem.IsAttributed)

		// Delete the attributed item
		err = svc.DeleteJobAppItem(user1.GetID(), attributedItem.ID)
		require.NoError(t, err)

		// Restore it after incrementing goal quantity - should not increment CurScorableItems since it's attributed
		tracker1, err = svc.UpdateJobAppTrackerFields(user1.GetID(), &JobAppTrackerUpdateFields{
			UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
				GoalQuantity: intPtr(2),
			},
		})
		require.NoError(t, err)
		tracker1, err = svc.RestoreJobAppItem(user1.GetID(), attributedItem.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, tracker1.CurScorableItems, "Attributed item restore should not change scorable count")
	})

	t.Run("restore-item-that-doesnt-exist", func(t *testing.T) {
		_, err := svc.RestoreJobAppItem(user1.GetID(), "item_nonexistent")
		assert.Error(t, err)
	})

	t.Run("restore-item-with-invalid-user", func(t *testing.T) {
		_, err := svc.RestoreJobAppItem("usr_invalid", item1ID)
		assert.Error(t, err)
	})

	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
}

func Test_RestoreJobAppItem_MultipleRestores(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
	repo := NewRepository(dBase)
	cs := collection.NewService(collection.NewRepository(dBase))
	svc := NewService(repo, scheduler.NewScheduler(), cs)
	uSvc := user.NewService(user.NewRepository(dBase), cs)

	dummyUser, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	tracker, err := svc.CreateNewJobAppTracker(dummyUser)
	require.NoError(t, err)

	// Set goal quantity high to prevent auto-scoring
	tracker, err = svc.UpdateJobAppTrackerFields(dummyUser.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			GoalQuantity: intPtr(10),
		},
	})
	require.NoError(t, err)

	statusComplete := StatusComplete
	itemIDs := []string{}

	// Create multiple items
	for i := 0; i < 3; i++ {
		tracker, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
			Title:        strPtr("Item " + string(rune('A'+i))),
			Body:         strPtr("Body"),
			Status:       &statusComplete,
			IsAttributed: boolPtr(false),
		})
		require.NoError(t, err)
		itemIDs = append(itemIDs, tracker.Items[0].ID)
	}

	// Verify initial state
	assert.Equal(t, 3, tracker.CurScorableItems)
	assert.Len(t, tracker.Items, 3)

	// Delete all items
	for _, itemID := range itemIDs {
		err = svc.DeleteJobAppItem(dummyUser.GetID(), itemID)
		require.NoError(t, err)
	}

	// Verify all items are deleted
	tracker, err = svc.GetJobAppTrackerWithItemsFromUserID(dummyUser.GetID())
	require.NoError(t, err)
	assert.Equal(t, 0, tracker.CurScorableItems)
	assert.Len(t, tracker.Items, 0)

	// Restore all items one by one
	for i, itemID := range itemIDs {
		tracker, err = svc.RestoreJobAppItem(dummyUser.GetID(), itemID)
		require.NoError(t, err)
		assert.Equal(t, i+1, tracker.CurScorableItems, "CurScorableItems should increment with each restore")
		assert.Len(t, tracker.Items, i+1, "Items list should grow with each restore")
	}

	// Verify final state
	assert.Equal(t, 3, tracker.CurScorableItems)
	assert.Len(t, tracker.Items, 3)

	// Try to restore an already restored item
	_, err = svc.RestoreJobAppItem(dummyUser.GetID(), itemIDs[0])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item is not deleted")

	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
}

func Test_RestoreJobAppItem_WithScoring(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
	repo := NewRepository(dBase)
	cs := collection.NewService(collection.NewRepository(dBase))
	svc := NewService(repo, scheduler.NewScheduler(), cs)
	uSvc := user.NewService(user.NewRepository(dBase), cs)

	dummyUser, err := uSvc.CreateNewUser(nil)
	require.NoError(t, err)
	tracker, err := svc.CreateNewJobAppTracker(dummyUser)
	require.NoError(t, err)

	// Set goal quantity to 2
	tracker, err = svc.UpdateJobAppTrackerFields(dummyUser.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			GoalQuantity: intPtr(2),
		},
	})
	require.NoError(t, err)

	statusComplete := StatusComplete

	// Create 2 items to hit the goal
	tracker, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Item 1"),
		Body:         strPtr("Body 1"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
	item1ID := tracker.Items[0].ID

	tracker, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Item 2"),
		Body:         strPtr("Body 2"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)

	// Verify scoring happened
	assert.Equal(t, 0, tracker.CurScorableItems, "Items should be scored")
	assert.Equal(t, 1, tracker.CurBoxesAwarded, "Should have awarded 1 box")

	// Both items should be attributed
	for _, item := range tracker.Items {
		assert.True(t, item.IsAttributed, "Items should be attributed after scoring")
	}

	// Delete item 1 (which is attributed)
	err = svc.DeleteJobAppItem(dummyUser.GetID(), item1ID)
	require.NoError(t, err)

	tracker, err = svc.GetJobAppTrackerWithItemsFromUserID(dummyUser.GetID())
	require.NoError(t, err)
	assert.Equal(t, 0, tracker.CurScorableItems)
	assert.Equal(t, 1, tracker.CurBoxesAwarded)

	// Restore item 1
	tracker, err = svc.RestoreJobAppItem(dummyUser.GetID(), item1ID)
	require.NoError(t, err)

	// Scorable items should not increment because item is attributed
	assert.Equal(t, 0, tracker.CurScorableItems, "Attributed item should not increment scorable count")
	assert.Equal(t, 1, tracker.CurBoxesAwarded, "Box count should remain the same")

	db.CleanDB(*dBase, JobAppTracker{}, JobAppItem{}, user.User{}, collection.UserInventory{})
}
