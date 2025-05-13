package tracker

import (
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

// heres some tests
func Test_formatForRepo(t *testing.T) {
	n := time.Now()
	tests := []struct {
		name                   string
		GoalDeadline           *time.Time
		GoalFrequency          *string
		GoalQuantity           *int
		CurItemsCompleted      *int
		CurBoxesAwarded        *int
		CurGoalStreak          *int
		MaxGoalStreak          *int
		MaxItemsCompletedDaily *int
		TotalItemsCompleted    *int
		TotalBoxesAwarded      *int
		FirstCompleted         *time.Time
		wantErrMsg             []string
	}{
		{
			name:                   "valid-fields",
			GoalDeadline:           &n,
			GoalQuantity:           intPtr(1),
			GoalFrequency:          strPtr(string(scheduler.DefaultFreq)),
			CurItemsCompleted:      intPtr(2),
			CurBoxesAwarded:        intPtr(3),
			CurGoalStreak:          intPtr(4),
			MaxGoalStreak:          intPtr(5),
			MaxItemsCompletedDaily: intPtr(6),
			TotalItemsCompleted:    intPtr(7),
			TotalBoxesAwarded:      intPtr(8),
			FirstCompleted:         &n,
			wantErrMsg:             nil,
		},
		{
			name:          "invalid-frequency",
			GoalFrequency: strPtr("1234671"),
			wantErrMsg:    []string{"invalid goal frequency"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateFields := &UnderlyingTrackerUpdateFields{
				GoalDeadline:           tt.GoalDeadline,
				GoalFrequency:          tt.GoalFrequency,
				GoalQuantity:           tt.GoalQuantity,
				CurItemsCompleted:      tt.CurItemsCompleted,
				CurBoxesAwarded:        tt.CurBoxesAwarded,
				CurGoalStreak:          tt.CurGoalStreak,
				MaxGoalStreak:          tt.MaxGoalStreak,
				MaxItemsCompletedDaily: tt.MaxItemsCompletedDaily,
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
			assert.Equal(t, *tt.GoalDeadline, underlyingTracker.GoalDeadline)
			assert.Equal(t, *tt.GoalFrequency, string(underlyingTracker.GoalFrequency))
			assert.Equal(t, *tt.GoalQuantity, underlyingTracker.GoalQuantity)
			assert.Equal(t, *tt.CurItemsCompleted, underlyingTracker.CurItemsCompleted)
			assert.Equal(t, *tt.CurBoxesAwarded, underlyingTracker.CurBoxesAwarded)
			assert.Equal(t, *tt.CurGoalStreak, underlyingTracker.CurGoalStreak)
			assert.Equal(t, *tt.MaxGoalStreak, underlyingTracker.MaxGoalStreak)
			assert.Equal(t, *tt.MaxItemsCompletedDaily, underlyingTracker.MaxItemsCompletedDaily)
			assert.Equal(t, *tt.TotalItemsCompleted, underlyingTracker.TotalItemsCompleted)
			assert.Equal(t, *tt.TotalBoxesAwarded, underlyingTracker.TotalBoxesAwarded)
			assert.Equal(t, tt.FirstCompleted, underlyingTracker.FirstCompleted)
		})
	}
}

func Test_UpdateJobAppTrackerFields(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler())

	n := time.Now().Round(time.Hour)
	tests := []struct {
		name                   string
		GoalDeadline           *time.Time
		GoalFrequency          *string
		GoalQuantity           *int
		CurItemsCompleted      *int
		CurBoxesAwarded        *int
		CurGoalStreak          *int
		MaxGoalStreak          *int
		MaxItemsCompletedDaily *int
		TotalItemsCompleted    *int
		TotalBoxesAwarded      *int
		FirstCompleted         *time.Time
		wantErrMsg             []string
	}{
		{
			name:                   "valid-fields",
			GoalDeadline:           &n,
			GoalQuantity:           intPtr(1),
			GoalFrequency:          strPtr(string(scheduler.DefaultFreq)),
			CurItemsCompleted:      intPtr(2),
			CurBoxesAwarded:        intPtr(3),
			CurGoalStreak:          intPtr(4),
			MaxGoalStreak:          intPtr(5),
			MaxItemsCompletedDaily: intPtr(6),
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
			require.NoError(t, err)
			assert.NotNil(t, t1)

			updateFields := &JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
				GoalDeadline:           tt.GoalDeadline,
				GoalFrequency:          tt.GoalFrequency,
				GoalQuantity:           tt.GoalQuantity,
				CurItemsCompleted:      tt.CurItemsCompleted,
				CurBoxesAwarded:        tt.CurBoxesAwarded,
				CurGoalStreak:          tt.CurGoalStreak,
				MaxGoalStreak:          tt.MaxGoalStreak,
				MaxItemsCompletedDaily: tt.MaxItemsCompletedDaily,
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
			assert.Equal(t, *tt.GoalDeadline, updateTracker.GoalDeadline)
			assert.Equal(t, *tt.GoalFrequency, string(updateTracker.GoalFrequency))
			assert.Equal(t, *tt.GoalQuantity, updateTracker.GoalQuantity)
			assert.Equal(t, *tt.CurItemsCompleted, updateTracker.CurItemsCompleted)
			assert.Equal(t, *tt.CurBoxesAwarded, updateTracker.CurBoxesAwarded)
			assert.Equal(t, *tt.CurGoalStreak, updateTracker.CurGoalStreak)
			assert.Equal(t, *tt.MaxGoalStreak, updateTracker.MaxGoalStreak)
			assert.Equal(t, *tt.MaxItemsCompletedDaily, updateTracker.MaxItemsCompletedDaily)
			assert.Equal(t, *tt.TotalItemsCompleted, updateTracker.TotalItemsCompleted)
			assert.Equal(t, *tt.TotalBoxesAwarded, updateTracker.TotalBoxesAwarded)
			assert.Equal(t, tt.FirstCompleted, updateTracker.FirstCompleted)

			// test subsequent update -- happy path
			newQuantity := *tt.GoalQuantity + 5
			updateTracker2, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(),
				&JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
					GoalQuantity:  &newQuantity,
					GoalFrequency: strPtr("daily"),
				}})
			require.NoError(t, err)
			require.NotNil(t, updateTracker2)
			assert.Equal(t, newQuantity, updateTracker2.GoalQuantity)
			assert.Equal(t, scheduler.FreqDaily, updateTracker2.GoalFrequency)

			// test scheduler reassignment
			newDeadline := tt.GoalDeadline.Add(time.Hour)
			moreItems := *tt.CurItemsCompleted + 10
			updateTracker2_5, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(),
				&JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
					GoalDeadline:      &newDeadline,
					GoalFrequency:     strPtr("weekly"),
					CurItemsCompleted: &moreItems,
				}})
			require.NotNil(t, updateTracker2_5)
			require.NoError(t, err)
			assert.Equal(t, newDeadline, updateTracker2_5.GoalDeadline)
			assert.Equal(t, scheduler.FreqWeekly, updateTracker2_5.GoalFrequency)
			assert.Equal(t, moreItems, updateTracker2_5.CurItemsCompleted)

			// expect validation errors
			wantErr := []string{maxGoalStreakField, maxItemsCompletedDailyField, totalItemsCompletedField, totalBoxesAwardedField, firstCompletedField}
			newFirstCompleted := tt.FirstCompleted.Add(100 * time.Hour)
			updateTracker3, err := svc.UpdateJobAppTrackerFields(t1.GetUserID(),
				&JobAppTrackerUpdateFields{UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
					MaxGoalStreak:          intPtr(*tt.MaxGoalStreak - 1),
					MaxItemsCompletedDaily: intPtr(*tt.MaxItemsCompletedDaily - 1),
					TotalItemsCompleted:    intPtr(*tt.TotalItemsCompleted - 1),
					TotalBoxesAwarded:      intPtr(*tt.TotalBoxesAwarded - 1),
					FirstCompleted:         &newFirstCompleted,
				}})
			assert.Nil(t, updateTracker3)
			for _, e := range wantErr {
				assert.ErrorContains(t, err, e)
			}
			dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
		})
	}
}

func Test_CreateJobAppItem(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{}, &JobAppItem{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	dBase.Exec("TRUNCATE TABLE job_app_items RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler())

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
			name:            "with-attr-time",
			Title:           strPtr("Title"),
			Body:            strPtr("Body"),
			Status:          &defaultStatus,
			IsAttributed:    boolPtr(true),
			AttributionTime: &n,
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
				AttributionTime: tt.AttributionTime,
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

func Test_Scheduler(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler())

	t0 := time.Now().Round(time.Second)
	t1 := t0.Add(time.Second * 4)
	t2 := t1.Add(time.Second * 4)
	t3 := t2.Add(time.Second * 4)

	u0 := user.User{ID: "usr_00000000"}
	u1 := user.User{ID: "usr_11111111"}
	u2 := user.User{ID: "usr_22222222"}
	u3 := user.User{ID: "usr_33333333"}

	// start scheduler
	go svc.scheduler.Start()
	go svc.StartTrackerUpdateListener()

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
			GoalDeadline:  &t0,
			GoalFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	jt11, err := svc.UpdateJobAppTrackerFields(u1.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			GoalDeadline:  &t1,
			GoalFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	jt22, err := svc.UpdateJobAppTrackerFields(u2.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			GoalDeadline:  &t2,
			GoalFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	jt33, err := svc.UpdateJobAppTrackerFields(u3.GetID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			GoalDeadline:  &t3,
			GoalFrequency: strPtr("daily"),
		},
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 28)

	jt000, err := svc.LookupJobAppTrackerFromTrackerID(jt0.GetID())
	require.NoError(t, err)
	jt111, err := svc.LookupJobAppTrackerFromTrackerID(jt1.GetID())
	require.NoError(t, err)
	jt222, err := svc.LookupJobAppTrackerFromTrackerID(jt2.GetID())
	require.NoError(t, err)
	jt333, err := svc.LookupJobAppTrackerFromTrackerID(jt3.GetID())
	require.NoError(t, err)
	assert.Equal(t, t0.AddDate(0, 0, 1), jt000.GoalDeadline)
	assert.Equal(t, t1.AddDate(0, 0, 1), jt111.GoalDeadline)
	assert.Equal(t, t2.AddDate(0, 0, 1), jt222.GoalDeadline)
	assert.Equal(t, t3.AddDate(0, 0, 1), jt333.GoalDeadline)
	// assert.NotEqual(t, jt000.GoalDeadline, jt00.GoalDeadline)
	// assert.NotEqual(t, jt111.GoalDeadline, jt11.GoalDeadline)
	// assert.NotEqual(t, jt222.GoalDeadline, jt22.GoalDeadline)
	// assert.NotEqual(t, jt333.GoalDeadline, jt33.GoalDeadline)
	log.Printf("after: %v, before: %v", jt000.GoalDeadline, jt00.GoalDeadline)
	log.Printf("after: %v, before: %v", jt111.GoalDeadline, jt11.GoalDeadline)
	log.Printf("after: %v, before: %v", jt222.GoalDeadline, jt22.GoalDeadline)
	log.Printf("after: %v, before: %v", jt333.GoalDeadline, jt33.GoalDeadline)
	svc.scheduler.Stop()
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")

}
