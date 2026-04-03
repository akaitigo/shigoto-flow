package model

import "time"

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type DataSource struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Provider     Provider  `json:"provider" db:"provider"`
	AccessToken  string    `json:"-" db:"access_token"`
	RefreshToken string    `json:"-" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderSlack  Provider = "slack"
	ProviderGitHub Provider = "github"
	ProviderGmail  Provider = "gmail"
)

type Template struct {
	ID        string            `json:"id" db:"id"`
	UserID    string            `json:"user_id" db:"user_id"`
	Name      string            `json:"name" db:"name"`
	Type      ReportType        `json:"type" db:"type"`
	Sections  []TemplateSection `json:"sections"`
	IsDefault bool              `json:"is_default" db:"is_default"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"`
}

type TemplateSection struct {
	Title string `json:"title"`
	Order int    `json:"order"`
}

type ReportType string

const (
	ReportTypeDaily   ReportType = "daily"
	ReportTypeWeekly  ReportType = "weekly"
	ReportTypeMonthly ReportType = "monthly"
)

type Report struct {
	ID         string     `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	Type       ReportType `json:"type" db:"type"`
	TemplateID string     `json:"template_id" db:"template_id"`
	Content    string     `json:"content" db:"content"`
	Date       time.Time  `json:"date" db:"date"`
	Status     string     `json:"status" db:"status"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

type Activity struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Source    Provider  `json:"source" db:"source"`
	Title     string    `json:"title" db:"title"`
	Body      string    `json:"body" db:"body"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Metadata  string    `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

const (
	ReportStatusDraft     = "draft"
	ReportStatusConfirmed = "confirmed"
	ReportStatusSent      = "sent"
)
