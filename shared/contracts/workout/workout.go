package workout

import "time"

type SetModel struct {
	WorkoutID  int       `json:"workout_id"`
	ExerciseID int       `json:"exercise_id"`
	Reps       int       `json:"reps"`
	Difficulty int       `json:"difficulty"`
	Weight     float64   `json:"weight"`
	LoggedAt   time.Time `json:"logged_at"`
}

type WorkoutModel struct {
	WorkoutID  int       `json:"workout_id"`
	UserID     int       `json:"user_id"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}
