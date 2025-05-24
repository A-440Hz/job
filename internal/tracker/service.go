package tracker

import (
	"errors"
	"fmt"
	"job/internal/scheduler"
	"job/internal/user"
	"log"
	"slices"
	"time"

	"gorm.io/gorm"
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
		UserID:         u.GetID(),
		CycleDeadline:  scheduler.GetDefaultCycleDeadline(u.Timezone),
		CycleFrequency: scheduler.GetDefaultCycleFrequency(),
		TrackerType:    JobAppTrackerType,
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
// func (s *Service) LookupJobAppTrackerFromTrackerID(id string) (*JobAppTracker, error) {
// 	return s.repo.lookupJobAppTrackerFromTrackerID(id)
// }

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
		for id, v := range badFields {
			err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
		}
		return nil, err
	}

	// the front end should also avoid sending update reqs for identical deadlines and frequencies
	if repoTracker.CycleDeadline == updateTracker.CycleDeadline {
		//
		updateFields = slices.DeleteFunc(updateFields, func(f string) bool {
			return f == cycleDeadlineField
		})
	}
	if repoTracker.CycleFrequency == updateTracker.CycleFrequency {
		updateFields = slices.DeleteFunc(updateFields, func(f string) bool {
			return f == cycleFrequencyField
		})
	}

	// push user-settable field updates to repo
	updateTracker.ID = repoTracker.GetID()
	updateTracker, err = s.repo.UpdateJobAppTrackerFields(updateTracker, updateFields)
	if err != nil {
		return nil, err
	}

	// communicate with scheduler if needed.. updateFields was trimmed earlier if this step is not needed
	if slices.Contains(updateFields, cycleDeadlineField) || slices.Contains(updateFields, cycleFrequencyField) {
		// fail scheduler gracefully?
		updateTracker.TrackerType = repoTracker.TrackerType
		err = s.scheduler.ReplaceTrackerGoal(repoTracker.ToTrackerGoal(), updateTracker.ToTrackerGoal())
		if err != nil {
			log.Print(err)
		}
	}

	// validate counters and score tracker if needed (technically curScorableItemsField should never be manually updated by the service update function)
	if slices.Contains(updateFields, goalQuantityField) || slices.Contains(updateFields, curScorableItemsField) {
		return s.updateTrackerState(repoTracker.GetID())
	}

	return s.repo.GetJobAppTrackerWithItemsFromUserID(uuid)
}

func (s *Service) DeleteJobAppTracker(id string) error {
	return s.repo.DeleteJobAppTracker(&JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}})
}

// CreateJobAppItem creates a new job app item and increments the tracker's CurScorableItems field by 1
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
	if fields.attributionTime != nil {
		badFields[attributionTimeField] = "cannot create an attributed item"
	}
	i, _, err := fields.formatForRepo()
	if err != nil {
		badFields["format"] = err.Error()
	}
	if len(badFields) > 0 {
		err := errors.New("validation error:")
		for id, v := range badFields {
			err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
		}
		return nil, err
	}

	i.TrackerID = t.GetID()
	i, err = s.repo.CreateJobAppTrackerItem(i)
	if err != nil {
		return nil, err
	}

	if i.IsScorable() {
		return s.addOneScorableItem(t)
	}
	return s.repo.GetJobAppTrackerWithItemsFromUserID(userID)
}

// TODO: remove this method if not needed
func (s *Service) LookupJobAppItems(trackerID string) ([]*JobAppItem, error) {
	return s.repo.LookupJobAppItems(trackerID)
}

// UpdateJobAppItemFields handles user submitted update requests for job app items
func (s *Service) UpdateJobAppItemFields(uuid string, itemID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker
	t, err := s.repo.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return nil, err
	}
	// validate item
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

	if repoItem.IsScorable() {
		// decrement tracker CurScorableItems if this update makes it no longer scorable
		if slices.Contains(updateFields, statusField) && updateItem.Status != StatusComplete {
			return s.subOneScorableItem(t)
		}
	} else if !repoItem.IsAttributed && slices.Contains(updateFields, statusField) && updateItem.Status == StatusComplete {
		// increment it if this update makes the item scorable
		return s.addOneScorableItem(t)
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
	// decrement tracker scorable item count if needed
	if repoItem.IsScorable() {
		_, err = s.subOneScorableItem(t)
		if err != nil {
			return err
		}
	}
	return s.repo.DeleteJobAppTrackerItem(&JobAppItem{ID: itemID})
}

// Tracker scoring-related methods:

// updateTrackerState checks for overflow of CurScorableItems and goalQuantity and adjusts
// curBoxesAwarded and CurScorableItems, performing a repo update if needed.
func (s *Service) updateTrackerState(tid string) (*JobAppTracker, error) {
	// I went back and forth on this for a while but I decided to just go with a lot (+2) of lookup calls
	// and have this function be easier to use
	repoTracker, err := s.repo.GetJobAppTrackerWithItemsFromTrackerID(tid)
	if err != nil {
		return nil, err
	}
	if repoTracker.CurScorableItems < repoTracker.GoalQuantity {
		return repoTracker, nil
	}

	// check if tracker needs to be scored; exit early if not
	n := time.Now()
	numToAward := repoTracker.CurScorableItems / repoTracker.GoalQuantity
	items, err := s.repo.GetScorableJobAppItems(repoTracker)
	if err != nil {
		return nil, err
	}
	if len(items) < repoTracker.GoalQuantity*numToAward {
		log.Print("mismatch between number of scorable items and tracker current items completed: " + fmt.Sprintf("%d < %d", len(items), repoTracker.CurScorableItems))
		// is it good to set curScorableItems to len(items) here?
		return nil, errors.New("mismatch between number of scorable items and tracker current items completed: " + fmt.Sprintf("%d < %d", len(items), repoTracker.CurScorableItems))
	}

	// score n items. They should be already sorted by create time
	updateItems := []string{}
	for i := range repoTracker.GoalQuantity * numToAward {
		updateItems = append(updateItems, items[i].GetID())
	}
	tr := true
	badFields := map[string]string{}
	for _, id := range updateItems {
		err := s.updateJobAppItemFields(repoTracker, id, &JobAppItemUpdateFields{
			attributionTime: &n,
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

	updateFields := []string{curScorableItemsField, curBoxesAwardedField, totalBoxesAwardedField}
	// increment CurGoalStreak if first box of the cycle
	if repoTracker.CurBoxesAwarded == 0 && numToAward > 0 {
		repoTracker.CurGoalStreak = repoTracker.CurGoalStreak + 1
		updateFields = append(updateFields, curGoalStreakField)
	}
	// increment MaxGoalStreak if needed
	if repoTracker.CurGoalStreak > repoTracker.MaxGoalStreak {
		repoTracker.MaxGoalStreak = repoTracker.CurGoalStreak
		updateFields = append(updateFields, maxGoalStreakField)
	}
	// increment item counter stats
	if numItems := len(updateItems); numItems > 0 {
		repoTracker.CurItemsCompletedDaily = repoTracker.CurItemsCompletedDaily + numItems
		updateFields = append(updateFields, curItemsCompletedDailyField)
		repoTracker.TotalItemsCompleted = repoTracker.TotalItemsCompleted + numItems
		updateFields = append(updateFields, totalItemsCompletedField)
	}
	if repoTracker.CurItemsCompletedDaily > repoTracker.MaxItemsCompletedDaily {
		repoTracker.MaxItemsCompletedDaily = repoTracker.CurItemsCompletedDaily
		updateFields = append(updateFields, maxItemsCompletedDailyField)
	}

	// update tracker state
	repoTracker.CurScorableItems = max(0, repoTracker.CurScorableItems-repoTracker.GoalQuantity*numToAward)
	repoTracker.CurBoxesAwarded = repoTracker.CurBoxesAwarded + numToAward
	repoTracker.TotalBoxesAwarded = repoTracker.TotalBoxesAwarded + numToAward
	_, err = s.repo.UpdateJobAppTrackerFields(repoTracker, updateFields)
	if err != nil {
		return nil, err
	}
	return s.repo.GetJobAppTrackerWithItemsFromTrackerID(repoTracker.GetID())
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

// addOneScorableItem increments the CurScorableItems count for the tracker
// t needs a valid tracker ID and accurate CurScorableItems count (to be incremented by 1)
// it returns the updated tracker with all its items
func (s *Service) addOneScorableItem(t *JobAppTracker) (*JobAppTracker, error) {
	t.CurScorableItems = t.CurScorableItems + 1
	_, err := s.repo.UpdateJobAppTrackerFields(t, []string{curScorableItemsField})
	if err != nil {
		return nil, err
	}
	return s.updateTrackerState(t.GetID())
}

// subOneScorableItem decrements the CurScorableItems count for the tracker
// t needs a valid tracker ID and accurate CurScorableItems count (to be decremented by 1)
// it returns the updated tracker with all its items
func (s *Service) subOneScorableItem(t *JobAppTracker) (*JobAppTracker, error) {
	t.CurScorableItems = max(0, t.CurScorableItems-1)
	_, err := s.repo.UpdateJobAppTrackerFields(t, []string{curScorableItemsField})
	if err != nil {
		return nil, err
	}
	return s.updateTrackerState(t.GetID())
}

// Scheduler-related methods:

func (s *Service) Start() {
	go s.scheduler.Start()
	go s.startTrackerUpdateListener()
}

func (s *Service) Stop() {
	s.scheduler.Stop()
}

// startTrackerUpdateListener is run as a goroutine to reset trackers as specified by the scheduler
func (s *Service) startTrackerUpdateListener() {
	for tg := range s.scheduler.OutputCh {
		err := s.resetTrackerDeadline(tg)
		if err != nil {
			log.Print(err)
		}
	}
}

// resetTrackerDeadline is a switch statement that queries the tracker type and calls the appropriate update method
func (s *Service) resetTrackerDeadline(newTg *scheduler.TrackerGoal) error {
	// update deadline and frequency TODO: refactor names to Cycle
	updateTracker, updateFields, err := getUpdateFields(newTg).formatForRepo()
	if err != nil {
		return err
	}

	// mark current items unscorable and return a repo tracker to count stats
	repoTracker, err := s.clearTrackerItems(newTg)
	if err != nil {
		return err
	}

	// reset goal streak if no boxes earned last cycle
	if repoTracker.CurItemsCompletedDaily < repoTracker.GoalQuantity {
		updateTracker.CurGoalStreak = 0
		updateFields = append(updateFields, curGoalStreakField)

		// TODO: here is also where the tracker would tell the collection to take away a reward (or unopened box?)
		// if the user enabled that tracker setting
	}

	// reset tracker cycle counters
	updateFields = append(updateFields, curScorableItemsField, curBoxesAwardedField, curItemsCompletedDailyField)
	updateTracker.CurScorableItems = 0
	updateTracker.CurBoxesAwarded = 0
	updateTracker.CurItemsCompletedDaily = 0

	updateTracker.ID = newTg.TrackerID
	updateTracker.TrackerType = TrackerType(newTg.TrackerType)

	err = s.repo.updateUnderlyingTrackerFields(updateTracker, updateFields)
	if err != nil {
		return err
	}
	return nil
}

// getUpdateFields extracts the updated Deadline and Frequency from a scheduler TrackerGoal,
// returning it in the form of UpdateFields for an update function
func getUpdateFields(tg *scheduler.TrackerGoal) *UnderlyingTrackerUpdateFields {
	return &UnderlyingTrackerUpdateFields{
		CycleDeadline:  &tg.CycleDeadline,
		CycleFrequency: tg.CycleFrequency.StrPtr(),
	}
}

// clearTrackerItems marks all current scorable tracker items as attributed to prepare for progressing to next tracker cycle
// it also returns a repo UnderlyingTracker for scoring purposes
func (s *Service) clearTrackerItems(tg *scheduler.TrackerGoal) (*UnderlyingTracker, error) {
	ut := &UnderlyingTracker{}
	switch TrackerType(tg.TrackerType) {
	case JobAppTrackerType:
		// try a repo lookup. Delete from scheduler if RecordNotFound
		// on paper this avoids deleting valid trackers from a false positive and clogging the scheduler with ghost trackers
		repoTracker, err := s.repo.lookupJobAppTrackerFromTrackerID(tg.TrackerID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// remove tracker goal from scheduler
				s.scheduler.ReplaceTrackerGoal(tg, nil)
				return nil, fmt.Errorf("tracker record not found during scheduler reset. Removing tracker goal %q from scheduler.", tg.TrackerID)
			} else {
				// TODO: maybe add a counter to the tracker goal and get rid of it after 3 failed cycles
				// (and remember to clear it on success)
				log.Print("error looking up tracker during scheduler reset:", err)
			}
		}

		// mark all current scorable items as attributed before progressing to next tracker cycle
		n := time.Now()
		tr := true
		badFields := map[string]string{}
		items, err := s.repo.GetScorableJobAppItems(repoTracker)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			err := s.updateJobAppItemFields(repoTracker, item.GetID(), &JobAppItemUpdateFields{
				attributionTime: &n,
				IsAttributed:    &tr,
			})
			if err != nil {
				badFields[item.GetID()] = err.Error()
			}
		}
		if len(badFields) > 0 {
			err := errors.New("update error:")
			for id, v := range badFields {
				err = errors.Join(err, fmt.Errorf("%q: %q, ", id, v))
			}
			return nil, err
		}
		ut = &repoTracker.UnderlyingTracker
	default:
		return nil, errors.New("invalid tracker type")
	}
	return ut, nil
}
