package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/open-workout/ow/services/api-gateway/internal/clients/userclient"
)

type UserHandler struct {
	client *userclient.Client
}

func NewUserHandler(client *userclient.Client) *UserHandler {
	return &UserHandler{client: client}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.client.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get user", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
