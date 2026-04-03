package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/akaitigo/shigoto-flow/backend/internal/report"
)

type generateReportRequest struct {
	Type model.ReportType `json:"type"`
	Date string           `json:"date"`
}

func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	var req generateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid date format, use YYYY-MM-DD")
		return
	}

	tmpl, err := h.repo.GetDefaultTemplate(r.Context(), userID, req.Type)
	if err != nil {
		slog.Error("failed to get template", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get template")
		return
	}
	if tmpl == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no default template found for this report type")
		return
	}

	generator := report.NewGenerator(h.repo)
	content, err := generator.Generate(r.Context(), userID, tmpl, date, req.Type)
	if err != nil {
		slog.Error("failed to generate report", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate report")
		return
	}

	now := time.Now()
	generatedReport := &model.Report{
		ID:         uuid.New().String(),
		UserID:     userID,
		Type:       req.Type,
		TemplateID: tmpl.ID,
		Content:    content,
		Date:       date,
		Status:     model.ReportStatusDraft,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.repo.CreateReport(r.Context(), generatedReport); err != nil {
		slog.Error("failed to save generated report", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save generated report")
		return
	}

	writeJSON(w, http.StatusCreated, generatedReport)
}
