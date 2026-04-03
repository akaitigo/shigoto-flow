package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
