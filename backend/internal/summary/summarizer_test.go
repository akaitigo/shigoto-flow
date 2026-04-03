package summary

import (
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	input := SummarizeInput{
		Reports:    []string{"日報1の内容", "日報2の内容"},
		ReportType: "日報",
		Period:     "週報",
	}

	prompt := buildPrompt(input)

	if !strings.Contains(prompt, "日報") {
		t.Error("expected prompt to contain '日報'")
	}
	if !strings.Contains(prompt, "週報") {
		t.Error("expected prompt to contain '週報'")
	}
	if !strings.Contains(prompt, "日報1の内容") {
		t.Error("expected prompt to contain first report")
	}
	if !strings.Contains(prompt, "日報2の内容") {
		t.Error("expected prompt to contain second report")
	}
	if !strings.Contains(prompt, "レポート 1") {
		t.Error("expected prompt to contain report numbering")
	}
}

func TestBuildPrompt_EmptyReports(t *testing.T) {
	input := SummarizeInput{
		Reports:    []string{},
		ReportType: "週報",
		Period:     "月報",
	}

	prompt := buildPrompt(input)

	if !strings.Contains(prompt, "週報") {
		t.Error("expected prompt to contain report type")
	}
	if !strings.Contains(prompt, "月報") {
		t.Error("expected prompt to contain period")
	}
}
