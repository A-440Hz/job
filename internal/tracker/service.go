package tracker

import (
	"errors"
	"job/internal/scheduler"
	"job/internal/user"
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
// This method functions like a create hook
func newUnderlyingTracker(u *user.User) *UnderlyingTracker {
	ut := &UnderlyingTracker{
		UserID:        u.GetID(),
		GoalDeadline:  scheduler.GetDefaultGoalDeadline(u.Timezone),
		GoalFrequency: scheduler.GetDefaultGoalFrequency(),
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

// LookupJobAppTrackerFromUserID (by uuid) is the default function used in the service layer to lookup trackers.
// It relies on a db.Where call in gorm, which is efficient because user_id is specified as a gorm index.
func (s *Service) LookupJobAppTrackerFromUserID(uuid string) (*JobAppTracker, error) {
	t, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// LookupJobAppTrackerFromTrackerID is probably needed for the scheduler to trigger tracker
func (s *Service) LookupJobAppTrackerFromTrackerID(id string) (*JobAppTracker, error) {
	t, err := s.repo.lookupJobAppTrackerFromTrackerID(id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) GetJobAppTrackerWithItemsFromUserID(uuid string) (*JobAppTracker, error) {
	t, err := s.repo.GetJobAppTrackerWithItemsFromUserID(uuid)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) UpdateJobAppTrackerFields(uuid string, fields *JobAppTrackerUpdateFields) (*JobAppTracker, error) {
	_, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
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
	updateTracker.UserID = uuid
	_, err = s.repo.UpdateJobAppTrackerFields(updateTracker, updateFields)
	if err != nil {
		return nil, err
	}
	return s.LookupJobAppTrackerFromUserID(uuid)
}

func (s *Service) DeleteJobAppTracker(id string) error {
	t := &JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}}
	if err := s.repo.DeleteJobAppTracker(t); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetScorableJobAppItems(t *JobAppTracker) ([]JobAppItem, error) {
	return s.repo.GetScorableJobAppItems(t)
}

func (s *Service) CreateJobAppItem(userID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker; return immediately on invalid id
	t, err := s.repo.LookupJobAppTrackerFromUserID(userID)
	if err != nil {
		return nil, err
	}
	badFields := map[string]string{}
	if fields.Title == nil {
		badFields["title"] = "missing item name"
	}
	if fields.Status == nil {
		badFields["status"] = "missing item status"
	}
	if fields.IsAttributed == nil {
		badFields["is_attributed"] = "missing isAttributed"
	}
	i, _, err := fields.formatForRepo()
	if err != nil {
		badFields["format"] = err.Error()
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for _, v := range badFields {
			err = errors.Join(err, errors.New(v))
		}
		return nil, err
	}

	i.TrackerID = t.GetID()
	_, err = s.repo.CreateJobAppTrackerItem(i)
	if err != nil {
		return nil, err
	}
	return s.repo.GetJobAppTrackerWithItemsFromUserID(userID)
}

func (s *Service) LookupJobAppItems(trackerID string) ([]*JobAppItem, error) {
	items, err := s.repo.LookupJobAppItems(trackerID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) UpdateJobAppItemFields(uuid string, itemID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker
	t, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return nil, err
	}
	// valid item
	_, err = s.repo.LookupJobAppItem(t.GetID(), itemID)
	if err != nil {
		return nil, err
	}

	// I cant think of a reason to validate any fields
	updateItem, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateJobAppTrackerItemFields(updateItem, updateFields)
	if err != nil {
		return nil, err
	}
	return s.repo.GetJobAppTrackerWithItemsFromUserID(uuid)
}

// func (t *JobAppTracker) ResetProgress() {
// 	t.numBoxesAwarded = 0
// 	t.numItemsCompleted = 0
// }

// func (t *JobAppTracker) GetNextDeadline() (time.Time, error) {
// 	cd := t.GoalDeadline
// 	switch t.GoalFrequency {
// 	case FreqDaily:
// 		return cd.Add(24 * time.Hour), nil
// 	case FreqWeekly:
// 		return cd.Add(24 * time.Hour * 7), nil
// 	default:
// 		return cd, fmt.Errorf("invalid goal frequency: %q", t.GoalFrequency)
// 	}
// }

// func (t *JobAppTracker) EditFrequency(f Frequency) error {
// 	if f == t.GoalFrequency {
// 		return nil
// 	}
// 	t.GoalFrequency = f
// 	// db logic here
// 	// scheduler logic here
// 	return nil
// }
