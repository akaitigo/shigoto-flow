package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

type GitHubSource struct {
	client *http.Client
}

func NewGitHub(client *http.Client) *GitHubSource {
	if client == nil {
		client = http.DefaultClient
	}
	return &GitHubSource{client: client}
}

func (g *GitHubSource) Provider() model.Provider {
	return model.ProviderGitHub
}

func (g *GitHubSource) Collect(ctx context.Context, accessToken string, date time.Time) ([]model.Activity, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	apiURL := "https://api.github.com/user/events?per_page=100"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch github events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(body))
	}

	var events []struct {
		Type string `json:"type"`
		Repo struct {
			Name string `json:"name"`
		} `json:"repo"`
		Payload json.RawMessage `json:"payload"`
		Created string          `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode github response: %w", err)
	}

	now := time.Now()
	var activities []model.Activity
	for _, event := range events {
		ts, _ := time.Parse(time.RFC3339, event.Created)
		if ts.Before(startOfDay) {
			continue
		}
		if ts.After(startOfDay.Add(24 * time.Hour)) {
			continue
		}

		title := formatGitHubEventTitle(event.Type, event.Repo.Name)

		activities = append(activities, model.Activity{
			ID:        uuid.New().String(),
			Source:    model.ProviderGitHub,
			Title:     title,
			Body:      string(event.Payload),
			Timestamp: ts,
			Metadata:  event.Type,
			CreatedAt: now,
		})
	}

	return activities, nil
}

func formatGitHubEventTitle(eventType, repoName string) string {
	switch eventType {
	case "PushEvent":
		return fmt.Sprintf("%s にプッシュ", repoName)
	case "PullRequestEvent":
		return fmt.Sprintf("%s でPR操作", repoName)
	case "IssuesEvent":
		return fmt.Sprintf("%s でIssue操作", repoName)
	case "CreateEvent":
		return fmt.Sprintf("%s でブランチ/タグ作成", repoName)
	case "PullRequestReviewEvent":
		return fmt.Sprintf("%s でPRレビュー", repoName)
	default:
		return fmt.Sprintf("%s で%s", repoName, eventType)
	}
}
