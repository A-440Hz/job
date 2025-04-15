package tracker

import (
	"errors"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateNewJobAppTracker() (*JobAppTracker, error) {
	t := &JobAppTracker{}
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

func (s *Service) DeleteJobAppTracker(id string) error {
	t := &JobAppTracker{ID: id}
	if err := s.repo.DeleteJobAppTracker(t); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetJobAppTrackerItems(trackerID string) ([]*JobAppItem, error) {
	items, err := s.repo.GetJobAppTrackerItems(trackerID)
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
		return cd, errors.New(fmt.Sprintf("invalid goal frequency: %q", t.GoalFrequency))
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
