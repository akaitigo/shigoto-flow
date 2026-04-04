package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth_AllowsPublicPaths(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		path string
	}{
		{"/api/v1/health"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("path %s: expected 200, got %d", tt.path, rec.Code)
		}
	}
}

func TestAuth_RejectsWithoutAuth(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_AcceptsXUserID(t *testing.T) {
	var gotUserID string
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.Header.Set("X-User-ID", "user-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected user-123, got %s", gotUserID)
	}
}

func TestAuth_AcceptsBearerToken(t *testing.T) {
	var gotUserID string
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	req.Header.Set("Authorization", "Bearer user-456")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID != "user-456" {
		t.Errorf("expected user-456, got %s", gotUserID)
	}
}

func TestIsPublicPath(t *testing.T) {
	tests := []struct {
		path   string
		public bool
	}{
		{"/api/v1/health", true},
		{"/api/v1/auth/google", false},
		{"/api/v1/auth/slack/callback", false},
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
