package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/akaitigo/shigoto-flow/backend/internal/middleware"
	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	reportType := model.ReportType(r.URL.Query().Get("type"))
	if reportType == "" {
		reportType = model.ReportTypeDaily
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	reports, err := h.repo.ListReportsByUser(r.Context(), userID, reportType, limit, offset)
	if err != nil {
		slog.Error("failed to list reports", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list reports")
		return
	}

	writeJSON(w, http.StatusOK, reports)
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	id := r.PathValue("id")
	report, err := h.repo.GetReportByID(r.Context(), id)
	if err != nil {
		slog.Error("failed to get report", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get report")
		return
	}
	if report == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "report not found")
		return
	}

	if report.UserID != userID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "report not found")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

type createReportRequest struct {
	Type       model.ReportType `json:"type"`
	TemplateID string           `json:"template_id"`
	Content    string           `json:"content"`
	Date       string           `json:"date"`
}

func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	var req createReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if !isValidReportType(req.Type) {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "type must be daily, weekly, or monthly")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid date format, use YYYY-MM-DD")
		return
	}

	now := time.Now()
	report := &model.Report{
		ID:         uuid.New().String(),
		UserID:     userID,
		Type:       req.Type,
		TemplateID: req.TemplateID,
		Content:    req.Content,
		Date:       date,
		Status:     model.ReportStatusDraft,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.repo.CreateReport(r.Context(), report); err != nil {
		slog.Error("failed to create report", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create report")
		return
	}

	writeJSON(w, http.StatusCreated, report)
}

type updateReportRequest struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

func (h *Handler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	id := r.PathValue("id")

	report, err := h.repo.GetReportByID(r.Context(), id)
	if err != nil || report == nil || report.UserID != userID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "report not found")
		return
	}

	var req updateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Status != "" && !isValidStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "status must be draft, confirmed, or sent")
		return
	}

	if err := h.repo.UpdateReportContent(r.Context(), id, req.Content, req.Status); err != nil {
		slog.Error("failed to update report", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update report")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func isValidReportType(t model.ReportType) bool {
	switch t {
	case model.ReportTypeDaily, model.ReportTypeWeekly, model.ReportTypeMonthly:
		return true
	default:
		return false
	}
}

func isValidStatus(s string) bool {
	switch s {
	case model.ReportStatusDraft, model.ReportStatusConfirmed, model.ReportStatusSent:
		return true
	default:
		return false
	}
}
