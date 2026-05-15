package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/open-workout/ow/services/user-service/internal/domain"
	"github.com/open-workout/ow/services/user-service/internal/infrastructure/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(t *testing.T) (*sql.DB, func()) {
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

	return db, func() {
		db.Close()
		container.Terminate(ctx)
	}
}

func setupSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
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
}

func TestSqlRepository_CreateUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	u := &domain.User{
		Email:      "alice@example.com",
		Username:   "alice",
		SportGoals: []string{"strength", "endurance"},
		Gender:     "female",
		Birthdate:  "1990-04-15",
	}

	created, err := repo.CreateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.UserId == 0 {
		t.Error("expected non-zero user_id")
	}
	if created.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", created.Email)
	}
	if len(created.SportGoals) != 2 {
		t.Errorf("expected 2 sport goals, got %d", len(created.SportGoals))
	}

	var email string
	db.QueryRow(`SELECT email FROM users WHERE user_id = $1`, created.UserId).Scan(&email)
	if email != "alice@example.com" {
		t.Errorf("DB has wrong email: %s", email)
	}
}

func TestSqlRepository_CreateUser_NilSportGoals(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	u := &domain.User{Email: "bob@example.com", Username: "bob"}

	created, err := repo.CreateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("unexpected error creating user with nil sport_goals: %v", err)
	}
	if created.SportGoals == nil {
		t.Error("expected non-nil sport_goals slice")
	}
}

func TestSqlRepository_CreateUser_DuplicateEmail(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	u := &domain.User{Email: "dup@example.com", Username: "dup"}
	_, err := repo.CreateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = repo.CreateUser(context.Background(), u)
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
}

func TestSqlRepository_GetUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateUser(context.Background(), &domain.User{
		Email:     "get@example.com",
		Username:  "getuser",
		Gender:    "male",
		Birthdate: "1985-01-01",
	})

	got, err := repo.GetUser(context.Background(), created.UserId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserId != created.UserId {
		t.Errorf("expected ID %d, got %d", created.UserId, got.UserId)
	}
	if got.Email != "get@example.com" {
		t.Errorf("expected email get@example.com, got %s", got.Email)
	}
}

func TestSqlRepository_GetUser_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.GetUser(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestSqlRepository_UpdateUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateUser(context.Background(), &domain.User{
		Email:     "update@example.com",
		Username:  "updateuser",
		Gender:    "male",
		Birthdate: "2000-06-01",
	})

	created.Email = "updated@example.com"
	created.SportGoals = []string{"flexibility"}
	created.Gender = "non-binary"

	updated, err := repo.UpdateUser(context.Background(), created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Email != "updated@example.com" {
		t.Errorf("expected email updated@example.com, got %s", updated.Email)
	}
	if updated.Gender != "non-binary" {
		t.Errorf("expected gender non-binary, got %s", updated.Gender)
	}
	if len(updated.SportGoals) != 1 || updated.SportGoals[0] != "flexibility" {
		t.Errorf("unexpected sport_goals: %v", updated.SportGoals)
	}
}

func TestSqlRepository_UpdateUser_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.UpdateUser(context.Background(), &domain.User{UserId: 999, Email: "ghost@example.com", Username: "ghost"})
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestSqlRepository_DeleteUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateUser(context.Background(), &domain.User{Email: "del@example.com", Username: "del"})

	if err := repo.DeleteUser(context.Background(), created.UserId); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM users WHERE user_id = $1`, created.UserId).Scan(&count)
	if count != 0 {
		t.Errorf("expected user to be deleted, got %d rows", count)
	}
}

func TestSqlRepository_DeleteUser_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	err := repo.DeleteUser(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestSqlRepository_UpdateSplit(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateUser(context.Background(), &domain.User{Email: "split@example.com", Username: "splituser"})

	split := domain.Split{
		Elements: []domain.SplitElement{
			{Title: "Push", Muscles: []string{"chest", "triceps"}},
			{Title: "Pull", Muscles: []string{"back", "biceps"}},
		},
	}

	updated, err := repo.UpdateSplit(context.Background(), created.UserId, split)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(updated.ExerciseSplit.Elements) != 2 {
		t.Errorf("expected 2 split elements, got %d", len(updated.ExerciseSplit.Elements))
	}
	if updated.ExerciseSplit.Elements[0].Title != "Push" {
		t.Errorf("expected first element title Push, got %s", updated.ExerciseSplit.Elements[0].Title)
	}
}

func TestSqlRepository_UpdateSplit_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.UpdateSplit(context.Background(), 999, domain.Split{})
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

// --- GetByUsername ---

func TestSqlRepository_GetByUsername(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, _ = repo.CreateUser(context.Background(), &domain.User{
		Email:        "byname@example.com",
		Username:     "byname",
		PasswordHash: "stored-hash",
	})

	got, err := repo.GetByUsername(context.Background(), "byname")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Username != "byname" {
		t.Errorf("expected username byname, got %s", got.Username)
	}
	if got.Email != "byname@example.com" {
		t.Errorf("expected email byname@example.com, got %s", got.Email)
	}
	if got.PasswordHash != "stored-hash" {
		t.Errorf("expected PasswordHash stored-hash, got %s", got.PasswordHash)
	}
}

func TestSqlRepository_GetByUsername_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.GetByUsername(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing username, got nil")
	}
}

// --- Refresh token ---

func TestSqlRepository_CreateAndGetRefreshToken(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	user, _ := repo.CreateUser(context.Background(), &domain.User{Email: "rt@example.com", Username: "rtuser"})

	expiresAt := time.Now().Add(time.Hour)
	if err := repo.CreateRefreshToken(context.Background(), user.UserId, "testhash", expiresAt); err != nil {
		t.Fatalf("CreateRefreshToken: %v", err)
	}

	userID, err := repo.GetUserIDByRefreshToken(context.Background(), "testhash")
	if err != nil {
		t.Fatalf("GetUserIDByRefreshToken: %v", err)
	}
	if userID != user.UserId {
		t.Errorf("expected userID %d, got %d", user.UserId, userID)
	}
}

func TestSqlRepository_GetUserIDByRefreshToken_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.GetUserIDByRefreshToken(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}

func TestSqlRepository_GetUserIDByRefreshToken_Expired(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	user, _ := repo.CreateUser(context.Background(), &domain.User{Email: "exp@example.com", Username: "expuser"})

	_, err := db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		user.UserId, "expiredhash", time.Now().Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf("insert expired token: %v", err)
	}

	_, err = repo.GetUserIDByRefreshToken(context.Background(), "expiredhash")
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestSqlRepository_DeleteRefreshToken(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	user, _ := repo.CreateUser(context.Background(), &domain.User{Email: "del2@example.com", Username: "del2user"})

	if err := repo.CreateRefreshToken(context.Background(), user.UserId, "delhash", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateRefreshToken: %v", err)
	}
	if err := repo.DeleteRefreshToken(context.Background(), "delhash"); err != nil {
		t.Fatalf("DeleteRefreshToken: %v", err)
	}

	_, err := repo.GetUserIDByRefreshToken(context.Background(), "delhash")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestSqlRepository_UpdateSplit_PreservesOtherFields(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateUser(context.Background(), &domain.User{
		Email:      "preserve@example.com",
		Username:   "preserve",
		Gender:     "female",
		SportGoals: []string{"cardio"},
	})

	split := domain.Split{Elements: []domain.SplitElement{{Title: "Legs", Muscles: []string{"quads"}}}}
	updated, err := repo.UpdateSplit(context.Background(), created.UserId, split)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Email != "preserve@example.com" {
		t.Errorf("email changed unexpectedly: %s", updated.Email)
	}
	if updated.Gender != "female" {
		t.Errorf("gender changed unexpectedly: %s", updated.Gender)
	}
	if len(updated.SportGoals) != 1 || updated.SportGoals[0] != "cardio" {
		t.Errorf("sport_goals changed unexpectedly: %v", updated.SportGoals)
	}
}
