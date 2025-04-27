package tracker

import (
	"errors"
	"fmt"
	"job/internal/scheduler"
	"job/internal/user"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

// The underlying tracker is a base struct that contains the common fields for all trackers.
// There is no good reason for it to be a separate entity in the database, so I will create it in memory and store the two trackers together in gorm.
// The separation is primarily to fulfill the factory pattern and create specific types of Items.
func newUnderlyingTracker(u *user.User) *UnderlyingTracker {
	ut := &UnderlyingTracker{
		UserID:       u.GetID(),
		GoalDeadline: scheduler.GetDefaultGoalDeadline(u.Timezone),
	}
	return ut
}

func (s *Service) CreateNewJobAppTracker(u *user.User) (*JobAppTracker, error) {
	ut := newUnderlyingTracker(u)
	t := &JobAppTracker{UnderlyingTracker: *ut}
	t, err := s.repo.CreateJobAppTracker(t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) LookupJobAppTracker(id string) (*JobAppTracker, error) {
	t, err := s.repo.LookupJobAppTracker(id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) updateJobAppTracker(t *JobAppTracker) (*JobAppTracker, error) {
	t, err := s.repo.UpdateJobAppTracker(t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) UpdateJobAppTrackerFields(id string, fields *JobAppTrackerUpdateFields) (*JobAppTracker, error) {
	_, err := s.repo.LookupJobAppTracker(id)
	if err != nil {
		return nil, err
	}
	updateTracker, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}

	// validate?
	updateTracker.ID = id
	_, err = s.repo.UpdateJobAppTrackerFields(updateTracker, updateFields)
	if err != nil {
		return nil, err
	}
	return s.LookupJobAppTracker(id)
}

func (s *Service) DeleteJobAppTracker(id string) error {
	t := &JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}}
	if err := s.repo.DeleteJobAppTracker(t); err != nil {
		return err
	}
	return nil
}

func (s *Service) LookupJobAppTrackerItems(trackerID string) ([]*JobAppItem, error) {
	items, err := s.repo.LookupJobAppTrackerItems(trackerID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (t *JobAppTracker) ResetProgress() {
	t.numBoxesAwarded = 0
	t.numItemsCompleted = 0
}

func (t *JobAppTracker) GetNextDeadline() (time.Time, error) {
	cd := t.GoalDeadline
	switch t.GoalFrequency {
	case FreqDaily:
		return cd.Add(24 * time.Hour), nil
	case FreqWeekly:
		return cd.Add(24 * time.Hour * 7), nil
	default:
		return cd, fmt.Errorf("invalid goal frequency: %q", t.GoalFrequency)
	}
}

func (t *JobAppTracker) EditFrequency(f Frequency) error {
	if f == t.GoalFrequency {
		return nil
	}
	t.GoalFrequency = f
	// db logic here
	// scheduler logic here
	return nil
}
