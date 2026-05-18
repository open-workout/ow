package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/open-workout/ow/services/api-gateway/internal/clients/userclient"
	appmw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

type UserHandler struct {
	client *userclient.Client
}

func NewUserHandler(client *userclient.Client) *UserHandler {
	return &UserHandler{client: client}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	callerID := appmw.GetUserID(r)

	var user userclient.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.client.CreateUser(r.Context(), callerID, user)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.client.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, userclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get user", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	callerID := appmw.GetUserID(r)

	var user userclient.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.client.UpdateUser(r.Context(), callerID, id, user)
	if err != nil {
		if errors.Is(err, userclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update user", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	callerID := appmw.GetUserID(r)

	if err := h.client.DeleteUser(r.Context(), callerID, id); err != nil {
		if errors.Is(err, userclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete user", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdateSplit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	callerID := appmw.GetUserID(r)

	var split userclient.Split
	if err := json.NewDecoder(r.Body).Decode(&split); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.client.UpdateSplit(r.Context(), callerID, id, split)
	if err != nil {
		if errors.Is(err, userclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update split", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
