package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/middleware"
)

func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	dateStr := r.URL.Query().Get("date")
	date := time.Now()
	if dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid date format, use YYYY-MM-DD")
			return
		}
	}

	activities, err := h.repo.ListActivitiesByUserAndDate(r.Context(), userID, date)
	if err != nil {
		slog.Error("failed to list activities", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list activities")
		return
	}

	writeJSON(w, http.StatusOK, activities)
}

func (h *Handler) CollectActivities(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	if h.collectorSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "collector service not configured")
		return
	}

	result, err := h.collectorSvc.CollectForUser(r.Context(), userID, time.Now())
	if err != nil {
		slog.Error("failed to collect activities", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to collect activities")
		return
	}

	resp := map[string]any{
		"status":    "completed",
		"collected": result.Collected,
		"errors":    len(result.Errors),
	}

	writeJSON(w, http.StatusOK, resp)
}
