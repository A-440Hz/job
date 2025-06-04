package tracker

import (
	"errors"
	"fmt"
	"job/internal/collection"
	"job/internal/scheduler"
	"job/internal/user"
	"log"
	"slices"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	repo       *Repository
	scheduler  *scheduler.Scheduler
	collection *collection.Service
}

func NewService(r *Repository, s *scheduler.Scheduler, c *collection.Service) *Service {
	return &Service{repo: r, scheduler: s, collection: c}
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
	t, err := s.repo.createJobAppTracker(t)
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
	return s.repo.lookupJobAppTrackerFromUserID(uuid)
}

// LookupJobAppTrackerFromTrackerID is probably needed for the scheduler to trigger tracker
// func (s *Service) LookupJobAppTrackerFromTrackerID(id string) (*JobAppTracker, error) {
// 	return s.repo.lookupJobAppTrackerFromTrackerID(id)
// }

func (s *Service) GetJobAppTrackerWithItemsFromUserID(uuid string) (*JobAppTracker, error) {
	return s.repo.getJobAppTrackerWithItemsFromUserID(uuid)
}

func (s *Service) UpdateJobAppTrackerFields(uuid string, fields *JobAppTrackerUpdateFields) (*JobAppTracker, error) {
	repoTracker, err := s.repo.lookupJobAppTrackerFromUserID(uuid)
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
	if slices.Contains(updateFields, maxCycleItemsCompletedField) && updateTracker.MaxCycleItemsCompleted < repoTracker.MaxCycleItemsCompleted {
		badFields[maxCycleItemsCompletedField] = fmt.Sprintf("cannot decrease %q", maxCycleItemsCompletedField)
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
	updateTracker, err = s.repo.updateJobAppTrackerFields(updateTracker, updateFields)
	if err != nil {
		return nil, err
	}

	// communicate with scheduler if needed.. updateFields was trimmed earlier if this step is not needed
	if slices.Contains(updateFields, cycleDeadlineField) || slices.Contains(updateFields, cycleFrequencyField) {
		// ReplaceTrackerGoal allows for nil deadline but not nil frequency
		if !slices.Contains(updateFields, cycleFrequencyField) {
			updateTracker.CycleFrequency = repoTracker.CycleFrequency
		}
		updateTracker.TrackerType = repoTracker.TrackerType
		err = s.scheduler.ReplaceTrackerGoal(repoTracker.ToTrackerGoal(), updateTracker.ToTrackerGoal())
		if err != nil {
			// fail scheduler gracefully?
			log.Print("scheduler replace failed:", err)
		}
	}

	// validate counters and score tracker if needed (technically curScorableItemsField should never be manually updated by the service update function)
	if slices.Contains(updateFields, goalQuantityField) || slices.Contains(updateFields, curScorableItemsField) {
		return s.updateTrackerState(repoTracker.GetID())
	}

	return s.repo.getJobAppTrackerWithItemsFromUserID(uuid)
}

func (s *Service) DeleteJobAppTracker(id string) error {
	return s.repo.deleteJobAppTracker(&JobAppTracker{UnderlyingTracker: UnderlyingTracker{ID: id}})
}

// CreateJobAppItem creates a new job app item and increments the tracker's CurScorableItems field by 1
func (s *Service) CreateJobAppItem(userID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker; return immediately on invalid id
	t, err := s.repo.lookupJobAppTrackerFromUserID(userID)
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
	i, err = s.repo.createJobAppTrackerItem(i)
	if err != nil {
		return nil, err
	}

	if i.IsScorable() {
		return s.addOneScorableItem(t)
	}
	return s.repo.getJobAppTrackerWithItemsFromUserID(userID)
}

// TODO: remove this method if not needed
func (s *Service) LookupJobAppItems(trackerID string) ([]*JobAppItem, error) {
	return s.repo.lookupJobAppItems(trackerID)
}

// UpdateJobAppItemFields handles user submitted update requests for job app items
func (s *Service) UpdateJobAppItemFields(uuid string, itemID string, fields *JobAppItemUpdateFields) (*JobAppTracker, error) {
	// validate tracker
	t, err := s.repo.lookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return nil, err
	}
	// validate item
	repoItem, err := s.repo.lookupJobAppItem(t.GetID(), itemID)
	if err != nil {
		return nil, err
	}

	// I cant think of a reason to validate any fields
	updateItem, updateFields, err := fields.formatForRepo()
	if err != nil {
		return nil, err
	}
	updateItem.ID = itemID
	_, err = s.repo.updateJobAppTrackerItemFields(updateItem, updateFields)
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
	return s.repo.getJobAppTrackerWithItemsFromUserID(uuid)
}

func (s *Service) DeleteJobAppItem(uuid string, itemID string) error {
	// validate item belongs to tracker
	t, err := s.repo.lookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		return err
	}
	repoItem, err := s.repo.lookupJobAppItem(t.GetID(), itemID)
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
	return s.repo.deleteJobAppTrackerItem(&JobAppItem{ID: itemID})
}

// Tracker scoring-related methods:

// updateTrackerState checks for overflow of CurScorableItems and goalQuantity and adjusts
// curBoxesAwarded and CurScorableItems, performing a repo update if needed.
func (s *Service) updateTrackerState(tid string) (*JobAppTracker, error) {
	// I went back and forth on this for a while but I decided to just go with a lot (+2) of lookup calls
	// and have this function be easier to use (just require a valid id)
	repoTracker, err := s.repo.getJobAppTrackerWithItemsFromTrackerID(tid)
	if err != nil {
		return nil, err
	}
	if repoTracker.CurScorableItems < repoTracker.GoalQuantity {
		return repoTracker, nil
	}

	// check if tracker needs to be scored; exit early if not
	n := time.Now()
	numToAward := repoTracker.CurScorableItems / repoTracker.GoalQuantity
	items, err := s.repo.getScorableJobAppItems(repoTracker)
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
		repoTracker.CurCycleItemsCompleted = repoTracker.CurCycleItemsCompleted + numItems
		updateFields = append(updateFields, curCycleItemsCompleted)
		repoTracker.TotalItemsCompleted = repoTracker.TotalItemsCompleted + numItems
		updateFields = append(updateFields, totalItemsCompletedField)
	}
	if repoTracker.CurCycleItemsCompleted > repoTracker.MaxCycleItemsCompleted {
		repoTracker.MaxCycleItemsCompleted = repoTracker.CurCycleItemsCompleted
		updateFields = append(updateFields, maxCycleItemsCompletedField)
	}

	// update collection lootbox count
	_, err = s.collection.AssignBoxes(repoTracker.GetUserID(), numToAward)
	if err != nil {
		log.Print(err)
	}

	// update tracker state
	repoTracker.CurScorableItems = max(0, repoTracker.CurScorableItems-repoTracker.GoalQuantity*numToAward)
	repoTracker.CurBoxesAwarded = repoTracker.CurBoxesAwarded + numToAward
	repoTracker.TotalBoxesAwarded = repoTracker.TotalBoxesAwarded + numToAward
	_, err = s.repo.updateJobAppTrackerFields(repoTracker, updateFields)
	if err != nil {
		return nil, err
	}
	return s.repo.getJobAppTrackerWithItemsFromTrackerID(repoTracker.GetID())
}

func (s *Service) updateJobAppItemFields(t *JobAppTracker, itemID string, fields *JobAppItemUpdateFields) error {
	// validate item belongs to tracker
	_, err := s.repo.lookupJobAppItem(t.GetID(), itemID)
	if err != nil {
		return err
	}
	// I cant think of a reason to validate any fields
	updateItem, updateFields, err := fields.formatForRepo()
	if err != nil {
		return err
	}
	updateItem.ID = itemID
	_, err = s.repo.updateJobAppTrackerItemFields(updateItem, updateFields)
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
	fields := []string{curScorableItemsField}
	// populate FirstCompleted to track averages
	if t.FirstCompleted == nil {
		n := time.Now()
		t.FirstCompleted = &n
		fields = append(fields, firstCompletedField)
	}
	_, err := s.repo.updateJobAppTrackerFields(t, fields)
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
	_, err := s.repo.updateJobAppTrackerFields(t, []string{curScorableItemsField})
	if err != nil {
		return nil, err
	}
	return s.updateTrackerState(t.GetID())
}

// Scheduler-related methods:

func (s *Service) Start() {
	go s.scheduler.Start()
	go s.startTrackerUpdateListener()
	s.MigrateTrackersIntoScheduler()
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
func (s *Service) resetTrackerDeadline(poppedTg *scheduler.TrackerGoal) error {
	updateTracker, updateFields, err := getCycleUpdateFields(poppedTg).formatForRepo()
	if err != nil {
		return err
	}
	newCycleStartTime, _ := updateTracker.GetValidTimeframe()
	// mark current items unscorable and return a repo tracker to manage stats
	repoTracker, curScorable, err := s.clearTrackerItems(poppedTg, newCycleStartTime)
	if err != nil {
		return err
	}

	// reset goal streak if no boxes earned last cycle
	if repoTracker.CurCycleItemsCompleted < repoTracker.GoalQuantity {
		updateTracker.CurGoalStreak = 0
		updateFields = append(updateFields, curGoalStreakField)

		// Take away a lootbox or a reward if user enabled this setting
		if repoTracker.MissedGoalPenalty {
			// TODO: can wrap a more complex penalty function here; maybe reduce boxes or take away collectable or take away coins
			_, err := s.collection.AssignBoxes(repoTracker.UserID, -1)
			if err != nil {
				return fmt.Errorf("unable to apply penalty on %q failed tracker %q goal: %w", repoTracker.UserID, repoTracker.ID, err)
			}
		}
	}

	// reset tracker cycle counters
	updateFields = append(updateFields, curScorableItemsField, curBoxesAwardedField, curCycleItemsCompleted)
	updateTracker.CurScorableItems = curScorable
	updateTracker.CurBoxesAwarded = 0
	updateTracker.CurCycleItemsCompleted = 0

	updateTracker.ID = poppedTg.TrackerID
	updateTracker.TrackerType = TrackerType(poppedTg.TrackerType)

	err = s.repo.updateUnderlyingTrackerFields(updateTracker, updateFields)
	if err != nil {
		return err
	}
	return nil
}

// getCycleUpdateFields extracts the updated Deadline and Frequency from a scheduler TrackerGoal,
// returning it in the form of UpdateFields for an update function
func getCycleUpdateFields(tg *scheduler.TrackerGoal) *UnderlyingTrackerUpdateFields {
	return &UnderlyingTrackerUpdateFields{
		CycleDeadline:  &tg.CycleDeadline,
		CycleFrequency: tg.CycleFrequency.StrPtr(),
	}
}

// clearTrackerItems marks all current scorable tracker items as attributed to prepare for progressing to next tracker cycle
// it also returns a repo UnderlyingTracker and int curScorable for scoring purposes
func (s *Service) clearTrackerItems(tg *scheduler.TrackerGoal, newCycleStartTime time.Time) (*UnderlyingTracker, int, error) {
	ut := &UnderlyingTracker{}
	curScorable := 0
	switch TrackerType(tg.TrackerType) {
	case JobAppTrackerType:
		// try a repo lookup. Delete from scheduler if RecordNotFound
		// on paper this avoids deleting valid trackers from a false positive and clogging the scheduler with ghost trackers
		repoTracker, err := s.repo.lookupJobAppTrackerFromTrackerID(tg.TrackerID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// remove tracker goal from scheduler
				s.scheduler.ReplaceTrackerGoal(tg, nil)
				return nil, 0, fmt.Errorf("tracker record not found during scheduler reset. Removing tracker goal %q from scheduler.", tg.TrackerID)
			} else {
				// TODO: maybe add a counter to the tracker goal and get rid of it after 3 failed cycles
				// (and remember to clear it on success)? very low priority unless the scehduler starts failing here
				log.Print("error looking up tracker during scheduler reset:", err)
			}
		}

		// mark all scorable items created after the new cycle start as attributed before progressing to next tracker cycle
		n := time.Now()
		tr := true
		badFields := map[string]string{}
		items, err := s.repo.getScorableJobAppItems(repoTracker)
		if err != nil {
			return nil, 0, err
		}
		for _, item := range items {
			if item.CreatedAt.After(newCycleStartTime) {
				curScorable++
				continue
			}
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
			return nil, 0, err
		}
		ut = &repoTracker.UnderlyingTracker
	default:
		return nil, 0, errors.New("invalid tracker type")
	}
	return ut, curScorable, nil
}

// MigrateTrackersIntoScheduler converts repo trackers into tracker goals and loads them into the scheduler.
// The purpose of this function is to allow tracker resets to keep working after taking down the server and spinning it back up, without user input
func (s *Service) MigrateTrackersIntoScheduler() {
	// before deploying to production I should to decide on the scope of trackers I will pull
	// or look into having a cron job to delete expired users/trackers
	// it's pretty easy to have a large number of objects created, espceially if i post a url on a forum or a chat
	// users can also spam refresh a page while clearing cookies to create a lot of objects
	// TODO: smart rate limiting somehow
	// it feels like the line between something that theoretically works and theoretically doesnt work is pretty thin at times

	// get all valid trackers from repo
	allUnderlying, err := s.repo.getAllUnderlyingTrackers()
	if err != nil {
		log.Print("getAllUnderlyingTrackers failed:", err)
	}
	for _, ut := range allUnderlying {
		// plug them in. scheduler will automatically scale them to the next upcoming deadline
		if err = s.scheduler.AddTrackerGoal(ut.ToTrackerGoal()); err != nil {
			log.Print(err)
		}
	}
}

// SelectAllJobAppTrackers is used for testing
func (s *Service) SelectAllJobAppTrackers() ([]JobAppTracker, error) {
	return s.repo.selectAllJobAppTrackers()
}
