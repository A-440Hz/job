package handler

import (
	"encoding/json"
	"job/internal/tracker"
	"job/internal/user"
	"net/http"

	"gorm.io/gorm"
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

func getJobAppItemUpdateFields(r *http.Request) (*tracker.JobAppItemUpdateFields, error) {
	var fields tracker.JobAppItemUpdateFields
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&fields); err != nil {
		return nil, err
	}
	return &fields, nil
}

// ServeJobAppTrackerMainPage handles main page of the webapp
func (h *Handler) ServeJobAppTrackerMainPage(w http.ResponseWriter, r *http.Request) {
	// get user from cookie or create user
	var user *user.User
	uuid, err := getUserIdFromCookie(r)
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
			clearUserCookie(w)
			return
		}
	} else {
		user, err = h.UserService.LookupUser(uuid)
		if err != nil {
			// should I clear the cookie if the user is not found? I don't see why not
			http.Error(w, err.Error(), http.StatusInternalServerError)
			clearUserCookie(w)
			return
		}
	}

	// create new tracker if tracker not found
	tracker, err := h.TrackerService.GetJobAppTrackerWithItemsFromUserID(user.GetID())
	if err == gorm.ErrRecordNotFound {
		tracker, err = h.TrackerService.CreateNewJobAppTracker(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user":    user,
		"tracker": tracker,
	})
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

// TODO: smooth out distinction
func (h *Handler) DeleteUserRequest(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) CreateJobAppItem(w http.ResponseWriter, r *http.Request) {
	uuid, err := getUserIdFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t, err := h.TrackerService.LookupJobAppTrackerFromUserID(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uf, err := getJobAppItemUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t, err = h.TrackerService.CreateJobAppItem(uuid, uf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) UpdateJobAppItemFields(w http.ResponseWriter, r *http.Request) {
	uuid, err := getUserIdFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uf, err := getJobAppItemUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itemID := r.URL.Query().Get("id")
	t, err := h.TrackerService.UpdateJobAppItemFields(uuid, itemID, uf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) DeleteJobAppItem(w http.ResponseWriter, r *http.Request) {
	uuid, err := getUserIdFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itemID := r.URL.Query().Get("id")
	err = h.TrackerService.DeleteJobAppItem(uuid, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
