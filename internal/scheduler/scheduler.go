package scheduler

import (
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"
)

// startingHeapCapacity scaled to the amount of users I expect
const startingHeapCapacity = 32

// if there are more than 50 trackers with the same deadline, maybe this channel will break
// one solution is to have more output channels
// other is to keep increasing the size
const outputChannelSize = 50

// Scheduler always lives in memory and manages when to trigger and reset TrackerGoals
type Scheduler struct {
	g        *GoalHeap
	mutex    sync.Mutex
	stopCh   chan bool
	nextTick chan *time.Time //
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

func (tg *TrackerGoal) resetDeadline() {
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
		g:        g,
		stopCh:   make(chan bool, outputChannelSize),
		nextTick: make(chan *time.Time), // unbuffered; remove if not
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

// peek breaks abstraction and returns item 0 in the heap. Be careful not to modify it
func (g *GoalHeap) peek() *TrackerGoal {
	if g.Len() <= 0 {
		return nil
	}
	return g.heap[0]
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

func (s *Scheduler) updateNextTick() {
	// make sure nextTick is the closest deadline
	if d := s.g.peek(); d != nil {
		<-s.nextTick
		s.nextTick <- &d.GoalDeadline
	}
}

// AddTrackerGoal adds a TrackerGoal into the scheduler
func (s *Scheduler) AddTrackerGoal(tg *TrackerGoal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	heap.Push(s.g, tg)
	s.updateNextTick()
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
	s.updateNextTick()
	return nil
}

func (s *Scheduler) Start() {
	go func() {
		select {
		case <-s.stopCh:
			// gracefully shut down
			// accept errors and close to prevent deadlocks? idk how i want to implement this
			// maybe have super errors that send me an email when the scheduler breaks
		case <-time.After(time.Until(*<-s.nextTick)): // fix this mechanism if needed
			// what happens when the first index in the heap changes?
			// I probably need a holding var in the scheduler to hold 1 TrackerGoal outside of the heap
			tg := heap.Pop(s.g).(*TrackerGoal)
			tg.resetDeadline()
			s.OutputCh <- tg
			s.AddTrackerGoal(tg)
			s.updateNextTick()
		default:
		}
	}()
}

// Load reads all the trackers in the repo and loads them into the scheduler
func (s *Scheduler) Load() {
	// do the description
}
