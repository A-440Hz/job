package tracker

type ItemStatus string

const (
	StatusComplete   ItemStatus = "complete"
	StatusInProgress ItemStatus = "in progress"
)

type Item interface {
	IsComplete() bool
	EditStatus(s ItemStatus) error
}

type JobAppItem struct {
	ItemId     string
	Title      string
	Body       string
	status     ItemStatus
	attributed bool // this flips when a tracker progress is assigned from this item
	// probably sync shenanigans to iron out? attempt sync on each item creation?
}

func (i *JobAppItem) IsComplete() bool {
	return i.status == StatusComplete
}

func (i *JobAppItem) EditStatus(s ItemStatus) {
	// noop (job apps are always complete in current design)
	return
}

func newJobAppItem()
