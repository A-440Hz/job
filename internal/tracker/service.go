package tracker

import (
	"fmt"
	"job/internal/user"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) createNewUnderlyingTracker(u *user.User) (*UnderlyingTracker, error) {
	ut, err := s.repo.CreateUnderlyingTracker(u)
	if err != nil {
		return nil, err
	}
	return ut, nil
}

func (s *Service) CreateNewJobAppTracker(u *user.User) (*JobAppTracker, error) {
	ut, err := s.createNewUnderlyingTracker(u)
	if err != nil {
		return nil, err
	}
	t := &JobAppTracker{UnderlyingTracker: *ut}
	t, err = s.repo.CreateJobAppTracker(t)
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

func (s *Service) UpdateJobAppTrackerFields(id string, fields *TrackerUpdateFields) (*JobAppTracker, error) {
	t, err := s.repo.LookupJobAppTracker(id)
	if err != nil {
		return nil, err
	}
	updateFields := []string{}

}

func (s *Service) DeleteJobAppTracker(id string) error {
	t := &JobAppTracker{ID: id}
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
	cd := t.Tracker.GoalDeadline
	switch t.Tracker.GoalFrequency {
	case FreqDaily:
		return cd.Add(24 * time.Hour), nil
	case FreqWeekly:
		return cd.Add(24 * time.Hour * 7), nil
	default:
		return cd, fmt.Errorf("invalid goal frequency: %q", t.Tracker.GoalFrequency)
	}
}

func (t *JobAppTracker) EditFrequency(f Frequency) error {
	if f == t.Tracker.GoalFrequency {
		return nil
	}
	t.Tracker.GoalFrequency = f
	// db logic here
	// scheduler logic here
	return nil
}
