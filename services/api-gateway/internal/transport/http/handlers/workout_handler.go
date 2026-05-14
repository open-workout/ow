package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/open-workout/ow/services/api-gateway/internal/clients/workoutclient"
	appmw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

type WorkoutHandler struct {
	client *workoutclient.Client
}

func NewWorkoutHandler(client *workoutclient.Client) *WorkoutHandler {
	return &WorkoutHandler{client: client}
}

// effectiveUserID returns 0 for admin callers (workout service treats 0 as bypass),
// or the caller's own user ID for regular users.
func effectiveUserID(r *http.Request) (int64, error) {
	if appmw.GetUserRole(r) == "admin" {
		return 0, nil
	}
	return strconv.ParseInt(appmw.GetUserID(r), 10, 64)
}

func (h *WorkoutHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	userID, err := effectiveUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	workoutID, err := strconv.ParseInt(chi.URLParam(r, "workout_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid workout_id", http.StatusBadRequest)
		return
	}

	workout, err := h.client.GetWorkoutById(r.Context(), userID, workoutID)
	if err != nil {
		if errors.Is(err, workoutclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get workout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workout)
}

func (h *WorkoutHandler) GetSets(w http.ResponseWriter, r *http.Request) {
	userID, err := effectiveUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	workoutID, err := strconv.ParseInt(chi.URLParam(r, "workout_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid workout_id", http.StatusBadRequest)
		return
	}

	sets, err := h.client.GetSetsByWorkoutID(r.Context(), userID, workoutID)
	if err != nil {
		if errors.Is(err, workoutclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get sets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sets)
}

func (h *WorkoutHandler) UpdateSet(w http.ResponseWriter, r *http.Request) {
	userID, err := effectiveUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	setID, err := strconv.ParseInt(chi.URLParam(r, "set_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid set_id", http.StatusBadRequest)
		return
	}

	var set workoutclient.SetModel
	if err := json.NewDecoder(r.Body).Decode(&set); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	set.SetID = setID

	updated, err := h.client.UpdateSet(r.Context(), userID, set)
	if err != nil {
		if errors.Is(err, workoutclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update set", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *WorkoutHandler) DeleteSet(w http.ResponseWriter, r *http.Request) {
	userID, err := effectiveUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	setID, err := strconv.ParseInt(chi.URLParam(r, "set_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid set_id", http.StatusBadRequest)
		return
	}

	if err := h.client.DeleteSet(r.Context(), userID, setID); err != nil {
		if errors.Is(err, workoutclient.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete set", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
