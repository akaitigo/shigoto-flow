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
	UserInfoURL  string
	Scopes       []string
	RedirectURL  string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// UserInfo holds user profile information retrieved from an OAuth provider.
type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
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

// StartStateCleanup starts a background goroutine that periodically removes
// expired OAuth state entries. Call the returned cancel function to stop it.
func (m *OAuthManager) StartStateCleanup(ctx context.Context, interval time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.CleanupExpiredStates()
			}
		}
	}()
	return cancel
}

// FetchUserInfo retrieves user profile information from the provider's userinfo
// endpoint using the given access token.
func (m *OAuthManager) FetchUserInfo(ctx context.Context, provider model.Provider, accessToken string) (*UserInfo, error) {
	cfg, ok := m.providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
	if cfg.UserInfoURL == "" {
		return nil, fmt.Errorf("provider %s has no userinfo URL configured", provider)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	switch provider {
	case model.ProviderGitHub:
		req.Header.Set("Authorization", "token "+accessToken)
	default:
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	info := &UserInfo{}

	// Parse email - different providers use different field names
	if emailRaw, ok := raw["email"]; ok {
		var email string
		if err := json.Unmarshal(emailRaw, &email); err == nil {
			info.Email = email
		}
	}

	// Parse name - try "name" then "login" (GitHub)
	if nameRaw, ok := raw["name"]; ok {
		var name string
		if err := json.Unmarshal(nameRaw, &name); err == nil {
			info.Name = name
		}
	}
	if info.Name == "" {
		if loginRaw, ok := raw["login"]; ok {
			var login string
			if err := json.Unmarshal(loginRaw, &login); err == nil {
				info.Name = login
			}
		}
	}

	if info.Email == "" {
		return nil, fmt.Errorf("provider %s did not return an email address", provider)
	}

	return info, nil
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
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes: []string{
			"openid",
			"email",
			"profile",
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
		UserInfoURL:  "https://slack.com/api/users.identity",
		Scopes:       []string{"identity.basic", "identity.email", "search:read", "users:read"},
		RedirectURL:  redirectBase + "/api/v1/auth/slack/callback",
	}
}

func DefaultGitHubConfig(clientID, clientSecret, redirectBase string) ProviderConfig {
	return ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scopes:       []string{"repo", "user:email"},
		RedirectURL:  redirectBase + "/api/v1/auth/github/callback",
	}
}
