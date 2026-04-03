package summary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Summarizer struct {
	apiKey string
	client *http.Client
}

func New(apiKey string, client *http.Client) *Summarizer {
	if client == nil {
		client = http.DefaultClient
	}
	return &Summarizer{apiKey: apiKey, client: client}
}

type SummarizeInput struct {
	Reports    []string
	ReportType string
	Period     string
}

func (s *Summarizer) Summarize(ctx context.Context, input SummarizeInput) (string, error) {
	prompt := buildPrompt(input)

	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 2048,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Claude API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Claude response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude API")
	}

	return result.Content[0].Text, nil
}

func buildPrompt(input SummarizeInput) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("以下の%sデータから%sを生成してください。\n\n", input.ReportType, input.Period))
	buf.WriteString("要約のルール:\n")
	buf.WriteString("- 重要な成果とアクションを箇条書きで整理\n")
	buf.WriteString("- 課題や懸念事項を明記\n")
	buf.WriteString("- 次期の予定・目標を提案\n")
	buf.WriteString("- 簡潔で読みやすい日本語で記述\n\n")

	for i, r := range input.Reports {
		buf.WriteString(fmt.Sprintf("--- レポート %d ---\n%s\n\n", i+1, r))
	}

	return buf.String()
}
