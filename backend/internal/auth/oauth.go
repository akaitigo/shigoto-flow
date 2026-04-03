package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	Scopes       []string
	RedirectURL  string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type OAuthManager struct {
	providers map[model.Provider]ProviderConfig
	states    map[string]stateEntry
	mu        sync.Mutex
	client    *http.Client
}

type stateEntry struct {
	provider  model.Provider
	createdAt time.Time
}

func NewOAuthManager(client *http.Client) *OAuthManager {
	if client == nil {
		client = http.DefaultClient
	}
	return &OAuthManager{
		providers: make(map[model.Provider]ProviderConfig),
		states:    make(map[string]stateEntry),
		client:    client,
	}
}

func (m *OAuthManager) RegisterProvider(provider model.Provider, cfg ProviderConfig) {
	m.providers[provider] = cfg
}

func (m *OAuthManager) AuthURL(provider model.Provider) (string, string, error) {
	cfg, ok := m.providers[provider]
	if !ok {
		return "", "", fmt.Errorf("unknown provider: %s", provider)
	}

	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	m.mu.Lock()
	m.states[state] = stateEntry{provider: provider, createdAt: time.Now()}
	m.mu.Unlock()

	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(cfg.Scopes, " "))
	params.Set("state", state)
	params.Set("access_type", "offline")

	return cfg.AuthURL + "?" + params.Encode(), state, nil
}

func (m *OAuthManager) ValidateState(state string) (model.Provider, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.states[state]
	if !ok {
		return "", fmt.Errorf("invalid state parameter")
	}

	delete(m.states, state)

	if time.Since(entry.createdAt) > 10*time.Minute {
		return "", fmt.Errorf("state expired")
	}

	return entry.provider, nil
}

func (m *OAuthManager) ExchangeCode(ctx context.Context, provider model.Provider, code string) (*TokenResponse, error) {
	cfg, ok := m.providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("redirect_uri", cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

func (m *OAuthManager) RefreshToken(ctx context.Context, provider model.Provider, refreshToken string) (*TokenResponse, error) {
	cfg, ok := m.providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed with %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &tokenResp, nil
}

func (m *OAuthManager) CleanupExpiredStates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for state, entry := range m.states {
		if time.Since(entry.createdAt) > 10*time.Minute {
			delete(m.states, state)
		}
	}
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func DefaultGoogleConfig(clientID, clientSecret, redirectBase string) ProviderConfig {
	return ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.readonly",
			"https://www.googleapis.com/auth/gmail.readonly",
		},
		RedirectURL: redirectBase + "/api/v1/auth/google/callback",
	}
}

func DefaultSlackConfig(clientID, clientSecret, redirectBase string) ProviderConfig {
	return ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://slack.com/oauth/v2/authorize",
		TokenURL:     "https://slack.com/api/oauth.v2.access",
		Scopes:       []string{"search:read", "users:read"},
		RedirectURL:  redirectBase + "/api/v1/auth/slack/callback",
	}
}

func DefaultGitHubConfig(clientID, clientSecret, redirectBase string) ProviderConfig {
	return ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		Scopes:       []string{"repo", "user:email"},
		RedirectURL:  redirectBase + "/api/v1/auth/github/callback",
	}
}
