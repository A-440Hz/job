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
)

func intPtr(i int) *int {
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
	n := time.Now()
	tests := []struct {
		name                   string
		CycleDeadline          *time.Time
		CycleFrequency         *string
		GoalQuantity           *int
		CurScorableItems       *int
		CurBoxesAwarded        *int
		CurGoalStreak          *int
		MaxGoalStreak          *int
		MaxCycleItemsCompleted *int
		TotalItemsCompleted    *int
		TotalBoxesAwarded      *int
		FirstCompleted         *time.Time
		wantErrMsg             []string
	}{
		{
			name:                   "valid-fields",
			CycleDeadline:          &n,
			GoalQuantity:           intPtr(1),
			CycleFrequency:         strPtr(string(scheduler.DefaultFreq)),
			CurScorableItems:       intPtr(2),
			CurBoxesAwarded:        intPtr(3),
			CurGoalStreak:          intPtr(4),
			MaxGoalStreak:          intPtr(5),
			MaxCycleItemsCompleted: intPtr(6),
			TotalItemsCompleted:    intPtr(7),
			TotalBoxesAwarded:      intPtr(8),
			FirstCompleted:         &n,
			wantErrMsg:             nil,
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
				CycleDeadline:          tt.CycleDeadline,
				CycleFrequency:         tt.CycleFrequency,
				GoalQuantity:           tt.GoalQuantity,
				CurScorableItems:       tt.CurScorableItems,
				CurBoxesAwarded:        tt.CurBoxesAwarded,
				CurGoalStreak:          tt.CurGoalStreak,
				MaxGoalStreak:          tt.MaxGoalStreak,
				MaxCycleItemsCompleted: tt.MaxCycleItemsCompleted,
				TotalItemsCompleted:    tt.TotalItemsCompleted,
				TotalBoxesAwarded:      tt.TotalBoxesAwarded,
				FirstCompleted:         tt.FirstCompleted,
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
			assert.Equal(t, *tt.CycleDeadline, underlyingTracker.CycleDeadline)
			assert.Equal(t, *tt.CycleFrequency, string(underlyingTracker.CycleFrequency))
			assert.Equal(t, *tt.GoalQuantity, underlyingTracker.GoalQuantity)
			assert.Equal(t, *tt.CurScorableItems, underlyingTracker.CurScorableItems)
			assert.Equal(t, *tt.CurBoxesAwarded, underlyingTracker.CurBoxesAwarded)
			assert.Equal(t, *tt.CurGoalStreak, underlyingTracker.CurGoalStreak)
			assert.Equal(t, *tt.MaxGoalStreak, underlyingTracker.MaxGoalStreak)
			assert.Equal(t, *tt.MaxCycleItemsCompleted, underlyingTracker.MaxCycleItemsCompleted)
			assert.Equal(t, *tt.TotalItemsCompleted, underlyingTracker.TotalItemsCompleted)
			assert.Equal(t, *tt.TotalBoxesAwarded, underlyingTracker.TotalBoxesAwarded)
			assert.Equal(t, tt.FirstCompleted, underlyingTracker.FirstCompleted)
		})
	}
}

func Test_UpdateJobAppTrackerFields(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler(), collection.NewService(collection.NewRepository(dBase)))

	n := time.Now().Round(time.Hour)
	tests := []struct {
		name                   string
		CycleDeadline          *time.Time
		CycleFrequency         *string
		GoalQuantity           *int
		CurScorableItems       *int
		CurBoxesAwarded        *int
		CurGoalStreak          *int
		MaxGoalStreak          *int
		MaxCycleItemsCompleted *int
		TotalItemsCompleted    *int
		TotalBoxesAwarded      *int
		FirstCompleted         *time.Time
		wantErrMsg             []string
	}{
		{
			name:                   "valid-fields",
			CycleDeadline:          &n,
			GoalQuantity:           intPtr(1),
			CycleFrequency:         strPtr(string(scheduler.DefaultFreq)),
			CurScorableItems:       intPtr(0),
			CurBoxesAwarded:        intPtr(3),
			CurGoalStreak:          intPtr(4),
			MaxGoalStreak:          intPtr(5),
			MaxCycleItemsCompleted: intPtr(6),
			TotalItemsCompleted:    intPtr(7),
			TotalBoxesAwarded:      intPtr(8),
			FirstCompleted:         &n,
			wantErrMsg:             nil,
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
			defer dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")

			require.NoError(t, err)
			assert.NotNil(t, t1)

			updateFields := &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
				CycleDeadline:          tt.CycleDeadline,
				CycleFrequency:         tt.CycleFrequency,
				GoalQuantity:           tt.GoalQuantity,
				CurScorableItems:       tt.CurScorableItems,
				CurBoxesAwarded:        tt.CurBoxesAwarded,
				CurGoalStreak:          tt.CurGoalStreak,
				MaxGoalStreak:          tt.MaxGoalStreak,
				MaxCycleItemsCompleted: tt.MaxCycleItemsCompleted,
				TotalItemsCompleted:    tt.TotalItemsCompleted,
				TotalBoxesAwarded:      tt.TotalBoxesAwarded,
				FirstCompleted:         tt.FirstCompleted,
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
			assert.Equal(t, *tt.CycleDeadline, updateTracker.CycleDeadline)
			assert.Equal(t, *tt.CycleFrequency, string(updateTracker.CycleFrequency))
			assert.Equal(t, *tt.GoalQuantity, updateTracker.GoalQuantity)
			assert.Equal(t, *tt.CurScorableItems, updateTracker.CurScorableItems)
			assert.Equal(t, *tt.CurBoxesAwarded, updateTracker.CurBoxesAwarded)
			assert.Equal(t, *tt.CurGoalStreak, updateTracker.CurGoalStreak)
			assert.Equal(t, *tt.MaxGoalStreak, updateTracker.MaxGoalStreak)
			assert.Equal(t, *tt.MaxCycleItemsCompleted, updateTracker.MaxCycleItemsCompleted)
			assert.Equal(t, *tt.TotalItemsCompleted, updateTracker.TotalItemsCompleted)
			assert.Equal(t, *tt.TotalBoxesAwarded, updateTracker.TotalBoxesAwarded)
			assert.Equal(t, tt.FirstCompleted, updateTracker.FirstCompleted)

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
			newDeadline := tt.CycleDeadline.Add(time.Hour)
			moreItems := *tt.CurScorableItems + 10
			updateTracker2_5, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(),
				&JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
					CycleDeadline:    &newDeadline,
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

			// expect validation errors
			wantErr := []string{maxGoalStreakField, maxCycleItemsCompletedField, totalItemsCompletedField, totalBoxesAwardedField, firstCompletedField}
			newFirstCompleted := tt.FirstCompleted.Add(100 * time.Hour)
			updateTracker3, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(),
				&JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
					MaxGoalStreak:          intPtr(*tt.MaxGoalStreak - 1),
					MaxCycleItemsCompleted: intPtr(*tt.MaxCycleItemsCompleted - 1),
					TotalItemsCompleted:    intPtr(*tt.TotalItemsCompleted - 1),
					TotalBoxesAwarded:      intPtr(*tt.TotalBoxesAwarded - 1),
					FirstCompleted:         &newFirstCompleted,
				}})
			assert.Nil(t, updateTracker3)
			for _, e := range wantErr {
				assert.ErrorContains(t, err, e)
			}
		})
	}
}

func Test_CreateJobAppItem(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{}, &JobAppItem{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	dBase.Exec("TRUNCATE TABLE job_app_items RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler(), collection.NewService(collection.NewRepository(dBase)))

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
			dummyUser := user.User{ID: "user_12345678"}
			t1, err := svc.CreateNewJobAppTracker(&dummyUser)
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
			dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
			dBase.Exec("TRUNCATE TABLE job_app_items RESTART IDENTITY CASCADE")
		})
	}

}

func Test_UpdateJobAppItemFields(t *testing.T) {
	// test updating the status complete field and seeing if tracker curItemsComplete decrements by 1
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{}, &JobAppItem{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	dBase.Exec("TRUNCATE TABLE job_app_items RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler(), collection.NewService(collection.NewRepository(dBase)))

	dummyUser := user.User{ID: "usr_99999999"}
	tracker, err := svc.CreateNewJobAppTracker(&dummyUser)
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
	_, err = svc.CreateJobAppItem(dummyUser.GetID(), &JobAppItemUpdateFields{
		Title:        strPtr("Item 1"),
		Body:         strPtr("Body 1"),
		Status:       &statusComplete,
		IsAttributed: boolPtr(false),
	})
	require.NoError(t, err)
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

	// assume the first item is the one that is created first; gorm sorting should ensure this
	item1 := tracker.Items[0]
	item2 := tracker.Items[1]

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

	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	dBase.Exec("TRUNCATE TABLE job_app_items RESTART IDENTITY CASCADE")
}

func Test_Scheduler(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler(), collection.NewService(collection.NewRepository(dBase)))

	t0 := time.Now().Round(time.Second)
	t1 := t0.Add(time.Second * 4)
	t2 := t1.Add(time.Second * 4)
	t3 := t2.Add(time.Second * 4)

	u0 := user.User{ID: "usr_00000000"}
	u1 := user.User{ID: "usr_11111111"}
	u2 := user.User{ID: "usr_22222222"}
	u3 := user.User{ID: "usr_33333333"}

	// start scheduler
	// go svc.scheduler.Start()
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
			CycleDeadline:  &t0,
			CycleFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt0, jt00)
	jt11, err := svc.UpdateJobAppTrackerFields(u1.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CycleDeadline:  &t1,
			CycleFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt1, jt11)
	jt22, err := svc.UpdateJobAppTrackerFields(u2.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CycleDeadline:  &t2,
			CycleFrequency: strPtr("weekly"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt2, jt22)
	jt33, err := svc.UpdateJobAppTrackerFields(u3.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CycleDeadline:  &t3,
			CycleFrequency: strPtr("weekly"),
		},
	})
	require.NoError(t, err)
	assert.NotEqual(t, jt3, jt33)
	time.Sleep(time.Second * 28)

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
	dBase.AutoMigrate(&JobAppTracker{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	dBase.Exec("TRUNCATE TABLE job_app_items RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler(), collection.NewService(collection.NewRepository(dBase)))

	dummyUser := user.User{ID: "usr_12345678"}
	t1, err := svc.CreateNewJobAppTracker(&dummyUser)
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
	dummyUser2 := user.User{ID: "usr_23456789"}
	t2, err := svc.CreateNewJobAppTracker(&dummyUser2)
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

	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	dBase.Exec("TRUNCATE TABLE job_app_items RESTART IDENTITY CASCADE")
}
