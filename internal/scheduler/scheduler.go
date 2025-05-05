package scheduler

import (
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"
)

// Scheduler always lives in memory and manages when to trigger and reset TrackerGoals
type Scheduler struct {
	goalHeap GoalHeap
	mutex    sync.Mutex
	ch       chan *GoalHeap
	stopCh   chan bool
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

// GoalHeap implements heap.Interface
type GoalHeap []*TrackerGoal

// sort.Interface methods
func (g GoalHeap) Len() int           { return len(g) }
func (g GoalHeap) Less(i, j int) bool { return g[i].GoalDeadline.Before(g[j].GoalDeadline) }
func (g GoalHeap) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
	g[i].index, g[j].index = i, j
}

// add x as element Len()
func (g *GoalHeap) Push(x any) {
	n := len(*g)
	item := x.(*TrackerGoal)
	item.index = n
	*g = append(*g, item)
}

// remove and return element Len() - 1
func (g *GoalHeap) Pop() any {
	old := *g
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*g = old[0 : n-1]
	return item
}

// PopTrackerGoal searches the heap for t.ID
func (g *GoalHeap) PopTrackerGoal(t *TrackerGoal) (*TrackerGoal, error) {
	j := g.findByID(t.TrackerID, 0, t.GoalDeadline)
	if j == -1 {
		log.Printf("attempt to find tracker unsuccessful: %v", t)
		return nil, fmt.Errorf("attempt to find tracker unsuccessful: %v", t)
	}
	ret := heap.Remove(g, j).(*TrackerGoal)
	return ret, nil
}

func (g *GoalHeap) popIdx(i int) (*TrackerGoal, error) {
	ret := heap.Remove(g, i).(*TrackerGoal)
	return ret, nil
}

// findByID returns the index or -1 if not found. it is valid but it breaks abstraction
// on the other hand the container/heap documentation describes invariants for index behavior:
// https://cs.opensource.google/go/go/+/refs/tags/go1.24.2:src/container/heap/heap.go;drc=ade730a96cdb07c60fe932373c0b05f9d15a4ec5;l=26
// I'd rather not linear search or keep a map of trackerIDs to index so I'll make a note here and keep it this way
// bottom line it relies on the heap.Interface invariant that children of i are 2*i+1 and 2*i+2 which i think is reasonably safe to assume true
func (g *GoalHeap) findByID(id string, curIdx int, deadline time.Time) int {
	if curIdx >= g.Len() || (*g)[curIdx].GoalDeadline.After(deadline) {
		return -1
	}
	if (*g)[curIdx].TrackerID == id {
		return curIdx
	}
	if l := g.findByID(id, curIdx*2+1, deadline); l != -1 {
		return l
	}
	return g.findByID(id, curIdx*2+2, deadline)
}

// Peek breaks abstraction to return item 0 in the queue
func (g *GoalHeap) Peek() *TrackerGoal {
	if g.Len() <= 0 {
		return nil
	}
	return (*g)[0]
}

func NewScheduler() *Scheduler {
	h := &GoalHeap{}
	heap.Init(h)
	return &Scheduler{
		goalHeap: *h,
		ch:       make(chan *GoalHeap),
		stopCh:   make(chan bool),
	}
}

// Load reads all the trackers in the repo and loads them into the scheduler
func (s *Scheduler) Load() {
	// do the description
}

// Push adds a TrackerGoal into the scheduler
func (s *Scheduler) Push(tg *TrackerGoal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.goalHeap.Push(tg)
	return nil
}

func (s *Scheduler) Update(beforeTg, toTg *TrackerGoal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	idx := s.goalHeap.findByID(beforeTg.TrackerID, 0, beforeTg.GoalDeadline)
	if idx == -1 {
		// we didnt find it
		log.Println("could not locate tracker %v", beforeTg)
		return fmt.Errorf("could not locate tracker %v", beforeTg)
	}

	if beforeTg.GoalDeadline == toTg.GoalDeadline {
		s.goalHeap[idx].GoalFrequency = toTg.GoalFrequency
		return nil
	}

	_, err := s.goalHeap.popIdx(idx)
	if err != nil {
		// we didn't find it
		log.Println(err)
		return err
	}
	err = s.Push(toTg)
	if err != nil {
		// maybe we should push back the previous one??
		log.Printf("failed to push tracker %v.. putting back %v", toTg, beforeTg)
		restoreErr := s.Push(beforeTg) // and then what if this one errors again
		if restoreErr != nil {
			log.Printf("failed to push back previous tracker %v", beforeTg)
		}
		return fmt.Errorf("failed to push tracker %v.. putting back %v", toTg, beforeTg)
	}
	return nil
}
