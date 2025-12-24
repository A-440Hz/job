package handler

import (
	"fmt"
	"job/internal/collection"
	"job/internal/tracker"
	"job/internal/user"
	"log"
	"os"
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

func SetEnvForProduction() error {
	if o, exists := os.LookupEnv("ALLOWED_ORIGINS"); exists {
		allowOrigin = o
		log.Println("Allowed origin is: ", allowOrigin)
		// Configure cookie domain based on allowed origin
		user.SetCookieDomain(allowOrigin)
	} else {
		return fmt.Errorf("ALLOWED_ORIGINS not set in production environment")
	}
	return nil
}
