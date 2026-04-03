package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (h *Handler) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := model.Provider(r.PathValue("provider"))

	authURL, _, err := h.oauth.AuthURL(provider)
	if err != nil {
		slog.Error("failed to generate auth URL", "provider", provider, "error", err)
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "unsupported provider")
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

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
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

	http.Redirect(w, r, h.cfg.FrontendURL+"/settings?connected="+string(provider), http.StatusTemporaryRedirect)
}
