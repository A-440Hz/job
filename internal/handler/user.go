package handler

import (
	"encoding/json"
	"job/internal/user"
	"net/http"
)

func getUserID(r *http.Request) (string, error) {
	var req struct {
		ID string `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return "", err
	}
	return req.ID, nil
}

func getUserUpdateFields(r *http.Request) (*user.UserUpdateFields, error) {
	var fields user.UserUpdateFields
	err := json.NewDecoder(r.Body).Decode(&fields)
	if err != nil {
		return nil, err
	}
	return &fields, nil
}

func (h *Handler) RegisterBaseUser(w http.ResponseWriter, r *http.Request) {
	// parse request info
	uuid, err := getUserCookie(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	uf, err := getUserUpdateFields(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.UserService.RegisterBaseUser(uuid, uf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
