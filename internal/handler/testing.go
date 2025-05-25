package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) SelectEverything(w http.ResponseWriter, r *http.Request) {
	trackers, err := h.TrackerService.SelectAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	users, err := h.UserService.SelectAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"users":    users,
		"trackers": trackers,
	})
}
