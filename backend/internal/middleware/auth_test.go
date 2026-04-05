package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var testSecret = []byte("test-secret-key-for-auth-testing")

func TestAuthWithSecret_AllowsPublicPaths(t *testing.T) {
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthWithSecret_RejectsWithoutAuth(t *testing.T) {
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthWithSecret_RejectsRawUserIDHeader(t *testing.T) {
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.Header.Set("X-User-ID", "attacker-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for raw X-User-ID header, got %d", rec.Code)
	}
}

func TestAuthWithSecret_AcceptsValidBearerToken(t *testing.T) {
	var gotUserID string
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	token, err := GenerateToken(testSecret, "user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected user-123, got %s", gotUserID)
	}
}

func TestAuthWithSecret_RejectsExpiredToken(t *testing.T) {
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token, err := GenerateToken(testSecret, "user-123", -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", rec.Code)
	}
}

func TestAuthWithSecret_RejectsInvalidSignature(t *testing.T) {
	wrongSecret := []byte("wrong-secret-key-for-auth-tests!")
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token, err := GenerateToken(wrongSecret, "user-123", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid signature, got %d", rec.Code)
	}
}

func TestAuthWithSecret_AcceptsValidCookieToken(t *testing.T) {
	var gotUserID string
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	token, err := GenerateToken(testSecret, "cookie-user-456", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID != "cookie-user-456" {
		t.Errorf("expected cookie-user-456, got %s", gotUserID)
	}
}

func TestAuthWithSecret_CookieTakesPrecedenceOverHeader(t *testing.T) {
	var gotUserID string
	handler := AuthWithSecret(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	cookieToken, err := GenerateToken(testSecret, "cookie-user", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate cookie token: %v", err)
	}
	headerToken, err := GenerateToken(testSecret, "header-user", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate header token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: cookieToken})
	req.Header.Set("Authorization", "Bearer "+headerToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID != "cookie-user" {
		t.Errorf("expected cookie-user (cookie takes precedence), got %s", gotUserID)
	}
}

func TestIsPublicPath(t *testing.T) {
	tests := []struct {
		path   string
		public bool
	}{
		{"/api/v1/health", true},
		{"/api/v1/auth/google", true},
		{"/api/v1/auth/slack/callback", true},
		{"/api/v1/reports", false},
		{"/api/v1/activities", false},
		{"/api/v1/templates", false},
	}

	for _, tt := range tests {
		if got := isPublicPath(tt.path); got != tt.public {
			t.Errorf("isPublicPath(%q) = %v, want %v", tt.path, got, tt.public)
		}
	}
}
