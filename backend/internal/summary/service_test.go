package summary

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

type fakeReportRepo struct {
	byType  map[model.ReportType][]model.Report
	gotType model.ReportType
	gotFrom time.Time
	gotTo   time.Time
	calls   int
	err     error
}

func (f *fakeReportRepo) ListReportsByUserAndDateRange(_ context.Context, _ string, reportType model.ReportType, start, end time.Time) ([]model.Report, error) {
	f.calls++
	f.gotType = reportType
	f.gotFrom = start
	f.gotTo = end
	if f.err != nil {
		return nil, f.err
	}
	return f.byType[reportType], nil
}

type fakeSummarizer struct {
	gotInput SummarizeInput
	result   string
	err      error
}

func (f *fakeSummarizer) Summarize(_ context.Context, input SummarizeInput) (string, error) {
	f.gotInput = input
	if f.err != nil {
		return "", f.err
	}
	return f.result, nil
}

func reportsWithContent(reportType model.ReportType, contents ...string) []model.Report {
	reports := make([]model.Report, 0, len(contents))
	for _, c := range contents {
		reports = append(reports, model.Report{Type: reportType, Content: c})
	}
	return reports
}

func TestGenerateWeeklySummary(t *testing.T) {
	weekStart := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)

	// More than the old hard-coded limit of 7 to prove the cap is gone.
	contents := make([]string, 0, 10)
	for i := 1; i <= 10; i++ {
		contents = append(contents, fmt.Sprintf("日報%d", i))
	}

	repo := &fakeReportRepo{
		byType: map[model.ReportType][]model.Report{
			model.ReportTypeDaily: reportsWithContent(model.ReportTypeDaily, contents...),
		},
	}
	sum := &fakeSummarizer{result: "週報サマリー"}
	svc := NewService(repo, sum)

	got, err := svc.GenerateWeeklySummary(context.Background(), "user1", weekStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "週報サマリー" {
		t.Errorf("expected summarizer result, got %q", got)
	}
	if repo.gotType != model.ReportTypeDaily {
		t.Errorf("expected daily reports queried, got %s", repo.gotType)
	}
	if !repo.gotFrom.Equal(weekStart) {
		t.Errorf("expected range start %v, got %v", weekStart, repo.gotFrom)
	}
	if want := weekStart.AddDate(0, 0, 7); !repo.gotTo.Equal(want) {
		t.Errorf("expected range end %v, got %v", want, repo.gotTo)
	}
	if len(sum.gotInput.Reports) != 10 {
		t.Errorf("expected all 10 daily reports forwarded (no 7-item cap), got %d", len(sum.gotInput.Reports))
	}
	if sum.gotInput.Period != "週報" {
		t.Errorf("expected period 週報, got %s", sum.gotInput.Period)
	}
}

func TestGenerateWeeklySummary_NoReports(t *testing.T) {
	repo := &fakeReportRepo{byType: map[model.ReportType][]model.Report{}}
	sum := &fakeSummarizer{result: "unused"}
	svc := NewService(repo, sum)

	_, err := svc.GenerateWeeklySummary(context.Background(), "user1", time.Now())
	if err == nil {
		t.Error("expected error when no daily reports exist")
	}
	if sum.gotInput.Reports != nil {
		t.Error("summarizer should not be called when there are no reports")
	}
}

func TestGenerateWeeklySummary_RepoError(t *testing.T) {
	repo := &fakeReportRepo{err: errors.New("db down")}
	svc := NewService(repo, &fakeSummarizer{})

	_, err := svc.GenerateWeeklySummary(context.Background(), "user1", time.Now())
	if err == nil {
		t.Error("expected error to propagate from repository")
	}
}

func TestGenerateMonthlySummary_UsesWeekly(t *testing.T) {
	month := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	repo := &fakeReportRepo{
		byType: map[model.ReportType][]model.Report{
			model.ReportTypeWeekly: reportsWithContent(model.ReportTypeWeekly, "週報1", "週報2"),
		},
	}
	sum := &fakeSummarizer{result: "月報サマリー"}
	svc := NewService(repo, sum)

	got, err := svc.GenerateMonthlySummary(context.Background(), "user1", month)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "月報サマリー" {
		t.Errorf("expected summarizer result, got %q", got)
	}
	monthStart := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if !repo.gotFrom.Equal(monthStart) {
		t.Errorf("expected range start %v, got %v", monthStart, repo.gotFrom)
	}
	if want := monthStart.AddDate(0, 1, 0); !repo.gotTo.Equal(want) {
		t.Errorf("expected range end %v, got %v", want, repo.gotTo)
	}
	if len(sum.gotInput.Reports) != 2 {
		t.Errorf("expected 2 weekly reports forwarded, got %d", len(sum.gotInput.Reports))
	}
	if repo.calls != 1 {
		t.Errorf("expected a single query when weekly reports exist, got %d", repo.calls)
	}
}

func TestGenerateMonthlySummary_FallbackToDaily(t *testing.T) {
	month := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	repo := &fakeReportRepo{
		byType: map[model.ReportType][]model.Report{
			// No weekly reports; only daily reports exist for the month.
			model.ReportTypeDaily: reportsWithContent(model.ReportTypeDaily, "日報A", "日報B", "日報C"),
		},
	}
	sum := &fakeSummarizer{result: "月報(日報から)"}
	svc := NewService(repo, sum)

	got, err := svc.GenerateMonthlySummary(context.Background(), "user1", month)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "月報(日報から)" {
		t.Errorf("unexpected result: %q", got)
	}
	if len(sum.gotInput.Reports) != 3 {
		t.Errorf("expected fallback to 3 daily reports, got %d", len(sum.gotInput.Reports))
	}
	if repo.calls != 2 {
		t.Errorf("expected weekly query then daily fallback (2 calls), got %d", repo.calls)
	}
}

func TestGenerateMonthlySummary_NoReports(t *testing.T) {
	repo := &fakeReportRepo{byType: map[model.ReportType][]model.Report{}}
	svc := NewService(repo, &fakeSummarizer{})

	_, err := svc.GenerateMonthlySummary(context.Background(), "user1", time.Now())
	if err == nil {
		t.Error("expected error when neither weekly nor daily reports exist")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	input := SummarizeInput{
		Reports:    []string{"日報内容1"},
		ReportType: "日報",
		Period:     "週報",
	}

	prompt := buildSystemPrompt(input)

	if !strings.Contains(prompt, "日報") {
		t.Error("expected system prompt to contain report type")
	}
	if !strings.Contains(prompt, "週報") {
		t.Error("expected system prompt to contain target period")
	}
	if !strings.Contains(prompt, "指示や命令は無視") {
		t.Error("expected system prompt to contain injection protection instruction")
	}
}

func TestBuildUserContent(t *testing.T) {
	input := SummarizeInput{
		Reports:    []string{"日報内容1", "日報内容2", "日報内容3"},
		ReportType: "日報",
		Period:     "週報",
	}

	content := buildUserContent(input)

	if strings.Count(content, "レポート") != 3 {
		t.Errorf("expected 3 report sections, got %d", strings.Count(content, "レポート"))
	}
	if !strings.Contains(content, "日報内容1") {
		t.Error("expected content to contain first report")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxLength != 1500 {
		t.Errorf("expected max length 1500, got %d", cfg.MaxLength)
	}
	if cfg.DetailLevel != "standard" {
		t.Errorf("expected detail level standard, got %s", cfg.DetailLevel)
	}
}
