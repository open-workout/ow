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
		user_id TEXT,
		is_private BOOLEAN,
		weight_direction BIGINT
	);

	CREATE TABLE exercise_media (
		exercise_id BIGINT,
		url TEXT,
		user_id TEXT
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
}

const pgTestUser1 = "auth0|user-1"
const pgTestUser2 = "auth0|user-2"
const pgTestUser99 = "auth0|user-99"

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
		UserID:           pgTestUser1,
		IsPrivate:        false,
		WeightDirection:  1,
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
		UserID:           pgTestUser1,
		IsPrivate:        false,
		WeightDirection:  1,
	}

	result, err := repo.CreateExercise(context.Background(), ex)

	if err != nil {
		t.Fatalf("failed to create exercise: %v", err)
	}

	media := &domain.ExerciseMedia{
		ExerciseID: result.ExerciseID,
		URL:        "http://image.com/squat.png",
		UserID:     pgTestUser1,
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
		UserID:           pgTestUser1,
		IsPrivate:        false,
		WeightDirection:  1,
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
		UserID:           pgTestUser1,
		IsPrivate:        true,
		WeightDirection:  1,
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
		UserID:           pgTestUser2,
		IsPrivate:        false,
		WeightDirection:  1,
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

func TestSqlRepository_UpdateExercise(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, err := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name:             "Pull Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"biceps"},
		Description:      "pull movement",
		UserID:           pgTestUser1,
		IsPrivate:        false,
		WeightDirection:  1,
	})
	if err != nil {
		t.Fatalf("failed to create exercise: %v", err)
	}

	created.Name = "Chin Up"
	created.IsPrivate = true

	updated, err := repo.UpdateExercise(context.Background(), pgTestUser1, created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Name != "Chin Up" {
		t.Errorf("expected name Chin Up, got %s", updated.Name)
	}
	if !updated.IsPrivate {
		t.Errorf("expected IsPrivate=true")
	}
	if updated.UserID != pgTestUser1 {
		t.Errorf("expected UserID=%s, got %s", pgTestUser1, updated.UserID)
	}
}

func TestSqlRepository_UpdateExercise_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.UpdateExercise(context.Background(), pgTestUser1, &domain.ExerciseModel{ExerciseID: 999, Name: "Ghost"})
	if err == nil {
		t.Fatal("expected error for missing exercise, got nil")
	}
}

func TestSqlRepository_UpdateExercise_WrongUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Pull Up", UserID: pgTestUser1, IsPrivate: false,
	})

	_, err := repo.UpdateExercise(context.Background(), pgTestUser99, &domain.ExerciseModel{
		ExerciseID: created.ExerciseID, Name: "Hacked",
	})
	if err == nil {
		t.Fatal("expected error when updating with wrong user, got nil")
	}

	var name string
	db.QueryRow(`SELECT name FROM exercises WHERE exercise_id = $1`, created.ExerciseID).Scan(&name)
	if name != "Pull Up" {
		t.Errorf("exercise should be unchanged, got name=%s", name)
	}
}

func TestSqlRepository_DeleteExercise(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, err := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name:      "Squat",
		UserID:    pgTestUser1,
		IsPrivate: false,
	})
	if err != nil {
		t.Fatalf("failed to create exercise: %v", err)
	}

	if err := repo.DeleteExercise(context.Background(), pgTestUser1, created.ExerciseID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM exercises WHERE exercise_id = $1`, created.ExerciseID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query db: %v", err)
	}
	if count != 0 {
		t.Errorf("expected exercise to be deleted, still found %d rows", count)
	}
}

func TestSqlRepository_DeleteExercise_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	err := repo.DeleteExercise(context.Background(), pgTestUser1, 999)
	if err == nil {
		t.Fatal("expected error for missing exercise, got nil")
	}
}

func TestSqlRepository_DeleteExercise_WrongUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Squat", UserID: pgTestUser1, IsPrivate: false,
	})

	err := repo.DeleteExercise(context.Background(), pgTestUser99, created.ExerciseID)
	if err == nil {
		t.Fatal("expected error when deleting with wrong user, got nil")
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM exercises WHERE exercise_id = $1`, created.ExerciseID).Scan(&count)
	if count != 1 {
		t.Fatal("exercise should still exist after failed delete")
	}
}

func TestSqlRepository_GetExerciseById_Public(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, err := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Deadlift", UserID: pgTestUser1, IsPrivate: false,
	})
	if err != nil {
		t.Fatalf("failed to create exercise: %v", err)
	}

	result, err := repo.GetExerciseById(context.Background(), created.ExerciseID, pgTestUser99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExerciseID != created.ExerciseID {
		t.Errorf("expected ID %d, got %d", created.ExerciseID, result.ExerciseID)
	}
}

func TestSqlRepository_GetExerciseById_Private_Owner(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Secret Move", UserID: pgTestUser1, IsPrivate: true,
	})

	result, err := repo.GetExerciseById(context.Background(), created.ExerciseID, pgTestUser1)
	if err != nil {
		t.Fatalf("owner should be able to get their private exercise: %v", err)
	}
	if result.ExerciseID != created.ExerciseID {
		t.Errorf("expected ID %d, got %d", created.ExerciseID, result.ExerciseID)
	}
}

func TestSqlRepository_GetExerciseById_Private_WrongUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	created, _ := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Secret Move", UserID: pgTestUser1, IsPrivate: true,
	})

	_, err := repo.GetExerciseById(context.Background(), created.ExerciseID, pgTestUser99)
	if err == nil {
		t.Fatal("expected error for private exercise accessed by wrong user, got nil")
	}
}

func TestSqlRepository_GetExerciseById_NotFound(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	_, err := repo.GetExerciseById(context.Background(), 999, pgTestUser1)
	if err == nil {
		t.Fatal("expected error for missing exercise, got nil")
	}
}

func TestSqlRepository_GetExerciseMedia_ReturnsMedia(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	ex, _ := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Pull Up", UserID: pgTestUser1, IsPrivate: false,
	})

	_ = repo.AddExerciseMedia(context.Background(), ex.ExerciseID, &domain.ExerciseMedia{
		ExerciseID: ex.ExerciseID, URL: "http://example.com/1.jpg", UserID: pgTestUser1,
	})
	_ = repo.AddExerciseMedia(context.Background(), ex.ExerciseID, &domain.ExerciseMedia{
		ExerciseID: ex.ExerciseID, URL: "http://example.com/2.png", UserID: pgTestUser1,
	})

	media, err := repo.GetExerciseMedia(context.Background(), ex.ExerciseID, pgTestUser1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(media) != 2 {
		t.Errorf("expected 2 media items, got %d", len(media))
	}
	for _, m := range media {
		if m.URL == "" {
			t.Error("expected non-empty URL")
		}
	}
}

func TestSqlRepository_GetExerciseMedia_PrivateExercise_Owner(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	ex, _ := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Secret Move", UserID: pgTestUser1, IsPrivate: true,
	})
	_ = repo.AddExerciseMedia(context.Background(), ex.ExerciseID, &domain.ExerciseMedia{
		ExerciseID: ex.ExerciseID, URL: "http://example.com/secret.jpg", UserID: pgTestUser1,
	})

	media, err := repo.GetExerciseMedia(context.Background(), ex.ExerciseID, pgTestUser1)
	if err != nil {
		t.Fatalf("owner should see their private exercise media: %v", err)
	}
	if len(media) != 1 {
		t.Errorf("expected 1 media item, got %d", len(media))
	}
}

func TestSqlRepository_GetExerciseMedia_PrivateExercise_WrongUser(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()
	setupSchema(t, db)
	repo := repository.NewSqlRepository(db)

	ex, _ := repo.CreateExercise(context.Background(), &domain.ExerciseModel{
		Name: "Secret Move", UserID: pgTestUser1, IsPrivate: true,
	})
	_ = repo.AddExerciseMedia(context.Background(), ex.ExerciseID, &domain.ExerciseMedia{
		ExerciseID: ex.ExerciseID, URL: "http://example.com/secret.jpg", UserID: pgTestUser1,
	})

	media, err := repo.GetExerciseMedia(context.Background(), ex.ExerciseID, pgTestUser99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(media) != 0 {
		t.Errorf("expected 0 media items for wrong user on private exercise, got %d", len(media))
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
		UserID:           pgTestUser1,
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
		UserID:           pgTestUser1,
		IsPrivate:        false,
		WeightDirection:  1,
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
