package handler

import (
	"log/slog"
	"net/http"

	"github.com/akaitigo/shigoto-flow/backend/internal/middleware"
	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (h *Handler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	sources, err := h.repo.ListDataSourcesByUser(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list data sources", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list data sources")
		return
	}

	writeJSON(w, http.StatusOK, sources)
}

func (h *Handler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	provider := model.Provider(r.PathValue("provider"))
	if !isValidProvider(provider) {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "unsupported provider")
		return
	}

	if err := h.repo.DeleteDataSource(r.Context(), userID, provider); err != nil {
		slog.Error("failed to delete data source", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete data source")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
