package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/open-workout/ow/services/workout-service/internal/domain"
)

type WorkoutHandler struct {
	svc domain.WorkoutService
}

func NewWorkoutHandler(svc domain.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{svc: svc}
}

func (h *WorkoutHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("X-User-ID")
	if userId == "" {
		http.Error(w, "missing X-User-ID header", http.StatusUnauthorized)
		return
	}

	workoutId, err := strconv.ParseInt(r.PathValue("workout_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid workout_id", http.StatusBadRequest)
		return
	}

	workout, err := h.svc.GetWorkoutById(r.Context(), workoutId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get workout", http.StatusInternalServerError)
		return
	}

	if workout.UserID != userId {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workout)
}

func (h *WorkoutHandler) GetSets(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("X-User-ID")
	if userId == "" {
		http.Error(w, "missing X-User-ID header", http.StatusUnauthorized)
		return
	}

	workoutId, err := strconv.ParseInt(r.PathValue("workout_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid workout_id", http.StatusBadRequest)
		return
	}

	sets, err := h.svc.GetSetsByWorkoutID(r.Context(), workoutId, userId)
	if err != nil {
		http.Error(w, "failed to get sets", http.StatusInternalServerError)
		return
	}

	if sets == nil {
		sets = []*domain.SetModel{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sets)
}

func (h *WorkoutHandler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	workoutId, err := strconv.ParseInt(r.PathValue("workout_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid workout_id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteWorkout(r.Context(), workoutId); err != nil {
		http.Error(w, "failed to delete workout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutHandler) UpdateSet(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("X-User-ID")
	if userId == "" {
		http.Error(w, "missing X-User-ID header", http.StatusUnauthorized)
		return
	}

	setId, err := strconv.ParseInt(r.PathValue("set_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid set_id", http.StatusBadRequest)
		return
	}

	var set domain.SetModel
	if err := json.NewDecoder(r.Body).Decode(&set); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	set.SetID = setId

	updated, err := h.svc.UpdateSet(r.Context(), userId, &set)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	userId := r.Header.Get("X-User-ID")
	if userId == "" {
		http.Error(w, "missing X-User-ID header", http.StatusUnauthorized)
		return
	}

	setId, err := strconv.ParseInt(r.PathValue("set_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid set_id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteSet(r.Context(), userId, setId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete set", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutHandler) DeleteWorkoutsByUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("user_id")
	if userId == "" {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteWorkoutsByUserID(r.Context(), userId); err != nil {
		http.Error(w, "failed to delete workouts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
