package middleware_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

const testKID = "test-key"
const testAudience = "test-audience"

func b64url(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func makeJWKS(pub *rsa.PublicKey, kid string) []byte {
	expBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(expBytes, uint32(pub.E))
	i := 0
	for i < len(expBytes)-1 && expBytes[i] == 0 {
		i++
	}
	data := map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"kid": kid,
			"n":   b64url(pub.N.Bytes()),
			"e":   b64url(expBytes[i:]),
			"alg": "RS256",
			"use": "sig",
		}},
	}
	b, _ := json.Marshal(data)
	return b
}

func signRS256(t *testing.T, priv *rsa.PrivateKey, kid, sub, issuer, audience string) string {
	t.Helper()
	header, _ := json.Marshal(map[string]any{"alg": "RS256", "typ": "JWT", "kid": kid})
	payload, _ := json.Marshal(map[string]any{
		"sub": sub,
		"iss": issuer,
		"aud": audience,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	msg := b64url(header) + "." + b64url(payload)
	h := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return msg + "." + b64url(sig)
}

func setupJWKSServer(t *testing.T, priv *rsa.PrivateKey, kid string) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	jwksData := makeJWKS(&priv.PublicKey, kid)
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"issuer":%q,"jwks_uri":%q}`, srv.URL+"/", srv.URL+"/.well-known/jwks.json")
	})
	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksData)
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestAuth_ValidToken(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := setupJWKSServer(t, priv, testKID)
	issuer := srv.URL + "/"

	tok := signRS256(t, priv, testKID, "auth0|user-123", issuer, testAudience)

	handler := mw.Auth(issuer, testAudience)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.GetUserID(r) != "auth0|user-123" {
			t.Errorf("got userID %q, want auth0|user-123", mw.GetUserID(r))
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rr.Code)
	}
}

func TestAuth_MissingToken(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := setupJWKSServer(t, priv, testKID)
	issuer := srv.URL + "/"

	handler := mw.Auth(issuer, testAudience)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", rr.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := setupJWKSServer(t, priv, testKID)
	issuer := srv.URL + "/"

	handler := mw.Auth(issuer, testAudience)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", rr.Code)
	}
}

func TestAuth_WrongKey(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	otherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := setupJWKSServer(t, priv, testKID)
	issuer := srv.URL + "/"

	tok := signRS256(t, otherKey, testKID, "auth0|user-1", issuer, testAudience)

	handler := mw.Auth(issuer, testAudience)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", rr.Code)
	}
}
