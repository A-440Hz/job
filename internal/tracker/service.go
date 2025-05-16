package tracker

import (
	"errors"
	"fmt"
	"job/internal/scheduler"
	"job/internal/user"
	"log"
	"slices"
	"time"
)

type Service struct {
	repo      *Repository
	scheduler *scheduler.Scheduler
}

func NewService(r *Repository, s *scheduler.Scheduler) *Service {
	return &Service{repo: r, scheduler: s}
}

// This method functions like a create hook for UnderlyingTracker.
func (t *JobAppTracker) createUnderlyingTracker(u *user.User) {
	t.UnderlyingTracker = UnderlyingTracker{
		UserID:        u.GetID(),
		GoalDeadline:  scheduler.GetDefaultGoalDeadline(u.Timezone),
		GoalFrequency: scheduler.GetDefaultGoalFrequency(),
		TrackerType:   JobAppTrackerType,
	}
}

func (s *Service) CreateNewJobAppTracker(u *user.User) (*JobAppTracker, error) {
	t := &JobAppTracker{}
	t.createUnderlyingTracker(u)
	t, err := s.repo.CreateJobAppTracker(t)
	if err != nil {
		return nil, err
	}
	tg := t.ToTrackerGoal()
	err = s.scheduler.AddTrackerGoal(tg)
	if err != nil {
		// TODO: decide on correct error response in scheduler
		log.Print(err)
		return nil, err
	}
	return t, nil
}

// LookupJobAppTrackerFromUserID (by uuid) is the default function used in the service layer to lookup trackers.
// It relies on a db.Where call in gorm, which is efficient because user_id is specified as a gorm index.
func (s *Service) LookupJobAppTrackerFromUserID(uuid string) (*JobAppTracker, error) {
	return s.repo.LookupJobAppTrackerFromUserID(uuid)
}

// LookupJobAppTrackerFromTrackerID is probably needed for the scheduler to trigger tracker
func (s *Service) LookupJobAppTrackerFromTrackerID(id string) (*JobAppTracker, error) {
	return s.repo.lookupJobAppTrackerFromTrackerID(id)
}

func (s *Service) GetJobAppTrackerWithItemsFromUserID(uuid string) (*JobAppTracker, error) {
	return s.repo.GetJobAppTrackerWithItemsFromUserID(uuid)
}

func (s *Service) UpdateJobAppTrackerFields(uuid string, fields *JobAppTrackerUpdateFields) (*JobAppTracker, error) {
	repoTracker, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return nil, err
	}

	// regarding the updateDeadline field: because it's a time.Time value,
	// I will assume the timezone is correctly accounted for from the front end
	// and I can throw it straight into the scheduler
	badFields := map[string]string{}
	updateTracker, updateFields, err := fields.formatForRepo()
	if err != nil {
		badFields["format"] = err.Error()
	}
	if len(updateFields) == 0 {
		badFields["fields"] = "no fields to update"
	}

	// TODO: probably move this out into a validate function
	// stats invariants
	if slices.Contains(updateFields, maxGoalStreakField) && updateTracker.MaxGoalStreak < repoTracker.MaxGoalStreak {
		badFields[maxGoalStreakField] = fmt.Sprintf("cannot decrease %q", maxGoalStreakField)
	}
	if slices.Contains(updateFields, maxItemsCompletedDailyField) && updateTracker.MaxItemsCompletedDaily < repoTracker.MaxItemsCompletedDaily {
		badFields[maxItemsCompletedDailyField] = fmt.Sprintf("cannot decrease %q", maxItemsCompletedDailyField)
	}
	if slices.Contains(updateFields, totalItemsCompletedField) && updateTracker.TotalItemsCompleted < repoTracker.TotalItemsCompleted {
		badFields[totalItemsCompletedField] = fmt.Sprintf("cannot decrease %q", totalItemsCompletedField)
	}
	if slices.Contains(updateFields, totalBoxesAwardedField) && updateTracker.TotalBoxesAwarded < repoTracker.TotalBoxesAwarded {
		badFields[totalBoxesAwardedField] = fmt.Sprintf("cannot decrease %q", totalBoxesAwardedField)
	}
	if slices.Contains(updateFields, firstCompletedField) && repoTracker.FirstCompleted != nil && updateTracker.FirstCompleted.After(*repoTracker.FirstCompleted) {
		badFields[firstCompletedField] = fmt.Sprintf("cannot increment %q", firstCompletedField)
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for _, v := range badFields {
			err = errors.Join(err, errors.New(v))
		}
		return nil, err
	}

	// the front end should also avoid sending update reqs for identical deadlines and frequencies
	if repoTracker.GoalDeadline == updateTracker.GoalDeadline {
		//
		updateFields = slices.DeleteFunc(updateFields, func(f string) bool {
			return f == goalDeadlineField
		})
	}
	if repoTracker.GoalFrequency == updateTracker.GoalFrequency {
		updateFields = slices.DeleteFunc(updateFields, func(f string) bool {
			return f == goalFrequencyField
		})
	}

	// push user updates to repo
	updateTracker.ID = repoTracker.GetID()
	updateTracker, err = s.repo.UpdateJobAppTrackerFields(updateTracker, updateFields)
	if err != nil {
		return nil, err
	}

	// communicate with scheduler if needed.. updateFields was trimmed earlier if this step is not needed
	if slices.Contains(updateFields, goalDeadlineField) || slices.Contains(updateFields, goalFrequencyField) {
		// fail scheduler gracefully?
		err = s.scheduler.Update(repoTracker.ToTrackerGoal(), updateTracker.ToTrackerGoal())
		if err != nil {
			log.Print(err)
		}
	}

	// validate counters and score tracker if needed
	if slices.Contains(updateFields, goalQuantityField) || slices.Contains(updateFields, curItemsCompletedField) {
		repoTracker, err = s.repo.LookupJobAppTrackerFromUserID(uuid)
		if err != nil {
			return nil, err
		}
		updateTracker, updateFields, err = s.ValidateCounters(repoTracker)
		if err != nil {
			return nil, err
		}
		if len(updateFields) > 0 {
			_, err = s.repo.UpdateJobAppTrackerFields(updateTracker, updateFields)
			if err != nil {
				return nil, err
			}
		}
	}

	return s.LookupJobAppTrackerFromUserID(uuid)
}

func (s *Service) DeleteJobAppTracker(id string) error {
	return s.repo.DeleteJobAppTracker(&JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}})
}

// GetScorableJobAppItems returns a list of job app items with create time within the tracker's goal timeframe
// and are StatusComplete and not yet attributed
func (s *Service) GetScorableJobAppItems(t *JobAppTracker) ([]JobAppItem, error) {
	return s.repo.GetScorableJobAppItems(t)
}

// ValidateCounters is intended for use on a tracker queried from the repo with all required fields populated.
// it checks for overflow of curItemsCompleted and goalQuantity, incrementing and resetting curBoxesAwarded and curItemsCompleted as needed, and returns the corresponding update fields.
// It calls the service to score the tracker items if the goal quantity is reached.
// The only user-updatable field that directly invokes this function is goalQuantity. Managed properly, this function should be fine to call on every update.
func (s *Service) ValidateCounters(repoTracker *JobAppTracker) (*JobAppTracker, []string, error) {
	fields := []string{}
	if repoTracker.CurItemsCompleted < repoTracker.GoalQuantity {
		return repoTracker, fields, nil
	}
	n := time.Now()
	numToAward := repoTracker.CurItemsCompleted / repoTracker.GoalQuantity
	items, err := s.GetScorableJobAppItems(repoTracker)
	if err != nil {
		return nil, nil, err
	}
	if len(items) < repoTracker.GoalQuantity*numToAward {
		// TODO: this will just be broken forever. handle it gracefully or make some process to help avoid this state.
		log.Print("mismatch between number of scorable items and tracker current items completed: " + fmt.Sprintf("%d < %d", len(items), repoTracker.CurItemsCompleted))
		// is it good to set curItemsCompleted to len(items) here?
		return nil, nil, errors.New("mismatch between number of scorable items and tracker current items completed: " + fmt.Sprintf("%d < %d", len(items), repoTracker.CurItemsCompleted))
	}
	badFields := map[string]string{}
	updateItems := []string{}
	// score n items. They should be already sorted by create time
	for i := range repoTracker.GoalQuantity * numToAward {
		updateItems = append(updateItems, items[i].GetID())
	}

	tr := true
	for _, id := range updateItems {
		err := s.updateJobAppItemFields(repoTracker, id, &JobAppItemUpdateFields{
			AttributionTime: &n,
			IsAttributed:    &tr,
		})
		if err != nil {
			badFields[id] = err.Error()
		}
	}
	if len(badFields) > 0 {
		err := errors.New("update error:")
		for id, v := range badFields {
			err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
		}
		return nil, nil, err
	}
	repoTracker.CurItemsCompleted = repoTracker.CurItemsCompleted - repoTracker.GoalQuantity*numToAward
	repoTracker.CurBoxesAwarded = repoTracker.CurBoxesAwarded + numToAward
	return repoTracker, []string{curItemsCompletedField, curBoxesAwardedField}, nil
}

// scoreTracker updates the tracker items and tracker fields if the goal quantity is reached,
// otherwise it returns the tracker as is
func (s *Service) scoreTracker(uuid string) (*JobAppTracker, error) {
	t, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return nil, err
	}
	if t.CurItemsCompleted < t.GoalQuantity {
		return t, nil
	}
	n := time.Now()
	items, err := s.GetScorableJobAppItems(t)
	if err != nil {
		return nil, err
	}
	if len(items) < t.GoalQuantity {
		return nil, errors.New("db query returned not enough items to score: " + fmt.Sprintf("%d < %d", len(items), t.GoalQuantity))
	}
	badFields := map[string]string{}
	updateItems := []string{}
	numToAward := t.CurItemsCompleted / t.GoalQuantity
	// score n items. They should be already sorted by create time
	for i := range t.GoalQuantity * numToAward {
		updateItems = append(updateItems, items[i].GetID())
	}

	tr := true
	for _, id := range updateItems {
		err := s.updateJobAppItemFields(t, id, &JobAppItemUpdateFields{
			AttributionTime: &n,
			IsAttributed:    &tr,
		})
		if err != nil {
			badFields[id] = err.Error()
		}
	}
	if len(badFields) > 0 {
		err := errors.New("update error:")
		for id, v := range badFields {
			err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
		}
		return nil, err
	}
	curCompleted := t.CurItemsCompleted - t.GoalQuantity*numToAward
	curAwarded := t.CurBoxesAwarded + numToAward
	// TODO: fix curStreak and maxStreak
	// curStreak := time.Since(*t.FirstCompleted).Hours() / 24
	// update tracker fields
	return s.UpdateJobAppTrackerFields(t.GetUserID(), &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CurItemsCompleted: &curCompleted,
			CurBoxesAwarded:   &curAwarded,
		}})
}

// CreateJobAppItem creates a new job app item and increments the tracker's CurItemsCompleted field by 1
func (s *Service) CreateJobAppItem(userID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker; return immediately on invalid id
	t, err := s.repo.LookupJobAppTrackerFromUserID(userID)
	if err != nil {
		return nil, err
	}
	badFields := map[string]string{}
	if fields.Title == nil {
		badFields[titleField] = "missing item name"
	}
	if fields.Status == nil {
		badFields[statusField] = "missing item status"
	}
	if fields.IsAttributed == nil {
		badFields[isAttributedField] = "missing isAttributed"
	} else if *fields.IsAttributed != false {
		badFields[isAttributedField] = "cannot create an attributed item"
	}
	if fields.AttributionTime != nil {
		badFields[attributionTimeField] = "cannot create an attributed item"
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

	if *fields.Status == StatusComplete && *fields.IsAttributed == false {
		t.CurItemsCompleted = t.CurItemsCompleted + 1
		updateTracker, fields, err := s.ValidateCounters(t)
		if err != nil {
			return nil, err
		}
		_, err = s.repo.UpdateJobAppTrackerFields(updateTracker, fields)
		if err != nil {
			return nil, err
		}
	}
	return s.repo.GetJobAppTrackerWithItemsFromUserID(userID)
}

// TODO: remove this method if not needed
func (s *Service) LookupJobAppItems(trackerID string) ([]*JobAppItem, error) {
	return s.repo.LookupJobAppItems(trackerID)
}

func (s *Service) updateJobAppItemFields(t *JobAppTracker, itemID string, fields *JobAppItemUpdateFields) error {

	// validate item belongs to tracker
	_, err := s.repo.LookupJobAppItem(t.GetID(), itemID)
	if err != nil {
		return err
	}

	// I cant think of a reason to validate any fields
	updateItem, updateFields, err := fields.formatForRepo()
	if err != nil {
		return err
	}
	updateItem.ID = itemID

	_, err = s.repo.UpdateJobAppTrackerItemFields(updateItem, updateFields)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateJobAppItemFields(uuid string, itemID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker
	t, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return nil, err
	}
	// valid item
	repoItem, err := s.repo.LookupJobAppItem(t.GetID(), itemID)
	if err != nil {
		return nil, err
	}

	// I cant think of a reason to validate any fields
	updateItem, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	updateItem.ID = itemID

	_, err = s.repo.UpdateJobAppTrackerItemFields(updateItem, updateFields)
	if err != nil {
		return nil, err
	}

	// user never manually updates IsAttributed, so this check should be sufficient as orchestrator logic
	if !repoItem.IsAttributed {
		// modify tracker curItemsCompleted as needed
	}
	return s.repo.GetJobAppTrackerWithItemsFromUserID(uuid)
}

func (s *Service) DeleteJobAppItem(uuid string, itemID string) error {
	// validate item belongs to tracker
	t, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return err
	}
	repoItem, err := s.repo.LookupJobAppItem(t.GetID(), itemID)
	if err != nil {
		return err
	}
	if !repoItem.IsAttributed {
		// decrement tracker curItemsCompleted
		t.CurItemsCompleted = max(0, t.CurItemsCompleted-1)
		updateTracker, fields, err := s.ValidateCounters(t)
		if err != nil {
			return err
		}
		_, err = s.repo.UpdateJobAppTrackerFields(updateTracker, fields)
		if err != nil {
			return err
		}
	}

	return s.repo.DeleteJobAppTrackerItem(&JobAppItem{ID: itemID})
}

// func (s *Service) ReceiveNewJobAppItem(userId string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
// 	tracker, err := s.CreateJobAppItem(userId, fields)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return s.ScoreTracker(tracker)
// }

// func (s *Service) ReceiveNewJobAppItemUpdate(userId string, itemID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
// 	tracker, err := s.UpdateJobAppItemFields(userId, itemID, fields)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return s.ScoreTracker(tracker)
// }

// StartTrackerUpdateListener is run as a goroutine to reset trackers as specified by the scheduler
func (s *Service) StartTrackerUpdateListener() {
	for tg := range s.scheduler.OutputCh {
		s.updateUnderlyingTrackerFields(tg)
	}
}

func (s *Service) updateUnderlyingTrackerFields(tg *scheduler.TrackerGoal) {
	t, fields, err := toUpdateFields(tg).formatForRepo()
	if err != nil {
		// try fail gracefully
		log.Print(err)
		// TODO: better to not assume without a check
		t.GoalFrequency = scheduler.DefaultFreq
	}
	// find tracker type
	tt, err := s.LookupJobAppTrackerFromTrackerID(tg.TrackerID)
	if err != nil {
		log.Print(err)
		return
	}
	if tt.TrackerType == "" {
		log.Print("tracker type not found")
		return
	}
	progressFields := t.ResetCurrentProgress()
	fields = append(fields, progressFields...)
	t.ID = tg.TrackerID
	t.TrackerType = tt.TrackerType
	err = s.repo.updateUnderlyingTrackerFields(t, fields) // <-- ideally one size fits all tracker types
	if err != nil {
		log.Print(err)
		return
	}
}

func toUpdateFields(tg *scheduler.TrackerGoal) *UnderlyingTrackerUpdateFields {
	return &UnderlyingTrackerUpdateFields{
		GoalDeadline:  &tg.GoalDeadline,
		GoalFrequency: (*string)(&tg.GoalFrequency),
	}
}

// ResetCurrentProgress is a general tracker reset function that may or may not need to be refactored
// it modifies the tracker struct in place and returns a list of fields that were modified
// I might need to expand to typed functions if different tracker types have different reset needs
func (t *UnderlyingTracker) ResetCurrentProgress() []string {
	t.CurItemsCompleted = 0
	t.CurBoxesAwarded = 0
	return []string{curItemsCompletedField, curBoxesAwardedField}
	// also update stats as needed
}
