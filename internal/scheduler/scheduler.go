package scheduler

import (
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"
)

// scaled to the amount of users I expect
const startingHeapCapacity = 32

// Scheduler always lives in memory and manages when to trigger and reset TrackerGoals
type Scheduler struct {
	g        *GoalHeap
	mutex    sync.Mutex
	stopCh   chan bool
	OutputCh chan *TrackerGoal
}

// GoalHeap implements heap.Interface and sort.Interface
type GoalHeap struct {
	heap   []*TrackerGoal
	idxMap map[string]int
}

type TrackerGoal struct {
	TrackerID     string
	GoalDeadline  time.Time
	GoalFrequency Frequency
	index         int
}

func (tg *TrackerGoal) getNext() {
	now := time.Now()
	// a loop should be fine as long as front end prevents setting a deadline a million years back
	// for tg.GoalDeadline.Before(now) {
	// 	tg.GoalDeadline = tg.GoalDeadline.Add(time.Duration(tg.GoalFrequency.NumDays()))
	// }
	if now.Before(tg.GoalDeadline) {
		return
	}
	numDaysBetween := int(now.Sub(tg.GoalDeadline).Hours()) / 24
	numDaysToNext := (numDaysBetween/tg.GoalFrequency.NumDays() + 1) * tg.GoalFrequency.NumDays()
	tg.GoalDeadline = tg.GoalDeadline.Add(time.Hour * 24 * time.Duration(numDaysToNext))
}

func NewScheduler() *Scheduler {
	g := &GoalHeap{
		heap:   make([]*TrackerGoal, 0, startingHeapCapacity),
		idxMap: make(map[string]int),
	}
	heap.Init(g)
	s := &Scheduler{
		g:      g,
		stopCh: make(chan bool),
	}
	return s
}

func (g GoalHeap) Len() int           { return len(g.heap) }
func (g GoalHeap) Less(i, j int) bool { return g.heap[i].GoalDeadline.Before(g.heap[j].GoalDeadline) }
func (g GoalHeap) Swap(i, j int) {
	g.heap[i], g.heap[j] = g.heap[j], g.heap[i]
	g.heap[i].index, g.heap[j].index = i, j
	g.idxMap[g.heap[i].TrackerID], g.idxMap[g.heap[j].TrackerID] = i, j
}

// add x as element Len()
func (g *GoalHeap) Push(x any) {
	n := (*g).Len()
	item := x.(*TrackerGoal)
	item.index = n
	(*g).idxMap[item.TrackerID] = n
	(*g).heap = append((*g).heap, item)
}

// remove and return element Len() - 1
func (g *GoalHeap) Pop() any {
	old := (*g).heap
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	delete(g.idxMap, item.TrackerID)
	(*g).heap = old[0 : n-1]
	return item
}

// findByID returns the index or error if not found.
// I originally wanted to break abstraction and have this be a binary search,
// but keeping a tracker:idx map helps me validate the existence of trackerIds,
// which I need to do anyway, which justifies this method over the cooler binary search
func (g *GoalHeap) findByID(id string) (int, error) {
	if i, exists := g.idxMap[id]; exists {
		return i, nil
	}
	log.Printf("attempt to find tracker %q unsuccessful", id)
	return 0, fmt.Errorf("attempt to find tracker %q unsuccessful", id)

}

// popTrackerGoal finds and pops tracker t
// func (g *GoalHeap) popTrackerGoal(t *TrackerGoal) (*TrackerGoal, error) {
// 	j, err := g.findByID(t.TrackerID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ret := heap.Remove(g, j).(*TrackerGoal)
// 	return ret, nil
// }

// AddTrackerGoal adds a TrackerGoal into the scheduler
func (s *Scheduler) AddTrackerGoal(tg *TrackerGoal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	heap.Push(s.g, tg)
	return nil
}

func (s *Scheduler) Update(oldTg, newTg *TrackerGoal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	idx, err := s.g.findByID(oldTg.TrackerID)
	if err != nil {
		return err
	}

	// update in place if we only need to change GoalFrequency
	if oldTg.GoalDeadline == newTg.GoalDeadline {
		s.g.heap[idx].GoalFrequency = newTg.GoalFrequency
		return nil
	}
	// otherwise pop and replace old TrackerGoal
	_ = heap.Remove(s.g, idx)
	err = s.AddTrackerGoal(newTg)
	if err != nil {
		// try to push back oldTg... this is silly because this function doesnt even return an error
		log.Printf("failed to push tracker %v.. putting back %v", newTg, oldTg)
		restoreErr := s.AddTrackerGoal(oldTg) // and then what if this one errors again
		if restoreErr != nil {
			log.Printf("failed to push back previous tracker %v", oldTg)
		}
		return fmt.Errorf("failed to push tracker %v.. putting back %v", newTg, oldTg)
	}
	return nil
}

func (s *Scheduler) Start() {

}

// Load reads all the trackers in the repo and loads them into the scheduler
func (s *Scheduler) Load() {
	// do the description
}
