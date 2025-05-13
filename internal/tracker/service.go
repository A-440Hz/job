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
	t, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
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

	// stats invariants
	if slices.Contains(updateFields, maxGoalStreakField) && updateTracker.MaxGoalStreak < t.MaxGoalStreak {
		badFields[maxGoalStreakField] = fmt.Sprintf("cannot decrease %q", maxGoalStreakField)
	}
	if slices.Contains(updateFields, maxItemsCompletedDailyField) && updateTracker.MaxItemsCompletedDaily < t.MaxItemsCompletedDaily {
		badFields[maxItemsCompletedDailyField] = fmt.Sprintf("cannot decrease %q", maxItemsCompletedDailyField)
	}
	if slices.Contains(updateFields, totalItemsCompletedField) && updateTracker.TotalItemsCompleted < t.TotalItemsCompleted {
		badFields[totalItemsCompletedField] = fmt.Sprintf("cannot decrease %q", totalItemsCompletedField)
	}
	if slices.Contains(updateFields, totalBoxesAwardedField) && updateTracker.TotalBoxesAwarded < t.TotalBoxesAwarded {
		badFields[totalBoxesAwardedField] = fmt.Sprintf("cannot decrease %q", totalBoxesAwardedField)
	}
	if slices.Contains(updateFields, firstCompletedField) && t.FirstCompleted != nil && updateTracker.FirstCompleted.After(*t.FirstCompleted) {
		badFields[firstCompletedField] = fmt.Sprintf("cannot increment %q", firstCompletedField)
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for _, v := range badFields {
			err = errors.Join(err, errors.New(v))
		}
		return nil, err
	}

	beforeTg := t.ToTrackerGoal()
	// TODO: try to prevent front end from sending update reqs for identical deadlines and frequencies
	if t.GoalDeadline == updateTracker.GoalDeadline {
		//
		updateFields = slices.DeleteFunc(updateFields, func(f string) bool {
			return f == goalDeadlineField
		})
	}
	if t.GoalFrequency == updateTracker.GoalFrequency {
		updateFields = slices.DeleteFunc(updateFields, func(f string) bool {
			return f == goalFrequencyField
		})
	}

	// validate?
	updateTracker.ID = t.GetID()
	_, err = s.repo.UpdateJobAppTrackerFields(updateTracker, updateFields)
	if err != nil {
		return nil, err
	}

	// communicate with scheduler if needed.. updateFields is correctly trimmed if no diffs are present
	if slices.Contains(updateFields, goalDeadlineField) || slices.Contains(updateFields, goalFrequencyField) {
		// fail scheduler gracefully?
		if err = s.scheduler.Update(beforeTg, updateTracker.ToTrackerGoal()); err != nil {
			log.Print(err)
		}
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

// GetScorableJobAppItems returns a list of job app items with create time within the tracker's goal timeframe
// and are StatusComplete and not yet attributed
func (s *Service) GetScorableJobAppItems(t *JobAppTracker) ([]JobAppItem, error) {
	return s.repo.GetScorableJobAppItems(t)
}

// ScoreTracker updates the tracker items and tracker fields if the goal quantity is reached,
// otherwise it returns the tracker as is
func (s *Service) ScoreTracker(t *JobAppTracker) (*JobAppTracker, error) {
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
	numToAward := t.CurItemsCompleted % t.GoalQuantity
	// score n items. They should be already sorted by create time
	for i := range t.GoalQuantity * numToAward {
		updateItems = append(updateItems, items[i].GetID())
	}

	tr := true
	st := StatusComplete
	for _, id := range updateItems {
		_, err := s.UpdateJobAppItemFields(t.GetID(), id, &JobAppItemUpdateFields{
			Status:          &st,
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
			err = errors.Join(err, errors.New(fmt.Sprintf("%q: %q, ", id, v)))
		}
		return nil, err
	}
	curCompleted := t.CurItemsCompleted - t.GoalQuantity
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
	num := t.CurItemsCompleted + 1
	return s.UpdateJobAppTrackerFields(userID, &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CurItemsCompleted: &num}})
}

// TODO: remove this method if not needed
func (s *Service) LookupJobAppItems(trackerID string) ([]*JobAppItem, error) {
	items, err := s.repo.LookupJobAppItems(trackerID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) UpdateJobAppItemFields(tid string, itemID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker
	t, err := s.repo.lookupJobAppTrackerFromTrackerID(tid)
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
	return s.repo.GetJobAppTrackerWithItemsFromUserID(tid)
}

func (s *Service) ReceiveNewJobAppItem(userId string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	tracker, err := s.CreateJobAppItem(userId, fields)
	if err != nil {
		return nil, err
	}
	// increment CurItemsCompleted
	numItems := tracker.CurItemsCompleted + 1
	tracker, err = s.UpdateJobAppTrackerFields(userId, &JobAppTrackerUpdateFields{
		UnderlyingTrackerUpdateFields: UnderlyingTrackerUpdateFields{
			CurItemsCompleted: &numItems,
		}})
	if err != nil {
		return nil, err
	}
	return s.ScoreTracker(tracker)
}

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
// I might need to expand to typed functions if different tracker types have different reset needs
func (t *UnderlyingTracker) ResetCurrentProgress() []string {
	t.CurItemsCompleted = 0
	t.CurBoxesAwarded = 0
	return []string{curItemsCompletedField, curBoxesAwardedField}
	// also update stats as needed
}
