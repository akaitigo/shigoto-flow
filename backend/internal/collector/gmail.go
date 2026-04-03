package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/google/uuid"
)

type GmailSource struct {
	client *http.Client
}

func NewGmail(client *http.Client) *GmailSource {
	if client == nil {
		client = http.DefaultClient
	}
	return &GmailSource{client: client}
}

func (g *GmailSource) Provider() model.Provider {
	return model.ProviderGmail
}

func (g *GmailSource) Collect(ctx context.Context, accessToken string, date time.Time) ([]model.Activity, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	query := fmt.Sprintf("after:%s before:%s",
		startOfDay.Format("2006/01/02"),
		startOfDay.Add(24*time.Hour).Format("2006/01/02"),
	)

	params := url.Values{}
	params.Set("q", query)
	params.Set("maxResults", "50")

	apiURL := "https://www.googleapis.com/gmail/v1/users/me/messages?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gmail messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gmail API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode gmail response: %w", err)
	}

	now := time.Now()
	var activities []model.Activity
	for _, msg := range result.Messages {
		detail, err := g.getMessageDetail(ctx, accessToken, msg.ID)
		if err != nil {
			continue
		}

		activities = append(activities, model.Activity{
			ID:        uuid.New().String(),
			Source:    model.ProviderGmail,
			Title:     detail.subject,
			Body:      detail.snippet,
			Timestamp: detail.date,
			Metadata:  detail.from,
			CreatedAt: now,
		})
	}

	return activities, nil
}

type messageDetail struct {
	subject string
	from    string
	snippet string
	date    time.Time
}

func (g *GmailSource) getMessageDetail(ctx context.Context, accessToken, messageID string) (*messageDetail, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/gmail/v1/users/me/messages/%s?format=metadata&metadataHeaders=Subject&metadataHeaders=From", messageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gmail API returned %d: %s", resp.StatusCode, string(body))
	}

	var msg struct {
		Snippet       string `json:"snippet"`
		InternalDate  string `json:"internalDate"`
		Payload       struct {
			Headers []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"headers"`
		} `json:"payload"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message detail: %w", err)
	}

	detail := &messageDetail{
		snippet: msg.Snippet,
		date:    time.Now(),
	}

	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "Subject":
			detail.subject = header.Value
		case "From":
			detail.from = header.Value
		}
	}

	return detail, nil
}
