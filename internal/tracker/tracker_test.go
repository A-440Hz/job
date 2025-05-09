package tracker

import (
	"job/internal/db"
	"job/internal/scheduler"
	"job/internal/user"
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
			dummyUser := user.User{ID: "user_12345678"}
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
			updateTracker, err := svc.UpdateJobAppTrackerFields(t1.GetID(), updateFields)
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
			updateTracker2, err := svc.UpdateJobAppTrackerFields(t1.GetID(),
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
			updateTracker2_5, err := svc.UpdateJobAppTrackerFields(t1.GetID(),
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
			updateTracker3, err := svc.UpdateJobAppTrackerFields(t1.GetID(),
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

func Test_Scheduler(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormDB()
	require.NoError(t, err)
	dBase.AutoMigrate(&JobAppTracker{})
	dBase.Exec("TRUNCATE TABLE job_app_trackers RESTART IDENTITY CASCADE")
	repo := NewRepository(dBase)
	svc := NewService(repo, scheduler.NewScheduler())
}
