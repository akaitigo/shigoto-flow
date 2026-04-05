package handler

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/akaitigo/shigoto-flow/backend/internal/middleware"
	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (h *Handler) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := model.Provider(r.PathValue("provider"))
	if !isValidProvider(provider) {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "unsupported provider")
		return
	}

	authURL, _, err := h.oauth.AuthURL(provider)
	if err != nil {
		slog.Error("failed to generate auth URL", "provider", provider, "error", err)
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "provider not configured")
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "missing state or code parameter")
		return
	}

	provider, err := h.oauth.ValidateState(state)
	if err != nil {
		slog.Error("invalid OAuth state", "error", err)
		writeError(w, http.StatusBadRequest, "INVALID_STATE", "invalid or expired state parameter")
		return
	}

	tokenResp, err := h.oauth.ExchangeCode(r.Context(), provider, code)
	if err != nil {
		slog.Error("failed to exchange OAuth code", "provider", provider, "error", err)
		writeError(w, http.StatusInternalServerError, "TOKEN_EXCHANGE_FAILED", "failed to exchange authorization code")
		return
	}

	// Fetch real user info from the provider's userinfo endpoint
	userInfo, err := h.oauth.FetchUserInfo(r.Context(), provider, tokenResp.AccessToken)
	if err != nil {
		slog.Error("failed to fetch user info from provider", "provider", provider, "error", err)
		writeError(w, http.StatusInternalServerError, "USERINFO_FAILED", "failed to retrieve user information from provider")
		return
	}

	// Look up existing user by email, or create a new one
	existingUser, err := h.repo.GetUserByEmail(r.Context(), userInfo.Email)
	if err != nil {
		slog.Error("failed to look up user by email", "email", userInfo.Email, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to look up user")
		return
	}

	var userID string
	if existingUser != nil {
		userID = existingUser.ID
		// Update name if changed
		if existingUser.Name != userInfo.Name && userInfo.Name != "" {
			existingUser.Name = userInfo.Name
			if err := h.repo.UpdateUser(r.Context(), existingUser); err != nil {
				slog.Warn("failed to update user name", "error", err)
			}
		}
	} else {
		userID = uuid.New().String()
		now := time.Now()
		user := &model.User{
			ID:        userID,
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.repo.CreateUser(r.Context(), user); err != nil {
			slog.Error("failed to create user", "error", err)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
			return
		}
	}

	encAccessToken, err := h.encryptor.Encrypt(tokenResp.AccessToken)
	if err != nil {
		slog.Error("failed to encrypt access token", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store credentials")
		return
	}

	encRefreshToken, err := h.encryptor.Encrypt(tokenResp.RefreshToken)
	if err != nil {
		slog.Error("failed to encrypt refresh token", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store credentials")
		return
	}

	now := time.Now()
	ds := &model.DataSource{
		ID:           uuid.New().String(),
		UserID:       userID,
		Provider:     provider,
		AccessToken:  encAccessToken,
		RefreshToken: encRefreshToken,
		ExpiresAt:    now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.repo.UpsertDataSource(r.Context(), ds); err != nil {
		slog.Error("failed to save data source", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save credentials")
		return
	}

	// Issue JWT token for the user
	jwtToken, err := middleware.GenerateToken([]byte(h.cfg.JWTSecret), userID, 24*time.Hour)
	if err != nil {
		slog.Error("failed to generate JWT", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate session token")
		return
	}

	// Set JWT as HttpOnly cookie instead of URL query parameter (RFC 6750 compliance)
	frontendURL, parseErr := url.Parse(h.cfg.FrontendURL)
	isSecure := parseErr == nil && frontendURL.Scheme == "https"
	sameSite := http.SameSiteLaxMode
	if isSecure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    jwtToken,
		Path:     "/",
		Domain:   cookieDomain(h.cfg.FrontendURL),
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: sameSite,
	})

	http.Redirect(w, r, h.cfg.FrontendURL+"/settings?connected="+string(provider), http.StatusTemporaryRedirect)
}

// cookieDomain extracts the hostname from a URL for use as cookie domain.
// Returns empty string if parsing fails (browser will use the response origin).
func cookieDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	// Don't set domain for localhost (browsers reject it)
	if host == "localhost" || strings.HasPrefix(host, "127.") {
		return ""
	}
	return host
}

// isValidProvider checks whether the given provider is supported.
// Gmail is NOT a separate provider; it is covered by the Google provider's scopes.
func isValidProvider(p model.Provider) bool {
	switch p {
	case model.ProviderGoogle, model.ProviderSlack, model.ProviderGitHub:
		return true
	default:
		return false
	}
}
