package tracker

type Tracker interface {
	// getGoal()
	EditGoal() error
	AddItem() error
	EditItem() error
}

type JobAppTracker struct {
	TrackerId         string
	numBoxesAwarded   int
	numItemsCompleted int
}

func (t *JobAppTracker) EditGoal() {}

func (t *JobAppTracker) ResetProgress() {
	t.numBoxesAwarded = 0
	t.numItemsCompleted = 0
}
