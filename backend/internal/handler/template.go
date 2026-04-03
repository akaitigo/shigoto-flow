package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/google/uuid"
)

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	templates, err := h.repo.ListTemplatesByUser(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list templates", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list templates")
		return
	}

	writeJSON(w, http.StatusOK, templates)
}

type createTemplateRequest struct {
	Name      string                  `json:"name"`
	Type      model.ReportType        `json:"type"`
	Sections  []model.TemplateSection `json:"sections"`
	IsDefault bool                    `json:"is_default"`
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user ID")
		return
	}

	var req createTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "template name is required")
		return
	}

	now := time.Now()
	tmpl := &model.Template{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		Sections:  req.Sections,
		IsDefault: req.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.repo.CreateTemplate(r.Context(), tmpl); err != nil {
		slog.Error("failed to create template", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create template")
		return
	}

	writeJSON(w, http.StatusCreated, tmpl)
}
