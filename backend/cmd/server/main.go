package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/auth"
	"github.com/akaitigo/shigoto-flow/backend/internal/config"
	"github.com/akaitigo/shigoto-flow/backend/internal/handler"
	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/akaitigo/shigoto-flow/backend/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := repository.NewDB(cfg.DB)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	encryptor, err := auth.NewTokenEncryptor([]byte(cfg.TokenEncryptionKey))
	if err != nil {
		slog.Warn("token encryption disabled", "error", err)
	}

	oauthMgr := auth.NewOAuthManager(nil)
	backendURL := fmt.Sprintf("http://localhost:%d", cfg.Port)

	if cfg.Google.ClientID != "" {
		oauthMgr.RegisterProvider(model.ProviderGoogle, auth.DefaultGoogleConfig(
			cfg.Google.ClientID, cfg.Google.ClientSecret, backendURL,
		))
	}
	if cfg.Slack.ClientID != "" {
		oauthMgr.RegisterProvider(model.ProviderSlack, auth.DefaultSlackConfig(
			cfg.Slack.ClientID, cfg.Slack.ClientSecret, backendURL,
		))
	}
	if cfg.GitHub.ClientID != "" {
		oauthMgr.RegisterProvider(model.ProviderGitHub, auth.DefaultGitHubConfig(
			cfg.GitHub.ClientID, cfg.GitHub.ClientSecret, backendURL,
		))
	}

	repo := repository.New(db)
	h := handler.New(repo, cfg, oauthMgr, encryptor)
	router := h.Routes()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
