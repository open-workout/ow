package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/storage"
	"github.com/open-workout/ow/services/exercise-service/internal/service"
	"github.com/open-workout/ow/services/exercise-service/internal/transport/http/handlers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// server holds everything needed to make requests against a live handler.
type server struct {
	ts       *httptest.Server
	db       *sql.DB
	mediaDir string
}

func setupServer(t *testing.T) (*server, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	_, err = db.Exec(`
		CREATE TABLE exercises (
			exercise_id SERIAL PRIMARY KEY,
			name TEXT, exercise_type TEXT, primary_muscle TEXT,
			secondary_muscles TEXT[], description TEXT,
			user_id BIGINT, is_private BOOLEAN, weight_direction BIGINT
		);
		CREATE TABLE exercise_media (
			exercise_id BIGINT, url TEXT, user_id BIGINT
		);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	mediaDir := t.TempDir()
	ms, err := storage.NewLocalMediaStorage(mediaDir, "http://localhost/uploads")
	if err != nil {
		t.Fatalf("create media storage: %v", err)
	}

	repo := repository.NewSqlRepository(db)
	svc := service.NewService(repo, ms)
	h := handlers.NewExerciseHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /exercises", h.CreateExercise)
	mux.HandleFunc("POST /exercises/{id}/media", h.AddExerciseMedia)
	mux.HandleFunc("GET /exercises/{id}/media", h.GetExerciseMedia)

	ts := httptest.NewServer(mux)

	cleanup := func() {
		ts.Close()
		db.Close()
		container.Terminate(ctx)
	}
	return &server{ts: ts, db: db, mediaDir: mediaDir}, cleanup
}

// createExercise inserts an exercise via the API and returns its ID.
func createExercise(t *testing.T, s *server, userID string, isPrivate bool) int64 {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"name": "Test Exercise", "exercise_type": "compound",
		"primary_muscle": "chest", "secondary_muscles": []string{},
		"description": "", "user_id": userID,
		"is_private": isPrivate, "weight_direction": 1,
	})
	req, _ := http.NewRequest(http.MethodPost, s.ts.URL+"/exercises", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%s", userID))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}
	defer resp.Body.Close()
	var ex domain.ExerciseModel
	json.NewDecoder(resp.Body).Decode(&ex)
	return ex.ExerciseID
}

// multipartBody builds a multipart/form-data body with a single file field.
func multipartBody(t *testing.T, filename, mimeType string, content []byte) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	h.Set("Content-Type", mimeType)
	fw, err := mw.CreatePart(h)
	if err != nil {
		t.Fatalf("create form part: %v", err)
	}
	fw.Write(content)
	mw.Close()
	return &buf, mw.FormDataContentType()
}

func TestAddExerciseMedia_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	exerciseID := createExercise(t, s, "1", false)

	body, ct := multipartBody(t, "photo.jpg", "image/jpeg", []byte("fake-jpeg-data"))
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-User-ID", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// URL should be persisted in DB (the main bug we fixed)
	var url string
	err = s.db.QueryRow(`SELECT url FROM exercise_media WHERE exercise_id = $1`, exerciseID).Scan(&url)
	if err != nil {
		t.Fatalf("querying exercise_media: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL in exercise_media, got empty string")
	}

	// File should exist on disk
	entries, _ := os.ReadDir(s.mediaDir)
	if len(entries) == 0 {
		t.Error("expected file to be written to media directory")
	}
}

func TestAddExerciseMedia_UnsupportedMIME(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	exerciseID := createExercise(t, s, "1", false)

	body, ct := multipartBody(t, "malware.exe", "application/octet-stream", []byte("not an image"))
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-User-ID", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for unsupported MIME, got %d", resp.StatusCode)
	}

	// Nothing should be saved
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM exercise_media WHERE exercise_id = $1`, exerciseID).Scan(&count)
	if count != 0 {
		t.Errorf("expected no media rows, got %d", count)
	}
}

func TestAddExerciseMedia_MissingAuthHeader(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	exerciseID := createExercise(t, s, "1", false)

	body, ct := multipartBody(t, "photo.jpg", "image/jpeg", []byte("data"))
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), body)
	req.Header.Set("Content-Type", ct)
	// no X-User-ID header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing X-User-ID, got %d", resp.StatusCode)
	}
}

func TestAddExerciseMedia_Forbidden(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	// Exercise is owned by user 1
	exerciseID := createExercise(t, s, "1", false)

	// User 2 tries to upload
	body, ct := multipartBody(t, "photo.jpg", "image/jpeg", []byte("data"))
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-User-ID", "2")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner upload, got %d", resp.StatusCode)
	}

	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM exercise_media WHERE exercise_id = $1`, exerciseID).Scan(&count)
	if count != 0 {
		t.Errorf("expected no media rows after forbidden upload, got %d", count)
	}
}

func TestGetExerciseMedia_ReturnsUploadedURLs(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	exerciseID := createExercise(t, s, "1", false)

	// Upload two files
	for _, name := range []string{"a.jpg", "b.png"} {
		mime := "image/jpeg"
		if name == "b.png" {
			mime = "image/png"
		}
		body, ct := multipartBody(t, name, mime, []byte("data"))
		req, _ := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), body)
		req.Header.Set("Content-Type", ct)
		req.Header.Set("X-User-ID", "1")
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()
	}

	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), nil)
	req.Header.Set("X-User-ID", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET media failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var media []domain.ExerciseMedia
	if err := json.NewDecoder(resp.Body).Decode(&media); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(media) != 2 {
		t.Errorf("expected 2 media items, got %d", len(media))
	}
	for _, m := range media {
		if m.URL == "" {
			t.Error("expected non-empty URL in media response")
		}
	}
}

func TestGetExerciseMedia_PrivateExercise_WrongUser(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	// Private exercise owned by user 1
	exerciseID := createExercise(t, s, "1", true)

	// Upload as owner
	body, ct := multipartBody(t, "photo.jpg", "image/jpeg", []byte("data"))
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-User-ID", "1")
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	// User 2 requests media for private exercise — should get empty list (exercise not visible)
	req, _ = http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/exercises/%d/media", s.ts.URL, exerciseID), nil)
	req.Header.Set("X-User-ID", "2")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET media failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var media []domain.ExerciseMedia
	json.NewDecoder(resp.Body).Decode(&media)
	if len(media) != 0 {
		t.Errorf("expected 0 media items for private exercise seen by non-owner, got %d", len(media))
	}
}
