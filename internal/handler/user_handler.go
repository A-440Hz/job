package handler

import (
	"encoding/json"
	"job/internal/user"
	"net/http"

	"gorm.io/gorm"
)

// func getUserID(r *http.Request) (string, error) {
// 	var req struct {
// 		ID string `json:"id"`
// 	}
// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		return "", err
// 	}
// 	return req.ID, nil
// }

func getUserUpdateFields(r *http.Request) (*user.UserUpdateFields, error) {
	var fields user.UserUpdateFields
	err := json.NewDecoder(r.Body).Decode(&fields)
	if err != nil {
		return nil, err
	}
	return &fields, nil
}

func (h *Handler) GetUserAndUserInventory(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
	if err == http.ErrNoCookie {
		http.Error(w, "no session cookie found -- try enabling cookies, logging in, or returning to the main page", http.StatusBadRequest)
		// TODO: remember to lead back to main tracker page
		// return
	} else if err != nil {
		// TODO: remember to lead back to main tracker page
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.UserService.GetUserAndUserInventory(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// uInv, err := h.CollectionService.LookupUserInventory(uuid)
	coll, err := h.CollectionService.GetAllCollectablesForUser(uuid)

	if err != nil && err != gorm.ErrRecordNotFound {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user":         user,
		"collectables": coll,
		// "user_inventory": uInv,
	})
}

func (h *Handler) RegisterBaseUser(w http.ResponseWriter, r *http.Request) {
	// parse request info
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	uf, err := getUserUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.UserService.RegisterBaseUser(uuid, uf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// do i return this? idk what react needs yet
	// json.NewEncoder(w).Encode(u)
}

func (h *Handler) HandleLoginRequest(w http.ResponseWriter, r *http.Request) {
	uf, err := getUserUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	repoUser, err := h.UserService.LoginUser(uf)
	if err != nil {
		// the service returns safe errors
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// if userID succeeds, session expiry is automatically updated
	// if fail, old session
	clientUserID, err := h.UserService.GetUserIDFromCookie(r, w)
	if err != nil {
		sn, err := h.UserService.CreateNewSession(repoUser.GetID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		h.UserService.SetSessionCookie(w, sn.ID)
	} else if clientUserID != repoUser.GetID() {
		h.UserService.DeleteSession(h.UserService.GetSessionIDFromCookie(r, w))
		sn, err := h.UserService.CreateNewSession(repoUser.GetID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		h.UserService.SetSessionCookie(w, sn.ID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// refer to tracker main page??
	// json.NewEncoder(w).Encode(user)
}

// HandleLogoutRequest deletes the current session and clears the session cookie
func (h *Handler) HandleLogoutRequest(w http.ResponseWriter, r *http.Request) {
	sID := h.UserService.GetSessionIDFromCookie(r, w)
	h.UserService.ClearSessionCookie(w)
	if sID != "" {
		err := h.UserService.DeleteSession(sID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// refer to main page
}

func (h *Handler) UpdateUserFields(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uf, err := getUserUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := h.UserService.UpdateUserFields(uuid, uf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

// TODO: do cascade gorm delete or series of delete calls here
func (h *Handler) DeleteUserRequest(w http.ResponseWriter, r *http.Request) {
	// after delete hook
	// https://stackoverflow.com/questions/76762629/how-to-cascade-a-delete-in-gorm

	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: delete all kinds of trackers, fancily or by hand
	h.TrackerService.DeleteJobAppTrackerByUserID(uuid)
	// deletes User and UserInventory from repo
	h.UserService.DeleteUser(uuid)
	// clears session cookie
	h.UserService.ClearSessionCookie(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ServeUserMainPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetUserAndUserInventory(w, r)
	case http.MethodPatch:
		h.UpdateUserFields(w, r)
	case http.MethodDelete:
		h.DeleteUserRequest(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleAwardCollectableRequest(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIDFromCookie(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	col, err := h.CollectionService.AwardOneRandomCollectable(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(col)
}
