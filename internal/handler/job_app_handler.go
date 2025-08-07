package handler

import (
	"encoding/json"
	"job/internal/tracker"
	"job/internal/user"
	"net/http"

	"gorm.io/gorm"
)

func getJobAppTrackerUpdateFields(r *http.Request) (*tracker.JobAppTrackerUpdateFields, error) {
	var fields tracker.JobAppTrackerUpdateFields
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&fields); err != nil {
		return nil, err
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

func (h *Handler) ServeMainPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, POST, PUT, UPDATE, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		h.GetUserAndTrackerItems(w, r)
	case http.MethodPatch:
		h.UpdateJobAppTrackerFields(w, r)
	case http.MethodPost:
		h.CreateJobAppItem(w, r)
	case http.MethodPut:
		h.UpdateJobAppItemFields(w, r)
	case http.MethodDelete:
		h.DeleteJobAppItem(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GetUserAndTrackerItems retrieves or creates a User, and resturns it along with its tracker and any items
func (h *Handler) GetUserAndTrackerItems(w http.ResponseWriter, r *http.Request) {
	// get user from cookie or create user
	var user *user.User
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
	if err != nil {
		// check error type(?); create and store new user in cookie and db if not found
		user, err = h.UserService.CreateNewUser(PollUserTimezone(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s, err := h.UserService.CreateNewSession(user.GetID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// TODO: probably just send alert for user to manually delete the cookie?
			return
		}
		h.UserService.SetSessionCookie(w, s.ID)
	} else {
		user, err = h.UserService.GetUserAndUserInventory(uuid)
		if err != nil {
			// should I clear the cookie if the user is not found? I don't see why not
			http.Error(w, err.Error(), http.StatusInternalServerError)
			h.UserService.ClearSessionCookie(w)
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
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
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
	json.NewEncoder(w).Encode(map[string]any{
		"tracker": t,
	})
}

func (h *Handler) CreateJobAppItem(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uf, err := getJobAppItemUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t, err := h.TrackerService.CreateJobAppItem(uuid, uf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"tracker": t,
	})
}

func (h *Handler) UpdateJobAppItemFields(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
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
	json.NewEncoder(w).Encode(map[string]any{
		"tracker": t,
	})
}

func (h *Handler) DeleteJobAppItem(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
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
	json.NewEncoder(w).Encode(itemID)
}
