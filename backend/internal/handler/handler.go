package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/akaitigo/shigoto-flow/backend/internal/auth"
	"github.com/akaitigo/shigoto-flow/backend/internal/collector"
	"github.com/akaitigo/shigoto-flow/backend/internal/config"
	"github.com/akaitigo/shigoto-flow/backend/internal/middleware"
	"github.com/akaitigo/shigoto-flow/backend/internal/repository"
)

type Handler struct {
	repo         *repository.Repository
	cfg          *config.Config
	oauth        *auth.OAuthManager
	encryptor    *auth.TokenEncryptor
	collectorSvc *collector.Service
}

func New(repo *repository.Repository, cfg *config.Config, oauth *auth.OAuthManager, encryptor *auth.TokenEncryptor) *Handler {
	return &Handler{repo: repo, cfg: cfg, oauth: oauth, encryptor: encryptor}
}

func (h *Handler) SetCollectorService(svc *collector.Service) {
	h.collectorSvc = svc
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", h.Health)
	mux.HandleFunc("GET /api/v1/reports", h.ListReports)
	mux.HandleFunc("POST /api/v1/reports", h.CreateReport)
	mux.HandleFunc("GET /api/v1/reports/{id}", h.GetReport)
	mux.HandleFunc("PUT /api/v1/reports/{id}", h.UpdateReport)
	mux.HandleFunc("POST /api/v1/reports/generate", h.GenerateReport)
	mux.HandleFunc("GET /api/v1/activities", h.ListActivities)
	mux.HandleFunc("POST /api/v1/activities/collect", h.CollectActivities)
	mux.HandleFunc("GET /api/v1/templates", h.ListTemplates)
	mux.HandleFunc("POST /api/v1/templates", h.CreateTemplate)
	mux.HandleFunc("GET /api/v1/datasources", h.ListDataSources)
	mux.HandleFunc("DELETE /api/v1/datasources/{provider}", h.DeleteDataSource)
	mux.HandleFunc("GET /api/v1/auth/{provider}/callback", h.OAuthCallback)
	mux.HandleFunc("GET /api/v1/auth/{provider}", h.OAuthRedirect)

	return corsMiddleware(h.cfg.FrontendURL)(middleware.Auth(maxBodySize(mux)))
}

func corsMiddleware(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func maxBodySize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
		"code":  code,
	})
}
