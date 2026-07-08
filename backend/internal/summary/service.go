package summary

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

// ErrNoReports indicates there were no source reports to summarize for the
// requested period. Callers can treat this as a client-actionable condition
// (the user needs to create reports first) rather than a server error.
var ErrNoReports = errors.New("no reports found for the requested period")

// reportRepository is the subset of the repository used by the summary service.
// Using an interface keeps the service unit-testable without a live database.
type reportRepository interface {
	ListReportsByUserAndDateRange(ctx context.Context, userID string, reportType model.ReportType, start, end time.Time) ([]model.Report, error)
}

// summarizerClient produces a summary from a set of reports.
type summarizerClient interface {
	Summarize(ctx context.Context, input SummarizeInput) (string, error)
}

type Service struct {
	repo       reportRepository
	summarizer summarizerClient
}

func NewService(repo reportRepository, summarizer summarizerClient) *Service {
	return &Service{repo: repo, summarizer: summarizer}
}

func (s *Service) GenerateWeeklySummary(ctx context.Context, userID string, weekStart time.Time) (string, error) {
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Query the exact week window so past weeks can be summarized, rather than
	// only being able to reach the most recent handful of daily reports.
	reports, err := s.repo.ListReportsByUserAndDateRange(ctx, userID, model.ReportTypeDaily, weekStart, weekEnd)
	if err != nil {
		return "", fmt.Errorf("failed to fetch daily reports: %w", err)
	}

	dailyContents := make([]string, 0, len(reports))
	for _, r := range reports {
		dailyContents = append(dailyContents, r.Content)
	}

	if len(dailyContents) == 0 {
		return "", fmt.Errorf("weekly summary: %w", ErrNoReports)
	}

	return s.summarizer.Summarize(ctx, SummarizeInput{
		Reports:    dailyContents,
		ReportType: "日報",
		Period:     "週報",
	})
}

func (s *Service) GenerateMonthlySummary(ctx context.Context, userID string, month time.Time) (string, error) {
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	reports, err := s.repo.ListReportsByUserAndDateRange(ctx, userID, model.ReportTypeWeekly, monthStart, monthEnd)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weekly reports: %w", err)
	}

	contents := make([]string, 0, len(reports))
	for _, r := range reports {
		contents = append(contents, r.Content)
	}

	// Fall back to daily reports when no weekly reports exist for the month.
	if len(contents) == 0 {
		dailyReports, err := s.repo.ListReportsByUserAndDateRange(ctx, userID, model.ReportTypeDaily, monthStart, monthEnd)
		if err != nil {
			return "", fmt.Errorf("failed to fetch daily reports: %w", err)
		}
		for _, r := range dailyReports {
			contents = append(contents, r.Content)
		}
	}

	if len(contents) == 0 {
		return "", fmt.Errorf("monthly summary: %w", ErrNoReports)
	}

	return s.summarizer.Summarize(ctx, SummarizeInput{
		Reports:    contents,
		ReportType: "週報",
		Period:     "月報",
	})
}
