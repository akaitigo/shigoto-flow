package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/middleware"
	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/akaitigo/shigoto-flow/backend/internal/summary"
)

// summaryService produces AI summaries (weekly/monthly) from stored reports.
// It is an interface so the handler can be tested without a live database or
// Claude API; *summary.Service satisfies it.
type summaryService interface {
	GenerateWeeklySummary(ctx context.Context, userID string, weekStart time.Time) (string, error)
	GenerateMonthlySummary(ctx context.Context, userID string, month time.Time) (string, error)
}

type summarizeRequest struct {
	Type model.ReportType `json:"type"`
	Date string           `json:"date"`
}

type summarizeResponse struct {
	Type    model.ReportType `json:"type"`
	Date    string           `json:"date"`
	Content string           `json:"content"`
}

// SummarizeReport generates a weekly or monthly AI summary from the user's
// existing reports for the period containing the given date.
func (h *Handler) SummarizeReport(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	if h.summarySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "summary service not configured")
		return
	}

	var req summarizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Type != model.ReportTypeWeekly && req.Type != model.ReportTypeMonthly {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "type must be weekly or monthly")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid date format, use YYYY-MM-DD")
		return
	}

	var content string
	switch req.Type {
	case model.ReportTypeWeekly:
		content, err = h.summarySvc.GenerateWeeklySummary(r.Context(), userID, date)
	case model.ReportTypeMonthly:
		content, err = h.summarySvc.GenerateMonthlySummary(r.Context(), userID, date)
	}
	if err != nil {
		if errors.Is(err, summary.ErrNoReports) {
			writeError(w, http.StatusUnprocessableEntity, "NO_REPORTS", "no reports available to summarize for this period")
			return
		}
		slog.Error("failed to generate summary", "type", req.Type, "error", err)
		writeError(w, http.StatusInternalServerError, "SUMMARY_FAILED", "failed to generate summary")
		return
	}

	writeJSON(w, http.StatusOK, summarizeResponse{Type: req.Type, Date: req.Date, Content: content})
}
