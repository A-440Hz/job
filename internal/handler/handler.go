package handler

import (
	"encoding/json"
	"job/internal/tracker"
	"job/internal/user"
	"net/http"
)

type Handler struct {
	UserService    *user.Service
	TrackerService *tracker.Service
}

// HandleJobAppTracker handles main page of the webapp
func (h *Handler) HandleJobAppTracker(w http.ResponseWriter, r *http.Request) {
	// get user from cookie or create user
	var user *user.User
	uuid, err := getUserCookie(w, r)
	if err != nil {
		// check error type(?); create and store new user in cookie and db if not found
		user, err = h.UserService.CreateNewUser(PollUserTimezone(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = setUserCookie(w, user.GetID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		user, err = h.UserService.LookupUser(uuid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	tracker, err := h.TrackerService.LookupJobAppTracker(user.GetID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items, err := h.TrackerService.LookupJobAppTrackerItems(tracker.GetID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user":    user,
		"tracker": tracker,
		"items":   items,
	})
}
