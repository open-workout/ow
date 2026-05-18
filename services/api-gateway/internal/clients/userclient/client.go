package userclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var ErrNotFound = errors.New("not found")

type SplitElement struct {
	Muscles []string `json:"muscles"`
	Title   string   `json:"title"`
}

type Split struct {
	Elements []SplitElement `json:"elements"`
}

type User struct {
	UserID        string   `json:"user_id"`
	Email         string   `json:"email"`
	SportGoals    []string `json:"sport_goals"`
	Gender        string   `json:"gender"`
	Birthdate     string   `json:"birthdate"`
	ExerciseSplit Split    `json:"split"`
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/users/%s", c.baseURL, url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) CreateUser(ctx context.Context, callerUserID string, user User) (*User, error) {
	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/users", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", callerUserID)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("user-service returned %d", resp.StatusCode)
	}

	var created User
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *Client) UpdateUser(ctx context.Context, callerUserID, id string, user User) (*User, error) {
	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/users/%s", c.baseURL, url.PathEscape(id)), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", callerUserID)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned %d", resp.StatusCode)
	}

	var updated User
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *Client) DeleteUser(ctx context.Context, callerUserID, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/users/%s", c.baseURL, url.PathEscape(id)), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-User-ID", callerUserID)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("forbidden")
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("user-service returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) UpdateSplit(ctx context.Context, callerUserID, id string, split Split) (*User, error) {
	body, err := json.Marshal(split)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/users/%s/split", c.baseURL, url.PathEscape(id)), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", callerUserID)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned %d", resp.StatusCode)
	}

	var updated User
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
}
