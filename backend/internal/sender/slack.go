package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SlackSender struct {
	webhookURL string
	client     *http.Client
}

func NewSlack(webhookURL string, client *http.Client) *SlackSender {
	if client == nil {
		client = http.DefaultClient
	}
	return &SlackSender{webhookURL: webhookURL, client: client}
}

func (s *SlackSender) Type() string {
	return "slack"
}

func (s *SlackSender) Send(ctx context.Context, to, subject, body string) error {
	payload := map[string]string{
		"channel": to,
		"text":    fmt.Sprintf("*%s*\n\n%s", subject, body),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
