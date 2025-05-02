package scheduler

import (
	"time"
)

// Scheduler always lives in memory and manages when to trigger and reset TrackerGoals
type Scheduler struct {
	Heap GoalHeap
}

type TrackerGoal struct {
	TrackerID     string
	GoalDeadline  time.Time
	GoalFrequency Frequency
	index         int
}

type GoalHeap []*TrackerGoal

func (g GoalHeap) Len() int           { return len(g) }
func (g GoalHeap) Less(i, j int) bool { return g[i].GoalDeadline.Before(g[j].GoalDeadline) }
func (g GoalHeap) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
	g[i].index, g[j].index = g[j].index, g[i].index
}
