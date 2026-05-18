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
}

const pgUser1 = "auth0|pg-1"
const pgUser2 = "auth0|pg-2"

func TestSqlRepository_CreateUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	u := &domain.User{
		UserId:     pgUser1,
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

	if created.UserId != pgUser1 {
		t.Errorf("expected user_id %s, got %s", pgUser1, created.UserId)
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

	u := &domain.User{UserId: pgUser1, Email: "bob@example.com", Username: "bob"}

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

	u := &domain.User{UserId: pgUser1, Email: "dup@example.com", Username: "dup"}
	_, err := repo.CreateUser(context.Background(), u)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	u2 := &domain.User{UserId: pgUser2, Email: "dup@example.com", Username: "dup2"}
	_, err = repo.CreateUser(context.Background(), u2)
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
}

func TestSqlRepository_GetUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.CreateUser(context.Background(), &domain.User{
		UserId: pgUser1, Email: "get@example.com", Username: "getuser",
		Gender: "male", Birthdate: "1985-01-01",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetUser(context.Background(), pgUser1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserId != pgUser1 {
		t.Errorf("expected ID %s, got %s", pgUser1, got.UserId)
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

	_, err := repo.GetUser(context.Background(), "auth0|nobody")
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
		UserId: pgUser1, Email: "update@example.com", Username: "updateuser",
		Gender: "male", Birthdate: "2000-06-01",
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

	_, err := repo.UpdateUser(context.Background(), &domain.User{UserId: "auth0|ghost", Email: "ghost@example.com", Username: "ghost"})
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestSqlRepository_DeleteUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.CreateUser(context.Background(), &domain.User{UserId: pgUser1, Email: "del@example.com", Username: "del"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.DeleteUser(context.Background(), pgUser1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM users WHERE user_id = $1`, pgUser1).Scan(&count)
	if count != 0 {
		t.Errorf("expected user to be deleted, got %d rows", count)
	}
}

func TestSqlRepository_DeleteUser_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	err := repo.DeleteUser(context.Background(), "auth0|nobody")
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestSqlRepository_UpdateSplit(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.CreateUser(context.Background(), &domain.User{UserId: pgUser1, Email: "split@example.com", Username: "splituser"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	split := domain.Split{
		Elements: []domain.SplitElement{
			{Title: "Push", Muscles: []string{"chest", "triceps"}},
			{Title: "Pull", Muscles: []string{"back", "biceps"}},
		},
	}

	updated, err := repo.UpdateSplit(context.Background(), pgUser1, split)
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

	_, err := repo.UpdateSplit(context.Background(), "auth0|nobody", domain.Split{})
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestSqlRepository_UpdateSplit_PreservesOtherFields(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.CreateUser(context.Background(), &domain.User{
		UserId:     pgUser1,
		Email:      "preserve@example.com",
		Username:   "preserve",
		Gender:     "female",
		SportGoals: []string{"cardio"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	split := domain.Split{Elements: []domain.SplitElement{{Title: "Legs", Muscles: []string{"quads"}}}}
	updated, err := repo.UpdateSplit(context.Background(), pgUser1, split)
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
