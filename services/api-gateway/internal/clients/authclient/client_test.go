package authclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-workout/ow/services/api-gateway/internal/clients/authclient"
)

// --- Login ---

func TestAuthClient_Login_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/auth/login" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"user_id": 42, "refresh_token": "opaque-tok"})
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	userID, token, err := c.Login(context.Background(), "john", "pw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 42 {
		t.Errorf("expected userID 42, got %d", userID)
	}
	if token != "opaque-tok" {
		t.Errorf("expected token opaque-tok, got %s", token)
	}
}

func TestAuthClient_Login_InvalidCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	_, _, err := c.Login(context.Background(), "john", "wrong")
	if !errors.Is(err, authclient.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthClient_Login_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oops", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	_, _, err := c.Login(context.Background(), "john", "pw")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

// --- Refresh ---

func TestAuthClient_Refresh_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/auth/refresh" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"user_id": 7})
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	userID, err := c.Refresh(context.Background(), "my-refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 7 {
		t.Errorf("expected userID 7, got %d", userID)
	}
}

func TestAuthClient_Refresh_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	_, err := c.Refresh(context.Background(), "bad-token")
	if !errors.Is(err, authclient.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthClient_Refresh_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oops", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	_, err := c.Refresh(context.Background(), "token")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

// --- Logout ---

func TestAuthClient_Logout_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/auth/logout" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	if err := c.Logout(context.Background(), "my-refresh-token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthClient_Logout_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oops", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := authclient.New(srv.URL)
	if err := c.Logout(context.Background(), "token"); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}
