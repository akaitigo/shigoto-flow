package report

import (
	"strings"
	"testing"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func TestRenderReport_Daily(t *testing.T) {
	tmpl := &model.Template{
		Sections: []model.TemplateSection{
			{Title: "やったこと", Order: 1},
			{Title: "わかったこと", Order: 2},
			{Title: "次やること", Order: 3},
		},
	}

	activities := []model.Activity{
		{Source: model.ProviderGoogle, Title: "チームミーティング"},
		{Source: model.ProviderGitHub, Title: "repo/project にプッシュ"},
		{Source: model.ProviderSlack, Title: "#dev での投稿"},
	}

	date := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)
	result := renderReport(tmpl, activities, date, model.ReportTypeDaily)

	if !strings.Contains(result, "日報") {
		t.Error("expected report to contain '日報'")
	}
	if !strings.Contains(result, "2026-04-04") {
		t.Error("expected report to contain date")
	}
	if !strings.Contains(result, "チームミーティング") {
		t.Error("expected report to contain calendar activity")
	}
	if !strings.Contains(result, "repo/project にプッシュ") {
		t.Error("expected report to contain github activity")
	}
	if !strings.Contains(result, "やったこと") {
		t.Error("expected report to contain section header")
	}
	if !strings.Contains(result, "わかったこと") {
		t.Error("expected report to contain section header")
	}
	if !strings.Contains(result, "次やること") {
		t.Error("expected report to contain section header")
	}
}

func TestRenderReport_EmptyActivities(t *testing.T) {
	tmpl := &model.Template{
		Sections: []model.TemplateSection{
			{Title: "やったこと", Order: 1},
		},
	}

	date := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)
	result := renderReport(tmpl, nil, date, model.ReportTypeDaily)

	if !strings.Contains(result, "日報") {
		t.Error("expected report to contain '日報'")
	}
	if !strings.Contains(result, "やったこと") {
		t.Error("expected report to contain section header")
	}
}

func TestReportTypeLabel(t *testing.T) {
	tests := []struct {
		input model.ReportType
		want  string
	}{
		{model.ReportTypeDaily, "日報"},
		{model.ReportTypeWeekly, "週報"},
		{model.ReportTypeMonthly, "月報"},
	}

	for _, tt := range tests {
		if got := reportTypeLabel(tt.input); got != tt.want {
			t.Errorf("reportTypeLabel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGroupBySource(t *testing.T) {
	activities := []model.Activity{
		{Source: model.ProviderGoogle, Title: "a"},
		{Source: model.ProviderGoogle, Title: "b"},
		{Source: model.ProviderSlack, Title: "c"},
	}

	grouped := groupBySource(activities)

	if len(grouped[model.ProviderGoogle]) != 2 {
		t.Errorf("expected 2 google activities, got %d", len(grouped[model.ProviderGoogle]))
	}
	if len(grouped[model.ProviderSlack]) != 1 {
		t.Errorf("expected 1 slack activity, got %d", len(grouped[model.ProviderSlack]))
	}
}
