package handler

import (
	"encoding/json"
	"net/http"
)

// TODO: gate or block this function in production
func (h *Handler) SelectEverything(w http.ResponseWriter, r *http.Request) {
	trackers, err := h.TrackerService.SelectAllJobAppTrackers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	users, err := h.UserService.SelectAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	collectables, err := h.CollectionService.SelectAllCollectables()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	sessions, err := h.UserService.SelectAllSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"users":        users,
		"trackers":     trackers,
		"collectables": collectables,
		"sessions":     sessions,
	})
}
