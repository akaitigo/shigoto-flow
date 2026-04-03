package handler

import (
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
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
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":  "accepted",
		"message": "activity collection started",
	})
}
