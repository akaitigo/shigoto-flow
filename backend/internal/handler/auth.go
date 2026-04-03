package handler

import "net/http"

func (h *Handler) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	_ = provider
	writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "OAuth redirect not yet implemented")
}

func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	_ = provider
	writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "OAuth callback not yet implemented")
}
