package summary

import (
	"context"
	"fmt"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/akaitigo/shigoto-flow/backend/internal/repository"
)

type Service struct {
	repo       *repository.Repository
	summarizer *Summarizer
}

func NewService(repo *repository.Repository, summarizer *Summarizer) *Service {
	return &Service{repo: repo, summarizer: summarizer}
}

func (s *Service) GenerateWeeklySummary(ctx context.Context, userID string, weekStart time.Time) (string, error) {
	weekEnd := weekStart.AddDate(0, 0, 7)

	reports, err := s.repo.ListReportsByUser(ctx, userID, model.ReportTypeDaily, 7, 0)
	if err != nil {
		return "", fmt.Errorf("failed to fetch daily reports: %w", err)
	}

	var dailyContents []string
	for _, r := range reports {
		if r.Date.After(weekStart) && r.Date.Before(weekEnd) {
			dailyContents = append(dailyContents, r.Content)
		}
	}

	if len(dailyContents) == 0 {
		return "", fmt.Errorf("no daily reports found for the specified week")
	}

	return s.summarizer.Summarize(ctx, SummarizeInput{
		Reports:    dailyContents,
		ReportType: "日報",
		Period:     "週報",
	})
}

func (s *Service) GenerateMonthlySummary(ctx context.Context, userID string, month time.Time) (string, error) {
	reports, err := s.repo.ListReportsByUser(ctx, userID, model.ReportTypeWeekly, 5, 0)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weekly reports: %w", err)
	}

	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	var weeklyContents []string
	for _, r := range reports {
		if r.Date.After(monthStart) && r.Date.Before(monthEnd) {
			weeklyContents = append(weeklyContents, r.Content)
		}
	}

	if len(weeklyContents) == 0 {
		dailyReports, err := s.repo.ListReportsByUser(ctx, userID, model.ReportTypeDaily, 31, 0)
		if err != nil {
			return "", fmt.Errorf("failed to fetch daily reports: %w", err)
		}

		for _, r := range dailyReports {
			if r.Date.After(monthStart) && r.Date.Before(monthEnd) {
				weeklyContents = append(weeklyContents, r.Content)
			}
		}
	}

	if len(weeklyContents) == 0 {
		return "", fmt.Errorf("no reports found for the specified month")
	}

	return s.summarizer.Summarize(ctx, SummarizeInput{
		Reports:    weeklyContents,
		ReportType: "週報",
		Period:     "月報",
	})
}
