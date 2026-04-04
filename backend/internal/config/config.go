package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port               int
	FrontendURL        string
	BackendURL         string
	TokenEncryptionKey string
	JWTSecret          string
	DB                 DBConfig
	Google             OAuthConfig
	Slack              OAuthConfig
	GitHub             OAuthConfig
	Anthropic          AnthropicConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
}

type AnthropicConfig struct {
	APIKey string
}

func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	return &Config{
		Port:               port,
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		BackendURL:         getEnv("BACKEND_URL", "http://localhost:8080"),
		TokenEncryptionKey: os.Getenv("TOKEN_ENCRYPTION_KEY"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "shigoto"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "shigoto_flow"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Google: OAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		},
		Slack: OAuthConfig{
			ClientID:     os.Getenv("SLACK_CLIENT_ID"),
			ClientSecret: os.Getenv("SLACK_CLIENT_SECRET"),
		},
		GitHub: OAuthConfig{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		},
		Anthropic: AnthropicConfig{
			APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
