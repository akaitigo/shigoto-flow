package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/google/uuid"
)

type SlackSource struct {
	client *http.Client
}

func NewSlack(client *http.Client) *SlackSource {
	if client == nil {
		client = http.DefaultClient
	}
	return &SlackSource{client: client}
}

func (s *SlackSource) Provider() model.Provider {
	return model.ProviderSlack
}

func (s *SlackSource) Collect(ctx context.Context, accessToken string, date time.Time) ([]model.Activity, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	apiURL := fmt.Sprintf(
		"https://slack.com/api/search.messages?query=from:me after:%s before:%s",
		startOfDay.Format("2006-01-02"),
		endOfDay.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slack messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("slack API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		OK       bool `json:"ok"`
		Messages struct {
			Matches []struct {
				Text    string `json:"text"`
				Channel struct {
					Name string `json:"name"`
				} `json:"channel"`
				TS string `json:"ts"`
			} `json:"matches"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode slack response: %w", err)
	}

	now := time.Now()
	var activities []model.Activity
	for _, msg := range result.Messages.Matches {
		tsFloat, _ := strconv.ParseFloat(msg.TS, 64)
		ts := time.Unix(int64(tsFloat), 0)

		activities = append(activities, model.Activity{
			ID:        uuid.New().String(),
			Source:    model.ProviderSlack,
			Title:     fmt.Sprintf("#%s での投稿", msg.Channel.Name),
			Body:      msg.Text,
			Timestamp: ts,
			CreatedAt: now,
		})
	}

	return activities, nil
}
