package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
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

func (h *ExerciseHandler) GetExerciseById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	callerUserID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil {
		http.Error(w, "missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	ex, err := h.svc.GetExerciseById(r.Context(), id, callerUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get exercise", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ex)
}

func (h *ExerciseHandler) UpdateExercise(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	callerUserID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil {
		http.Error(w, "missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	var exercise domain.ExerciseModel
	if err := json.NewDecoder(r.Body).Decode(&exercise); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	exercise.ExerciseID = id

	updated, err := h.svc.UpdateExercise(r.Context(), callerUserID, &exercise)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update exercise", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *ExerciseHandler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	callerUserID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil {
		http.Error(w, "missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	if err := h.svc.DeleteExercise(r.Context(), callerUserID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete exercise", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var allowedMIME = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"video/mp4":       true,
	"video/quicktime": true,
	"video/webm":      true,
}

func (h *ExerciseHandler) AddExerciseMedia(w http.ResponseWriter, r *http.Request) {
	exerciseID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	callerUserID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil {
		http.Error(w, "missing or invalid X-User-ID header", http.StatusUnauthorized)
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

	mimeType := header.Header.Get("Content-Type")
	if !allowedMIME[mimeType] {
		http.Error(w, "unsupported file type", http.StatusBadRequest)
		return
	}

	media := &domain.ExerciseMedia{ExerciseID: exerciseID, UserID: callerUserID}
	upload := &domain.ExerciseMediaUpload{
		ExerciseID: exerciseID,
		UserID:     callerUserID,
		File:       file,
		Filename:   header.Filename,
		MimeType:   mimeType,
	}

	if err := h.svc.AddExerciseMedia(r.Context(), exerciseID, callerUserID, media, upload); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to add media", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExerciseHandler) GetExerciseMedia(w http.ResponseWriter, r *http.Request) {
	exerciseID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	callerUserID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil {
		http.Error(w, "missing or invalid X-User-ID header", http.StatusUnauthorized)
		return
	}

	media, err := h.svc.GetExerciseMedia(r.Context(), exerciseID, callerUserID)
	if err != nil {
		http.Error(w, "failed to get media", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}
