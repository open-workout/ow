package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
)

type ExerciseHandler struct {
	svc domain.ExerciseService
}

func NewExerciseHandler(svc domain.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{svc: svc}
}

func (h *ExerciseHandler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var exercise domain.ExerciseModel
	if err := json.NewDecoder(r.Body).Decode(&exercise); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.svc.CreateExercise(r.Context(), &exercise)
	if err != nil {
		http.Error(w, "failed to create exercise", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	exercises, err := h.svc.ListExercises(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list exercises", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercises)
}

type getTopRequest struct {
	Muscles map[string]float64 `json:"muscles"`
	UserID  int64              `json:"user_id"`
	Limit   int                `json:"limit"`
}

func (h *ExerciseHandler) GetTopExercises(w http.ResponseWriter, r *http.Request) {
	var req getTopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	state := domain.MuscleState{Muscles: req.Muscles, UserID: req.UserID}
	exercises, err := h.svc.GetTopExercises(r.Context(), state, req.Limit)
	if err != nil {
		http.Error(w, "failed to get top exercises", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercises)
}

func (h *ExerciseHandler) AddExerciseMedia(w http.ResponseWriter, r *http.Request) {
	exerciseID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	media := &domain.ExerciseMedia{ExerciseID: exerciseID, UserID: userID}
	upload := &domain.ExerciseMediaUpload{
		ExerciseID: exerciseID,
		UserID:     userID,
		File:       file,
		Filename:   header.Filename,
		MimeType:   header.Header.Get("Content-Type"),
	}

	if err := h.svc.AddExerciseMedia(r.Context(), exerciseID, media, upload); err != nil {
		http.Error(w, "failed to add media", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
