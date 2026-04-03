package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port               int
	FrontendURL        string
	TokenEncryptionKey string
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
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
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
		TokenEncryptionKey: getEnv("TOKEN_ENCRYPTION_KEY", "01234567890123456789012345678901"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "shigoto"),
			Password: getEnv("DB_PASSWORD", "shigoto_dev"),
			Name:     getEnv("DB_NAME", "shigoto_flow"),
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
