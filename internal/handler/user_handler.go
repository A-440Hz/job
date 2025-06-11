package handler

import (
	"encoding/json"
	"job/internal/user"
	"net/http"
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

func (h *Handler) ServeUserMainPage(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIdFromCookie(r, w)
	if err != nil {
		// TODO: do I create a new user here? I hope not.
		// perhaps lead back to main tracker page
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.UserService.LookupUser(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uInv, err := h.CollectionService.LookupUserInventory(uuid)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user":           user,
		"user_inventory": uInv,
	})
}

func (h *Handler) RegisterBaseUser(w http.ResponseWriter, r *http.Request) {
	// parse request info
	uuid, err := h.UserService.GetUserIdFromCookie(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	uf, err := getUserUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := h.UserService.RegisterBaseUser(uuid, uf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// do i return this? idk what react needs yet
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) HandleLoginRequest(w http.ResponseWriter, r *http.Request) {
	uf, err := getUserUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	user, err := h.UserService.LoginUser(uf)
	if err != nil {
		// the service returns safe errors
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	h.UserService.SetUserCookie(w, user.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(user)
}

func (h *Handler) UpdateUserFields(w http.ResponseWriter, r *http.Request) {
	uuid, err := h.UserService.GetUserIdFromCookie(r, w)
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
