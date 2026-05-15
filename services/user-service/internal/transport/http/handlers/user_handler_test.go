package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
			user_id       BIGSERIAL PRIMARY KEY,
			email         TEXT NOT NULL UNIQUE,
			username      TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL DEFAULT '',
			sport_goals   TEXT[] NOT NULL DEFAULT '{}',
			gender        TEXT,
			birthdate     TEXT,
			split         JSONB NOT NULL DEFAULT '{}'
		);
		CREATE TABLE refresh_tokens (
			id         BIGSERIAL PRIMARY KEY,
			user_id    BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
	mux.HandleFunc("POST /internal/auth/login", h.Login)
	mux.HandleFunc("POST /internal/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /internal/auth/logout", h.Logout)

	ts := httptest.NewServer(mux)

	return &server{ts: ts, db: db}, func() {
		ts.Close()
		db.Close()
		container.Terminate(ctx)
	}
}

func createUser(t *testing.T, s *server, email string) domain.User {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"email": email, "username": email, "password": "test-password", "gender": "male", "birthdate": "1990-01-01"})
	resp, err := http.Post(s.ts.URL+"/users", "application/json", bytes.NewReader(body))
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

// --- CreateUser ---

func TestCreateUser_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]any{
		"email":       "new@example.com",
		"username":    "newuser",
		"password":    "securepassword",
		"sport_goals": []string{"strength"},
		"gender":      "female",
		"birthdate":   "1995-03-20",
	})
	resp, err := http.Post(s.ts.URL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var u domain.User
	json.NewDecoder(resp.Body).Decode(&u)

	if u.UserId == 0 {
		t.Error("expected non-zero user_id")
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

func TestCreateUser_InvalidBody(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	resp, err := http.Post(s.ts.URL+"/users", "application/json", bytes.NewReader([]byte("not json")))
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

	created := createUser(t, s, "get@example.com")

	resp, err := http.Get(fmt.Sprintf("%s/users/%d", s.ts.URL, created.UserId))
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
		t.Errorf("expected user_id %d, got %d", created.UserId, u.UserId)
	}
	if u.Email != "get@example.com" {
		t.Errorf("expected email get@example.com, got %s", u.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	resp, err := http.Get(s.ts.URL + "/users/999")
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

	created := createUser(t, s, "upd@example.com")

	body, _ := json.Marshal(map[string]any{
		"email":       "upd2@example.com",
		"sport_goals": []string{"cardio"},
		"gender":      "female",
		"birthdate":   "1992-07-04",
	})
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d", s.ts.URL, created.UserId),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", created.UserId))

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

func TestUpdateUser_AdminBypass(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "admin-upd@example.com")

	body, _ := json.Marshal(map[string]any{"email": "admin-updated@example.com"})
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d", s.ts.URL, created.UserId),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for admin update, got %d", resp.StatusCode)
	}
}

func TestUpdateUser_Forbidden(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "forbidden@example.com")

	body, _ := json.Marshal(map[string]any{"email": "hacked@example.com"})
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d", s.ts.URL, created.UserId),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", created.UserId+100))

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
	req, _ := http.NewRequest(http.MethodPut, s.ts.URL+"/users/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "999")

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

	created := createUser(t, s, "del@example.com")

	req, _ := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/users/%d", s.ts.URL, created.UserId), nil)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", created.UserId))

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

func TestDeleteUser_AdminBypass(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "admin-del@example.com")

	req, _ := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/users/%d", s.ts.URL, created.UserId), nil)
	req.Header.Set("X-User-ID", "0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 for admin delete, got %d", resp.StatusCode)
	}
}

func TestDeleteUser_Forbidden(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "nodelete@example.com")

	req, _ := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/users/%d", s.ts.URL, created.UserId), nil)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", created.UserId+100))

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

	req, _ := http.NewRequest(http.MethodDelete, s.ts.URL+"/users/999", nil)
	req.Header.Set("X-User-ID", "999")

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

	created := createUser(t, s, "split@example.com")

	split := domain.Split{
		Elements: []domain.SplitElement{
			{Title: "Push", Muscles: []string{"chest", "triceps"}},
			{Title: "Pull", Muscles: []string{"back", "biceps"}},
			{Title: "Legs", Muscles: []string{"quads", "hamstrings"}},
		},
	}
	body, _ := json.Marshal(split)

	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d/split", s.ts.URL, created.UserId),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", created.UserId))

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

func TestUpdateSplit_AdminBypass(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "adminsplit@example.com")

	body, _ := json.Marshal(domain.Split{Elements: []domain.SplitElement{{Title: "Full Body", Muscles: []string{"all"}}}})
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d/split", s.ts.URL, created.UserId),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for admin split update, got %d", resp.StatusCode)
	}
}

func TestUpdateSplit_Forbidden(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	created := createUser(t, s, "nosplit@example.com")

	body, _ := json.Marshal(domain.Split{})
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d/split", s.ts.URL, created.UserId),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", created.UserId+100))

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
	req, _ := http.NewRequest(http.MethodPut, s.ts.URL+"/users/999/split", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "999")

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

	created := createUser(t, s, "badsplit@example.com")

	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d/split", s.ts.URL, created.UserId),
		bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", created.UserId))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- CreateUser validation ---

func TestCreateUser_MissingUsername(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]any{"email": "nousername@example.com", "password": "pw"})
	resp, err := http.Post(s.ts.URL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateUser_MissingPassword(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]any{"email": "nopw@example.com", "username": "nopw"})
	resp, err := http.Post(s.ts.URL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// helper shared by Login/Refresh/Logout tests

func loginUser(t *testing.T, s *server, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post(s.ts.URL+"/internal/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	var r struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	return r.RefreshToken
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	createUser(t, s, "loginok@example.com")

	body, _ := json.Marshal(map[string]string{"username": "loginok@example.com", "password": "test-password"})
	resp, err := http.Post(s.ts.URL+"/internal/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result struct {
		UserID       int64  `json:"user_id"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.UserID == 0 {
		t.Error("expected non-zero user_id")
	}
	if len(result.RefreshToken) != 64 {
		t.Errorf("expected 64-char refresh token, got %d chars", len(result.RefreshToken))
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()
	createUser(t, s, "wrongpw@example.com")

	body, _ := json.Marshal(map[string]string{"username": "wrongpw@example.com", "password": "bad-password"})
	resp, err := http.Post(s.ts.URL+"/internal/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_UnknownUsername(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]string{"username": "ghost", "password": "pw"})
	resp, err := http.Post(s.ts.URL+"/internal/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	resp, err := http.Post(s.ts.URL+"/internal/auth/login", "application/json", bytes.NewReader([]byte("not json")))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- Refresh ---

func TestRefresh_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	createUser(t, s, "refresh@example.com")
	token := loginUser(t, s, "refresh@example.com", "test-password")

	body, _ := json.Marshal(map[string]string{"refresh_token": token})
	resp, err := http.Post(s.ts.URL+"/internal/auth/refresh", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result struct {
		UserID int64 `json:"user_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.UserID == 0 {
		t.Error("expected non-zero user_id")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]string{"refresh_token": "fake-token"})
	resp, err := http.Post(s.ts.URL+"/internal/auth/refresh", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRefresh_InvalidBody(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	resp, err := http.Post(s.ts.URL+"/internal/auth/refresh", "application/json", bytes.NewReader([]byte("not json")))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	createUser(t, s, "logout@example.com")
	token := loginUser(t, s, "logout@example.com", "test-password")

	// Logout
	body, _ := json.Marshal(map[string]string{"refresh_token": token})
	req, _ := http.NewRequest(http.MethodPost, s.ts.URL+"/internal/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// Refresh should now fail
	body, _ = json.Marshal(map[string]string{"refresh_token": token})
	resp2, err := http.Post(s.ts.URL+"/internal/auth/refresh", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("refresh after logout failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 after logout, got %d", resp2.StatusCode)
	}
}

func TestLogout_InvalidBody(t *testing.T) {
	s, cleanup := setupServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodPost, s.ts.URL+"/internal/auth/logout", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
