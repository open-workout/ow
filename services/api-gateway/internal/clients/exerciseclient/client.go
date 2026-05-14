package exerciseclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Exercise struct {
	ExerciseID       int64    `json:"exercise_id"`
	Name             string   `json:"name"`
	ExerciseType     string   `json:"exercise_type"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Description      string   `json:"description"`
	UserID           int64    `json:"user_id"`
	IsPrivate        bool     `json:"is_private"`
	WeightDirection  int64    `json:"weight_direction"`
}

type TopExercisesRequest struct {
	Muscles map[string]float64 `json:"muscles"`
	UserID  int64              `json:"user_id"`
	Limit   int                `json:"limit"`
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

func (c *Client) CreateExercise(ctx context.Context, exercise Exercise) (*Exercise, error) {
	body, err := json.Marshal(exercise)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/exercises", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("exercise-service returned %d", resp.StatusCode)
	}

	var created Exercise
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *Client) ListExercises(ctx context.Context, userID int64) ([]Exercise, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/exercises?user_id=%d", c.baseURL, userID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exercise-service returned %d", resp.StatusCode)
	}

	var exercises []Exercise
	if err := json.NewDecoder(resp.Body).Decode(&exercises); err != nil {
		return nil, err
	}
	return exercises, nil
}

func (c *Client) GetTopExercises(ctx context.Context, req TopExercisesRequest) ([]Exercise, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/exercises/recommendations", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exercise-service returned %d", resp.StatusCode)
	}

	var exercises []Exercise
	if err := json.NewDecoder(resp.Body).Decode(&exercises); err != nil {
		return nil, err
	}
	return exercises, nil
}

func (c *Client) GetExerciseById(ctx context.Context, id int64, callerUserID int64) (*Exercise, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/exercises/%d", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", callerUserID))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("exercise not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exercise-service returned %d", resp.StatusCode)
	}

	var ex Exercise
	if err := json.NewDecoder(resp.Body).Decode(&ex); err != nil {
		return nil, err
	}
	return &ex, nil
}

func (c *Client) AddExerciseMedia(ctx context.Context, exerciseID, userID int64, filename, mimeType string, file io.Reader) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, file); err != nil {
		return err
	}
	mw.WriteField("user_id", fmt.Sprintf("%d", userID))
	mw.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/exercises/%d/media", c.baseURL, exerciseID), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("exercise-service returned %d", resp.StatusCode)
	}
	return nil
}
