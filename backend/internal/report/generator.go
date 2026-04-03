package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/akaitigo/shigoto-flow/backend/internal/repository"
)

type Generator struct {
	repo *repository.Repository
}

func NewGenerator(repo *repository.Repository) *Generator {
	return &Generator{repo: repo}
}

func (g *Generator) Generate(ctx context.Context, userID string, tmpl *model.Template, date time.Time, reportType model.ReportType) (string, error) {
	var activities []model.Activity
	var err error

	switch reportType {
	case model.ReportTypeDaily:
		activities, err = g.repo.ListActivitiesByUserAndDate(ctx, userID, date)
	case model.ReportTypeWeekly:
		start := date.AddDate(0, 0, -int(date.Weekday()))
		end := start.AddDate(0, 0, 7)
		activities, err = g.repo.ListActivitiesByUserAndRange(ctx, userID, start, end)
	case model.ReportTypeMonthly:
		start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		end := start.AddDate(0, 1, 0)
		activities, err = g.repo.ListActivitiesByUserAndRange(ctx, userID, start, end)
	default:
		return "", fmt.Errorf("unsupported report type: %s", reportType)
	}

	if err != nil {
		return "", fmt.Errorf("failed to fetch activities: %w", err)
	}

	return renderReport(tmpl, activities, date, reportType), nil
}

func renderReport(tmpl *model.Template, activities []model.Activity, date time.Time, reportType model.ReportType) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — %s\n\n", reportTypeLabel(reportType), date.Format("2006-01-02")))

	grouped := groupBySource(activities)

	for _, section := range tmpl.Sections {
		sb.WriteString(fmt.Sprintf("## %s\n\n", section.Title))

		switch section.Title {
		case "やったこと":
			for source, acts := range grouped {
				sb.WriteString(fmt.Sprintf("### %s\n", sourceLabel(source)))
				for _, a := range acts {
					sb.WriteString(fmt.Sprintf("- %s\n", a.Title))
				}
				sb.WriteString("\n")
			}
		case "わかったこと":
			sb.WriteString("（自動集約されたデータから記入してください）\n\n")
		case "次やること":
			sb.WriteString("（明日の予定を記入してください）\n\n")
		default:
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func reportTypeLabel(rt model.ReportType) string {
	switch rt {
	case model.ReportTypeDaily:
		return "日報"
	case model.ReportTypeWeekly:
		return "週報"
	case model.ReportTypeMonthly:
		return "月報"
	default:
		return string(rt)
	}
}

func sourceLabel(provider model.Provider) string {
	switch provider {
	case model.ProviderGoogle:
		return "Google Calendar"
	case model.ProviderSlack:
		return "Slack"
	case model.ProviderGitHub:
		return "GitHub"
	case model.ProviderGmail:
		return "Gmail"
	default:
		return string(provider)
	}
}

func groupBySource(activities []model.Activity) map[model.Provider][]model.Activity {
	grouped := make(map[model.Provider][]model.Activity)
	for _, a := range activities {
		grouped[a.Source] = append(grouped[a.Source], a)
	}
	return grouped
}
