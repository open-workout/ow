package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/open-workout/ow/services/user-service/internal/domain"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user := &domain.User{UserId: id}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
