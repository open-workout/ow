package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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

func TestSqlRepository_CreateExercise(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	setupSchema(t, db)

	repo := repository.NewSqlRepository(db)

	ex := &domain.ExerciseModel{
		Name:             "Pull Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"biceps"},
		Description:      "pull movement",
		UserID:           1,
		IsPrivate:        false,
	}

	result, err := repo.CreateExercise(context.Background(), ex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var name string
	err = db.QueryRow(`SELECT name FROM exercises WHERE exercise_id = $1`, result.ExerciseID).
		Scan(&name)

	if err != nil {
		t.Fatalf("failed to query db: %v", err)
	}

	if name != "Pull Up" {
		t.Errorf("expected name Pull Up, got %s", name)
	}
}

func TestSqlRepository_AddExerciseMedia(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	ex := &domain.ExerciseModel{
		Name:             "Pull Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"biceps"},
		Description:      "pull movement",
		UserID:           1,
		IsPrivate:        false,
	}

	result, err := repo.CreateExercise(context.Background(), ex)

	if err != nil {
		t.Fatalf("failed to create exercise: %v", err)
	}

	media := &domain.ExerciseMedia{
		ExerciseID: result.ExerciseID,
		URL:        "http://image.com/squat.png",
		UserID:     1,
	}

	err = repo.AddExerciseMedia(context.Background(), result.ExerciseID, media)
	if err != nil {
		t.Fatalf("failed to add media to exercise: %v", err)
	}

	var url string
	err = db.QueryRow(`
		SELECT url FROM exercise_media
		WHERE exercise_id = $1`, result.ExerciseID,
	).Scan(&url)
	if err != nil {
		t.Fatalf("failed to query exercise: %v", err)
	}

	if url != "http://image.com/squat.png" {
		t.Errorf("expected url http://image.com/squat.png, got %s", url)
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

func TestSqlRepository_GetUserExercises(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	ex1 := &domain.ExerciseModel{
		Name:             "Push Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "chest",
		SecondaryMuscles: []string{"triceps"},
		Description:      "basic push movement",
		UserID:           1,
		IsPrivate:        true,
	}
	_, err := repo.CreateExercise(context.Background(), ex1)
	if err != nil {
		t.Fatalf("failed to create exercise ex1: %v", err)
	}

	ex2 := &domain.ExerciseModel{
		Name:             "Bench Press",
		ExerciseType:     "compound",
		PrimaryMuscle:    "chest",
		SecondaryMuscles: []string{"triceps", "shoulders"},
		Description:      "basic push movement",
		UserID:           2,
		IsPrivate:        false,
	}

	_, err = repo.CreateExercise(context.Background(), ex2)
	if err != nil {
		t.Fatalf("failed to create exercise ex2: %v", err)
	}

	exercisesOfUser1, err := repo.ListUserExercises(context.Background(), ex1.UserID)
	if err != nil {
		t.Fatalf("failed to list exercises: %v", err)
	}

	if len(exercisesOfUser1) != 1 {
		t.Errorf("expected 1 exercise of user1, got %d", len(exercisesOfUser1))
	}

	if exercisesOfUser1[0].ExerciseID != 1 {
		t.Fatalf("expected exercise %v for user 1, but got %v", ex1.ExerciseID, exercisesOfUser1[0].ExerciseID)
	}

}

func TestSqlRepository_GetPublicExercises(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)
	ex1 := &domain.ExerciseModel{
		Name:             "Push Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "chest",
		SecondaryMuscles: []string{"triceps"},
		Description:      "basic push movement",
		UserID:           1,
		IsPrivate:        true,
	}

	_, err := repo.CreateExercise(context.Background(), ex1)
	if err != nil {
		t.Fatalf("failed to create exercise ex1: %v", err)
	}

	ex2 := &domain.ExerciseModel{
		Name:             "Bench Press",
		ExerciseType:     "compound",
		PrimaryMuscle:    "chest",
		SecondaryMuscles: []string{"triceps", "shoulders"},
		Description:      "basic push movement",
		UserID:           1,
		IsPrivate:        false,
	}

	_, err = repo.CreateExercise(context.Background(), ex2)
	if err != nil {
		t.Fatalf("failed to create exercise ex2: %v", err)
	}

	publicExercises, err := repo.ListPublicExercises(context.Background())
	if err != nil {
		t.Fatalf("failed to list public exercises: %v", err)
	}

	if len(publicExercises) != 1 {
		t.Fatalf("expected 1 public exercises, got %d", len(publicExercises))
	}

	if publicExercises[0].ExerciseID != ex2.ExerciseID {
		t.Fatalf("expected %v as public, got %d", ex2.ExerciseID, publicExercises[0].ExerciseID)
	}

}

//TODO: add tests for getting all user/public exercises.
