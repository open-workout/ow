package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/open-workout/ow/services/user-service/internal/domain"
	"github.com/open-workout/ow/services/user-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/user-service/internal/service"
	"github.com/open-workout/ow/services/user-service/internal/transport/http/handlers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type server struct {
	ts *httptest.Server
	db *sql.DB
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
		CREATE TABLE users (
			user_id     TEXT PRIMARY KEY,
			email       TEXT NOT NULL UNIQUE,
			username    TEXT NOT NULL UNIQUE,
			sport_goals TEXT[] NOT NULL DEFAULT '{}',
			gender      TEXT,
			birthdate   TEXT,
			split       JSONB NOT NULL DEFAULT '{}'
		);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	repo := repository.NewSqlRepository(db)
	svc := service.NewService(repo)
	h := handlers.NewUserHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users/{id}", h.GetUser)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)
	mux.HandleFunc("PUT /users/{id}/split", h.UpdateSplit)

	ts := httptest.NewServer(mux)

	return &server{ts: ts, db: db}, func() {
		ts.Close()
		db.Close()
		container.Terminate(ctx)
	}
}

// createUser creates a user via POST /users with the X-User-ID header.
func createUser(t *testing.T, s *server, userID, email string) domain.User {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"email": email, "username": email, "gender": "male", "birthdate": "1990-01-01"})
	req, _ := http.NewRequest(http.MethodPost, s.ts.URL+"/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var u domain.User
	json.NewDecoder(resp.Body).Decode(&u)
	return u
}

func userURL(s *server, id string) string {
	return fmt.Sprintf("%s/users/%s", s.ts.URL, url.PathEscape(id))
}

// --- CreateUser ---

func TestCreateUser_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]any{
		"email":       "new@example.com",
		"username":    "newuser",
		"sport_goals": []string{"strength"},
		"gender":      "female",
		"birthdate":   "1995-03-20",
	})
	req, _ := http.NewRequest(http.MethodPost, s.ts.URL+"/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "auth0|new-1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var u domain.User
	json.NewDecoder(resp.Body).Decode(&u)

	if u.UserId != "auth0|new-1" {
		t.Errorf("expected user_id auth0|new-1, got %s", u.UserId)
	}
	if u.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", u.Email)
	}

	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = 'new@example.com'`).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 row in db, got %d", count)
	}
}

func TestCreateUser_MissingUserID(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]any{"email": "noid@example.com", "username": "noid"})
	resp, err := http.Post(s.ts.URL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 (missing X-User-ID), got %d", resp.StatusCode)
	}
}

func TestCreateUser_InvalidBody(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodPost, s.ts.URL+"/users", bytes.NewReader([]byte("not json")))
	req.Header.Set("X-User-ID", "auth0|x")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateUser_MissingUsername(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]any{"email": "nousername@example.com"})
	req, _ := http.NewRequest(http.MethodPost, s.ts.URL+"/users", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "auth0|x")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- GetUser ---

func TestGetUser_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|get-1", "get@example.com")

	resp, err := http.Get(userURL(s, created.UserId))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var u domain.User
	json.NewDecoder(resp.Body).Decode(&u)
	if u.UserId != created.UserId {
		t.Errorf("expected user_id %s, got %s", created.UserId, u.UserId)
	}
	if u.Email != "get@example.com" {
		t.Errorf("expected email get@example.com, got %s", u.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	resp, err := http.Get(userURL(s, "auth0|nobody"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --- UpdateUser ---

func TestUpdateUser_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|upd-1", "upd@example.com")

	body, _ := json.Marshal(map[string]any{
		"email":       "upd2@example.com",
		"sport_goals": []string{"cardio"},
		"gender":      "female",
		"birthdate":   "1992-07-04",
	})
	req, _ := http.NewRequest(http.MethodPut, userURL(s, created.UserId), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", created.UserId)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var u domain.User
	json.NewDecoder(resp.Body).Decode(&u)
	if u.Email != "upd2@example.com" {
		t.Errorf("expected email upd2@example.com, got %s", u.Email)
	}
}

func TestUpdateUser_Forbidden(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|forbidden-1", "forbidden@example.com")

	body, _ := json.Marshal(map[string]any{"email": "hacked@example.com"})
	req, _ := http.NewRequest(http.MethodPut, userURL(s, created.UserId), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "auth0|other-user")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]any{"email": "ghost@example.com"})
	req, _ := http.NewRequest(http.MethodPut, userURL(s, "auth0|ghost"), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "auth0|ghost")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --- DeleteUser ---

func TestDeleteUser_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|del-1", "del@example.com")

	req, _ := http.NewRequest(http.MethodDelete, userURL(s, created.UserId), nil)
	req.Header.Set("X-User-ID", created.UserId)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE user_id = $1`, created.UserId).Scan(&count)
	if count != 0 {
		t.Errorf("expected user to be deleted, got %d rows", count)
	}
}

func TestDeleteUser_Forbidden(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|nodel-1", "nodelete@example.com")

	req, _ := http.NewRequest(http.MethodDelete, userURL(s, created.UserId), nil)
	req.Header.Set("X-User-ID", "auth0|other")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodDelete, userURL(s, "auth0|nobody"), nil)
	req.Header.Set("X-User-ID", "auth0|nobody")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --- UpdateSplit ---

func TestUpdateSplit_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|split-1", "split@example.com")

	split := domain.Split{
		Elements: []domain.SplitElement{
			{Title: "Push", Muscles: []string{"chest", "triceps"}},
			{Title: "Pull", Muscles: []string{"back", "biceps"}},
			{Title: "Legs", Muscles: []string{"quads", "hamstrings"}},
		},
	}
	body, _ := json.Marshal(split)

	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%s/split", s.ts.URL, url.PathEscape(created.UserId)),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", created.UserId)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var u domain.User
	json.NewDecoder(resp.Body).Decode(&u)

	if len(u.ExerciseSplit.Elements) != 3 {
		t.Errorf("expected 3 split elements, got %d", len(u.ExerciseSplit.Elements))
	}
	if u.ExerciseSplit.Elements[0].Title != "Push" {
		t.Errorf("expected first element Push, got %s", u.ExerciseSplit.Elements[0].Title)
	}
}

func TestUpdateSplit_Forbidden(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|nosplit-1", "nosplit@example.com")

	body, _ := json.Marshal(domain.Split{})
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%s/split", s.ts.URL, url.PathEscape(created.UserId)),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "auth0|other")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestUpdateSplit_NotFound(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(domain.Split{})
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%s/split", s.ts.URL, url.PathEscape("auth0|nobody")),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "auth0|nobody")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateSplit_InvalidBody(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "auth0|badsplit-1", "badsplit@example.com")

	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%s/split", s.ts.URL, url.PathEscape(created.UserId)),
		bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", created.UserId)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
