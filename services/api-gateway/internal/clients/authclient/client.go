package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid or expired token")

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	UserID       int64  `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	UserID int64 `json:"user_id"`
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

func (c *Client) Login(ctx context.Context, username, password string) (userID int64, refreshToken string, err error) {
	body, _ := json.Marshal(loginRequest{Username: username, Password: password})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/auth/login", bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return 0, "", ErrInvalidCredentials
	}
	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("user-service returned %d", resp.StatusCode)
	}

	var result loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", err
	}
	return result.UserID, result.RefreshToken, nil
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (int64, error) {
	body, _ := json.Marshal(refreshRequest{RefreshToken: refreshToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/auth/refresh", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return 0, ErrInvalidToken
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("user-service returned %d", resp.StatusCode)
	}

	var result refreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.UserID, nil
}

func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	body, _ := json.Marshal(refreshRequest{RefreshToken: refreshToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/auth/logout", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("user-service returned %d", resp.StatusCode)
	}
	return nil
}
