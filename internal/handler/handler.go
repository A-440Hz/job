package handler

import (
	"job/internal/collection"
	"job/internal/tracker"
	"job/internal/user"
)

// I can split this into modular handlers if I need to scale my project
type Handler struct {
	UserService       *user.Service
	TrackerService    *tracker.Service
	CollectionService *collection.Service

	// this is where I would put my middleware... if I had any!
}
