package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/open-workout/ow/services/api-gateway/internal/clients/exerciseclient"
	appmw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

type ExerciseHandler struct {
	client *exerciseclient.Client
}

func NewExerciseHandler(client *exerciseclient.Client) *ExerciseHandler {
	return &ExerciseHandler{client: client}
}

func (h *ExerciseHandler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var exercise exerciseclient.Exercise
	if err := json.NewDecoder(r.Body).Decode(&exercise); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.client.CreateExercise(r.Context(), exercise)
	if err != nil {
		http.Error(w, "failed to create exercise", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	userID := appmw.GetUserID(r)

	exercises, err := h.client.ListExercises(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list exercises", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercises)
}

func (h *ExerciseHandler) GetTopExercises(w http.ResponseWriter, r *http.Request) {
	var req exerciseclient.TopExercisesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	exercises, err := h.client.GetTopExercises(r.Context(), req)
	if err != nil {
		http.Error(w, "failed to get top exercises", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercises)
}

func (h *ExerciseHandler) GetExerciseById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	userID := appmw.GetUserID(r)

	ex, err := h.client.GetExerciseById(r.Context(), id, userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ex)
}

func (h *ExerciseHandler) AddExerciseMedia(w http.ResponseWriter, r *http.Request) {
	exerciseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	userID := appmw.GetUserID(r)

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

	if err := h.client.AddExerciseMedia(r.Context(), exerciseID, userID,
		header.Filename, header.Header.Get("Content-Type"), io.Reader(file)); err != nil {
		http.Error(w, "failed to add media", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExerciseHandler) GetExerciseMedia(w http.ResponseWriter, r *http.Request) {
	exerciseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid exercise id", http.StatusBadRequest)
		return
	}

	userID := appmw.GetUserID(r)

	media, err := h.client.GetExerciseMedia(r.Context(), exerciseID, userID)
	if err != nil {
		http.Error(w, "failed to get media", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}
