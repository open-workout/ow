package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMockDB(t *testing.T) (*repository.SqlRepository, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	repo := repository.NewSqlRepository(db)
	return repo, mock, db
}

func TestSqlRepository_CreateExercise(t *testing.T) {
	repo, mock, db := setupMockDB(t)
	defer db.Close()

	ex := &domain.ExerciseModel{
		Name:             "Push Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "chest",
		SecondaryMuscles: []string{"triceps"},
		Description:      "basic push exercise",
		UserID:           1,
		IsPrivate:        false,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
	INSERT INTO exercises  (name, exercise_type, primary_muscle, secondary_muscles, description, user_id, is_private)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING exercise_id
	`)).
		WithArgs(
			ex.Name,
			ex.ExerciseType,
			ex.PrimaryMuscle,
			pq.Array(ex.SecondaryMuscles),
			ex.Description,
			ex.UserID,
			ex.IsPrivate,
		).
		WillReturnRows(sqlmock.NewRows([]string{"exercise_id"}).AddRow(42))

	result, err := repo.CreateExercise(context.Background(), ex)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExerciseID != 42 {
		t.Errorf("expected ID 42, got %d", result.ExerciseID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSqlRepository_UpdateExercise(t *testing.T) {
	repo, mock, db := setupMockDB(t)
	defer db.Close()

	ex := &domain.ExerciseModel{
		ExerciseID:       10,
		Name:             "Pull Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"biceps"},
		Description:      "pull movement",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
	UPDATE exercises
	SET
		name = COALESCE(name, $1),
		exercise_type = COALESCE(exercise_type, $2),
		primary_muscle = COALESCE(primary_muscle, $3),
		secondary_muscles = COALESCE(secondary_muscles, $4),
		description = COALESCE(description, $5)
	WHERE exercise_id = $6
	RETURNING exercise_id
	`)).
		WithArgs(
			ex.Name,
			ex.ExerciseType,
			ex.PrimaryMuscle,
			pq.Array(ex.SecondaryMuscles),
			ex.Description,
			ex.ExerciseID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"exercise_id"}).AddRow(10))

	result, err := repo.UpdateExercise(context.Background(), ex)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExerciseID != 10 {
		t.Errorf("expected ID 10, got %d", result.ExerciseID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSqlRepository_AddExerciseMedia(t *testing.T) {
	repo, mock, db := setupMockDB(t)
	defer db.Close()

	media := &domain.ExerciseMedia{
		ExerciseID: 1,
		URL:        "https://cdn.com/file.jpg",
		UserID:     99,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO exercise_media (exercise_id, url, user_id) 
		VALUES ($1, $2, $3)
	`)).
		WithArgs(
			media.ExerciseID,
			media.URL,
			media.UserID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddExerciseMedia(context.Background(), 1, media)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func setupPostgres(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf(
		"postgres://test:test@%s:%s/testdb?sslmode=disable",
		host,
		port.Port(),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// wait for db readiness
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	cleanup := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, cleanup
}
func setupSchema(t *testing.T, db *sql.DB) {
	schema := `
	CREATE TABLE exercises (
		exercise_id SERIAL PRIMARY KEY,
		name TEXT,
		exercise_type TEXT,
		primary_muscle TEXT,
		secondary_muscles TEXT[],
		description TEXT,
		user_id BIGINT,
		is_private BOOLEAN
	);

	CREATE TABLE exercise_media (
		exercise_id BIGINT,
		url TEXT,
		user_id BIGINT
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
}

func TestSqlRepository_CreateExercise_Integration(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	setupSchema(t, db)

	repo := repository.NewSqlRepository(db)

	ex := &domain.ExerciseModel{
		Name:             "Push Up",
		ExerciseType:     "strength",
		PrimaryMuscle:    "chest",
		SecondaryMuscles: []string{"triceps"},
		Description:      "basic push movement",
		UserID:           1,
		IsPrivate:        false,
	}

	result, err := repo.CreateExercise(context.Background(), ex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExerciseID == 0 {
		t.Errorf("expected valid ID, got 0")
	}
}
