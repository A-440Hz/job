package handler

import (
	"job/internal/collection"
	"job/internal/tracker"
	"job/internal/user"
)

const localOrigin string = "http://localhost:5173"

var allowOrigin = localOrigin

// I can split this into modular handlers if I need to scale my project
type Handler struct {
	UserService       *user.Service
	TrackerService    *tracker.Service
	CollectionService *collection.Service

	// this is where I would put my middleware... if I had any!
}

func SetEnvForProduction() {
	allowOrigin = "*"
}
