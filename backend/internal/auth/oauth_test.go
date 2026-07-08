package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func TestOAuthManager_AuthURL(t *testing.T) {
	mgr := NewOAuthManager(nil)
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID:    "test-client-id",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		Scopes:      []string{"calendar.readonly"},
		RedirectURL: "http://localhost:8080/api/v1/auth/google/callback",
	})

	authURL, state, err := mgr.AuthURL(model.ProviderGoogle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if authURL == "" {
		t.Error("expected non-empty auth URL")
	}

	if state == "" {
		t.Error("expected non-empty state")
	}

	if len(authURL) < 50 {
		t.Error("auth URL seems too short")
	}
}

func TestOAuthManager_AuthURL_UnknownProvider(t *testing.T) {
	mgr := NewOAuthManager(nil)

	_, _, err := mgr.AuthURL("unknown")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestOAuthManager_ValidateState(t *testing.T) {
	mgr := NewOAuthManager(nil)
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID: "test",
		AuthURL:  "https://example.com/auth",
	})

	_, state, err := mgr.AuthURL(model.ProviderGoogle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	provider, err := mgr.ValidateState(state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider != model.ProviderGoogle {
		t.Errorf("expected provider google, got %s", provider)
	}
}

func TestOAuthManager_ValidateState_Invalid(t *testing.T) {
	mgr := NewOAuthManager(nil)

	_, err := mgr.ValidateState("invalid-state")
	if err == nil {
		t.Error("expected error for invalid state")
	}
}

func TestOAuthManager_ValidateState_ReplayPrevention(t *testing.T) {
	mgr := NewOAuthManager(nil)
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID: "test",
		AuthURL:  "https://example.com/auth",
	})

	_, state, err := mgr.AuthURL(model.ProviderGoogle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = mgr.ValidateState(state)
	if err != nil {
		t.Fatalf("first validation should succeed: %v", err)
	}

	_, err = mgr.ValidateState(state)
	if err == nil {
		t.Error("second validation should fail (replay prevention)")
	}
}

func TestOAuthManager_ExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		TokenURL:     server.URL,
		RedirectURL:  "http://localhost/callback",
	})

	resp, err := mgr.ExchangeCode(context.Background(), model.ProviderGoogle, "test-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.AccessToken != "test-access-token" {
		t.Errorf("expected test-access-token, got %s", resp.AccessToken)
	}

	if resp.RefreshToken != "test-refresh-token" {
		t.Errorf("expected test-refresh-token, got %s", resp.RefreshToken)
	}
}

func TestOAuthManager_ExchangeCode_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"error":"invalid_grant"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID:     "test",
		ClientSecret: "test",
		TokenURL:     server.URL,
		RedirectURL:  "http://localhost/callback",
	})

	_, err := mgr.ExchangeCode(context.Background(), model.ProviderGoogle, "bad-code")
	if err == nil {
		t.Error("expected error for failed exchange")
	}
}

func TestOAuthManager_RefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "new-access-token",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID:     "test",
		ClientSecret: "test",
		TokenURL:     server.URL,
	})

	resp, err := mgr.RefreshToken(context.Background(), model.ProviderGoogle, "old-refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.AccessToken != "new-access-token" {
		t.Errorf("expected new-access-token, got %s", resp.AccessToken)
	}
}

func TestOAuthManager_CleanupExpiredStates(t *testing.T) {
	mgr := NewOAuthManager(nil)
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID: "test",
		AuthURL:  "https://example.com/auth",
	})

	_, state, err := mgr.AuthURL(model.ProviderGoogle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mgr.CleanupExpiredStates()

	_, err = mgr.ValidateState(state)
	if err != nil {
		t.Error("recent state should not be cleaned up")
	}
}

func TestOAuthManager_StartStateCleanup(t *testing.T) {
	mgr := NewOAuthManager(nil)
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID: "test",
		AuthURL:  "https://example.com/auth",
	})

	ctx := context.Background()
	cancel := mgr.StartStateCleanup(ctx, 50*time.Millisecond)

	// Add an expired state entry manually
	mgr.mu.Lock()
	mgr.states["expired-state"] = stateEntry{
		provider:  model.ProviderGoogle,
		createdAt: time.Now().Add(-15 * time.Minute),
	}
	mgr.mu.Unlock()

	// Wait for the cleanup goroutine to run at least once
	time.Sleep(150 * time.Millisecond)

	mgr.mu.Lock()
	_, exists := mgr.states["expired-state"]
	mgr.mu.Unlock()

	if exists {
		t.Error("expired state should have been cleaned up by background goroutine")
	}

	// Stop the cleanup goroutine
	cancel()
}

func TestOAuthManager_FetchUserInfo_Google(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-access-token" {
			t.Errorf("expected Bearer test-access-token, got %s", authHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"email":"user@example.com","name":"Test User"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID:    "test",
		UserInfoURL: server.URL,
	})

	info, err := mgr.FetchUserInfo(context.Background(), model.ProviderGoogle, "test-access-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Email != "user@example.com" {
		t.Errorf("expected user@example.com, got %s", info.Email)
	}
	if info.Name != "Test User" {
		t.Errorf("expected Test User, got %s", info.Name)
	}
}

func TestOAuthManager_FetchUserInfo_GitHub(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "token gh-token" {
			t.Errorf("expected token gh-token, got %s", authHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"email":"dev@github.com","name":"","login":"octocat"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderGitHub, ProviderConfig{
		ClientID:    "test",
		UserInfoURL: server.URL,
	})

	info, err := mgr.FetchUserInfo(context.Background(), model.ProviderGitHub, "gh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Email != "dev@github.com" {
		t.Errorf("expected dev@github.com, got %s", info.Email)
	}
	if info.Name != "octocat" {
		t.Errorf("expected octocat (fallback to login), got %s", info.Name)
	}
}

func TestOAuthManager_FetchUserInfo_NoEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"name":"No Email User"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID:    "test",
		UserInfoURL: server.URL,
	})

	_, err := mgr.FetchUserInfo(context.Background(), model.ProviderGoogle, "token")
	if err == nil {
		t.Error("expected error when email is missing")
	}
}

func TestOAuthManager_FetchUserInfo_NoUserInfoURL(t *testing.T) {
	mgr := NewOAuthManager(nil)
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID: "test",
	})

	_, err := mgr.FetchUserInfo(context.Background(), model.ProviderGoogle, "token")
	if err == nil {
		t.Error("expected error when userinfo URL is not configured")
	}
}

// TestOAuthManager_AuthURL_AccessType verifies that access_type=offline is only
// sent to Google, since it is a Google-specific parameter (issue #23).
func TestOAuthManager_AuthURL_AccessType(t *testing.T) {
	mgr := NewOAuthManager(nil)
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID: "g", AuthURL: "https://accounts.google.com/o/oauth2/v2/auth",
	})
	mgr.RegisterProvider(model.ProviderSlack, ProviderConfig{
		ClientID: "s", AuthURL: "https://slack.com/oauth/v2/authorize",
	})
	mgr.RegisterProvider(model.ProviderGitHub, ProviderConfig{
		ClientID: "gh", AuthURL: "https://github.com/login/oauth/authorize",
	})

	tests := []struct {
		provider model.Provider
		wantSent bool
	}{
		{model.ProviderGoogle, true},
		{model.ProviderSlack, false},
		{model.ProviderGitHub, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			authURL, _, err := mgr.AuthURL(tt.provider)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			u, err := url.Parse(authURL)
			if err != nil {
				t.Fatalf("failed to parse auth URL: %v", err)
			}
			got := u.Query().Get("access_type")
			if tt.wantSent && got != "offline" {
				t.Errorf("expected access_type=offline for %s, got %q", tt.provider, got)
			}
			if !tt.wantSent && got != "" {
				t.Errorf("expected no access_type for %s, got %q", tt.provider, got)
			}
		})
	}
}

// TestOAuthManager_ExchangeCode_SlackAuthedUser verifies that Slack's user
// token (authed_user.access_token) is used rather than the top-level bot
// token (issue #17).
func TestOAuthManager_ExchangeCode_SlackAuthedUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"ok": true,
			"access_token": "xoxb-bot-token",
			"token_type": "bot",
			"authed_user": {
				"id": "U123",
				"access_token": "xoxp-user-token",
				"token_type": "user",
				"refresh_token": "xoxp-refresh",
				"expires_in": 43200
			}
		}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderSlack, ProviderConfig{
		ClientID:     "test",
		ClientSecret: "test",
		TokenURL:     server.URL,
		RedirectURL:  "http://localhost/callback",
	})

	resp, err := mgr.ExchangeCode(context.Background(), model.ProviderSlack, "code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "xoxp-user-token" {
		t.Errorf("expected user token xoxp-user-token, got %s", resp.AccessToken)
	}
	if resp.RefreshToken != "xoxp-refresh" {
		t.Errorf("expected user refresh token xoxp-refresh, got %s", resp.RefreshToken)
	}
	if resp.ExpiresIn != 43200 {
		t.Errorf("expected expires_in 43200, got %d", resp.ExpiresIn)
	}
}

// TestOAuthManager_ExchangeCode_GoogleKeepsTopLevel ensures the promotion of
// authed_user only applies to Slack and does not affect other providers.
func TestOAuthManager_ExchangeCode_GoogleKeepsTopLevel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "google-access",
			RefreshToken: "google-refresh",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderGoogle, ProviderConfig{
		ClientID: "test", ClientSecret: "test", TokenURL: server.URL,
	})

	resp, err := mgr.ExchangeCode(context.Background(), model.ProviderGoogle, "code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "google-access" {
		t.Errorf("expected google-access, got %s", resp.AccessToken)
	}
}

// TestOAuthManager_FetchUserInfo_Slack verifies that the email/name are parsed
// from Slack's nested "user" object (issue #21).
func TestOAuthManager_FetchUserInfo_Slack(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer xoxp-user-token" {
			t.Errorf("expected Bearer xoxp-user-token, got %s", authHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"ok":true,"user":{"id":"U123","name":"Slack User","email":"user@slack.example"},"team":{"id":"T1"}}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	mgr := NewOAuthManager(server.Client())
	mgr.RegisterProvider(model.ProviderSlack, ProviderConfig{
		ClientID:    "test",
		UserInfoURL: server.URL,
	})

	info, err := mgr.FetchUserInfo(context.Background(), model.ProviderSlack, "xoxp-user-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Email != "user@slack.example" {
		t.Errorf("expected user@slack.example, got %s", info.Email)
	}
	if info.Name != "Slack User" {
		t.Errorf("expected Slack User, got %s", info.Name)
	}
}
