package handler

import (
	"encoding/json"
	"job/internal/tracker"
	"net/http"
)

func getJobAppTrackerUpdateFields(r *http.Request) (*tracker.JobAppTrackerUpdateFields, error) {
	var underlyingFields tracker.UnderlyingTrackerUpdateFields
	var fields tracker.JobAppTrackerUpdateFields
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&underlyingFields); err != nil {
		return nil, err
	}
	if err := d.Decode(&fields); err != nil {
		return nil, err
	}
	if !underlyingFields.IsNil() {
		fields.UnderlyingTrackerUpdateFields = underlyingFields
	}
	return &fields, nil
}

func (h *Handler) UpdateJobAppTrackerFields(w http.ResponseWriter, r *http.Request) {
	uuid, err := getUserIdFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uf, err := getJobAppTrackerUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t, err := h.TrackerService.UpdateJobAppTrackerFields(uuid, uf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)
}
