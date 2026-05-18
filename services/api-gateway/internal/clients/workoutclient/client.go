package workoutclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("not found")

type WorkoutModel struct {
	WorkoutID  int64     `json:"workout_id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

type SetModel struct {
	SetID      int64     `json:"set_id"`
	WorkoutID  int64     `json:"workout_id"`
	ExerciseID int64     `json:"exercise_id"`
	Reps       int       `json:"reps"`
	Difficulty int       `json:"difficulty"`
	Weight     float64   `json:"weight"`
	Unit       string    `json:"unit"`
	LoggedAt   time.Time `json:"logged_at"`
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetWorkoutById(ctx context.Context, userID string, workoutID int64) (*WorkoutModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/workouts/%d", c.baseURL, workoutID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-ID", userID)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workout-service returned %d", resp.StatusCode)
	}

	var workout WorkoutModel
	if err := json.NewDecoder(resp.Body).Decode(&workout); err != nil {
		return nil, err
	}
	return &workout, nil
}

func (c *Client) GetSetsByWorkoutID(ctx context.Context, userID string, workoutID int64) ([]SetModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/workouts/%d/sets", c.baseURL, workoutID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-ID", userID)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workout-service returned %d", resp.StatusCode)
	}

	var sets []SetModel
	if err := json.NewDecoder(resp.Body).Decode(&sets); err != nil {
		return nil, err
	}
	return sets, nil
}

func (c *Client) UpdateSet(ctx context.Context, userID string, set SetModel) (*SetModel, error) {
	body, err := json.Marshal(set)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/sets/%d", c.baseURL, set.SetID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workout-service returned %d", resp.StatusCode)
	}

	var updated SetModel
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *Client) DeleteSet(ctx context.Context, userID string, setID int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/sets/%d", c.baseURL, setID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-User-ID", userID)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("workout-service returned %d", resp.StatusCode)
	}
	return nil
}
