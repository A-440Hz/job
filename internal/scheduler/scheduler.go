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
	OutputCh chan *TrackerGoal
	timer    *time.Timer
}

// GoalHeap implements heap.Interface and sort.Interface
type GoalHeap struct {
	heap   []*TrackerGoal
	idxMap map[string]int
}

type TrackerGoal struct {
	TrackerID      string
	CycleDeadline  time.Time
	CycleFrequency Frequency
	index          int
	TrackerType    string
}

func (tg *TrackerGoal) resetDeadline(withRetry bool) {
	now := time.Now()
	// a loop should be fine as long as front end prevents setting a deadline a million years back
	// for tg.CycleDeadline.Before(now) {
	// 	tg.CycleDeadline = tg.CycleDeadline.Add(time.Duration(tg.CycleFrequency.NumDays()))
	// }
	if now.Before(tg.CycleDeadline) {
		log.Printf("reset deadline passthrough on %v", tg.CycleDeadline)
		return
	}
	// // do not use time.Add in case of large deficits to try to avoid looping edge case.
	// // hard reset to current year and adjust fine deadline next loop in the scheduler.
	// if tg.CycleDeadline.Year()+1 < now.Year() {
	// 	log.Printf("tg year: %v, now year: %v", tg.CycleDeadline.Year(), now.Year())
	// 	log.Printf("deadline year deficit too large (%v vs %v): first resetting %v to current year", tg.CycleDeadline.Year(), now.Year(), tg.CycleDeadline)
	// 	tg.CycleDeadline = time.Date(now.Year(), tg.CycleDeadline.Month(), tg.CycleDeadline.Day(),
	// 		tg.CycleDeadline.Hour(), tg.CycleDeadline.Minute(), tg.CycleDeadline.Second(), tg.CycleDeadline.Nanosecond(), tg.CycleDeadline.Location())

	// 	// try returning this function again to avoid edge case where user is penalized 2 lootboxes instead of 1,
	// 	// from having scheduler report 2 deadline resets
	// 	if withRetry {
	// 		tg.resetDeadline(false)
	// 		return
	// 	}
	// 	log.Printf("the scheduler was unsuccessful in adjusting the deadline year to the current year: %v", tg)
	// 	return
	// }
	// TODO: refactor this numDays part out and prevent any newTg from having a nil CycleDeadline value
	prev := tg.CycleDeadline
	numDaysBetween := int64(now.Sub(tg.CycleDeadline).Round(time.Hour) / (time.Hour * 24))
	if numDaysBetween > 100000 {
		// ???? how does it reset to 0 and then bounce back and retain the
	}
	log.Printf("numDaysBetween: %d, now: %v, prev: %v", numDaysBetween, now, prev)
	log.Printf("tg.CycleFrequency.NumDays(): %d", tg.CycleFrequency.NumDays())
	numDaysToNext := (int(numDaysBetween/int64(tg.CycleFrequency.NumDays())) + 1) * tg.CycleFrequency.NumDays()
	tg.CycleDeadline = tg.CycleDeadline.Add(time.Hour * 24 * time.Duration(numDaysToNext))
	log.Printf("reset %v to %v", prev, tg.CycleDeadline)
}

func NewScheduler() *Scheduler {
	g := &GoalHeap{
		heap:   make([]*TrackerGoal, 0, startingHeapCapacity),
		idxMap: make(map[string]int),
	}
	heap.Init(g)
	s := &Scheduler{
		g:        g,
		OutputCh: make(chan *TrackerGoal, outputChannelSize),
		stopCh:   make(chan bool, 1),
		timer:    time.NewTimer(time.Hour * 24 * 365), // set it 1 year in the future
	}
	return s
}

func (g GoalHeap) Len() int           { return len(g.heap) }
func (g GoalHeap) Less(i, j int) bool { return g.heap[i].CycleDeadline.Before(g.heap[j].CycleDeadline) }
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

func (s *Scheduler) updateTimer() {
	log.Printf("updateTimer triggered. heap top is %v", s.g.peek())
	if d := s.g.peek(); d != nil {
		log.Print("reset timer to ", d.CycleDeadline)
		s.timer.Reset(time.Until(d.CycleDeadline))
	} else {
		log.Print("updateTimer -> nil; heap is nil")
	}

}

// AddTrackerGoal adds a TrackerGoal into the scheduler
func (s *Scheduler) AddTrackerGoal(tg *TrackerGoal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	heap.Push(s.g, tg)
	s.updateTimer()
	return nil
}

// ReplaceTrackerGoal swaps oldTg with newTg in the scheduler. If newTg is nil, it removes oldTg.
func (s *Scheduler) ReplaceTrackerGoal(oldTg, newTg *TrackerGoal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	idx, err := s.g.findByID(oldTg.TrackerID)
	if err != nil {
		log.Print("Could not find old TrackerGoal in heap, pushing new one to finish replace process")
		heap.Push(s.g, newTg)
		s.updateTimer()
		return nil
	}
	// remove oldTg if replacing with nil
	if newTg == nil {
		heap.Remove(s.g, idx)
		return nil
	}
	// update in place if we only need to change CycleFrequency
	if oldTg.CycleDeadline.Equal(newTg.CycleDeadline) {
		s.g.heap[idx].CycleFrequency = newTg.CycleFrequency
		return nil
	}
	// otherwise pop and replace old TrackerGoal
	heap.Remove(s.g, idx)
	heap.Push(s.g, newTg)
	s.updateTimer()
	return nil
}

// Start needs to be called as a goroutine
func (s *Scheduler) Start() {
	for {
		select {
		case <-s.stopCh:
			// gracefully shut down
			// accept errors and close to prevent deadlocks? idk how i want to implement this
			// maybe have super errors that send me an email when the scheduler breaks
			// close(s.nextTick)
			s.timer.Stop()
			close(s.stopCh)
			close(s.OutputCh)
			log.Print("Scheduler stopped")
			// TODO: this loop doesnt print
			for _, tg := range s.g.heap {
				log.Print("heap: ", tg.CycleDeadline, tg.CycleFrequency)
			}
			return
		case t := <-s.timer.C:
			s.mutex.Lock()
			log.Printf("timer tick at %v", t)
			tg := heap.Pop(s.g).(*TrackerGoal)
			log.Printf("popped %v", tg.CycleDeadline)
			tg.resetDeadline(true)
			s.OutputCh <- tg
			heap.Push(s.g, tg)
			s.updateTimer()
			s.mutex.Unlock()
		}
	}
}

func (s *Scheduler) Stop() {
	s.stopCh <- true
}
