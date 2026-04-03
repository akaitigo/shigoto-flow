package summary

import (
	"strings"
	"testing"
)

func TestBuildPrompt_WithConfig(t *testing.T) {
	input := SummarizeInput{
		Reports:    []string{"日報内容1", "日報内容2", "日報内容3"},
		ReportType: "日報",
		Period:     "週報",
	}

	prompt := buildPrompt(input)

	if !strings.Contains(prompt, "日報") {
		t.Error("expected prompt to contain report type")
	}
	if !strings.Contains(prompt, "週報") {
		t.Error("expected prompt to contain target period")
	}
	if strings.Count(prompt, "レポート") != 3 {
		t.Errorf("expected 3 report sections, got %d", strings.Count(prompt, "レポート"))
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
