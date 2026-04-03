package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

type GoogleCalendarSource struct {
	client *http.Client
}

func NewGoogleCalendar(client *http.Client) *GoogleCalendarSource {
	if client == nil {
		client = http.DefaultClient
	}
	return &GoogleCalendarSource{client: client}
}

func (g *GoogleCalendarSource) Provider() model.Provider {
	return model.ProviderGoogle
}

func (g *GoogleCalendarSource) Collect(ctx context.Context, accessToken string, date time.Time) ([]model.Activity, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	params := url.Values{}
	params.Set("timeMin", startOfDay.Format(time.RFC3339))
	params.Set("timeMax", endOfDay.Format(time.RFC3339))
	params.Set("singleEvents", "true")
	params.Set("orderBy", "startTime")

	apiURL := "https://www.googleapis.com/calendar/v3/calendars/primary/events?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("calendar API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Summary     string `json:"summary"`
			Description string `json:"description"`
			Start       struct {
				DateTime string `json:"dateTime"`
			} `json:"start"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode calendar response: %w", err)
	}

	now := time.Now()
	var activities []model.Activity
	for _, item := range result.Items {
		ts, _ := time.Parse(time.RFC3339, item.Start.DateTime)
		if ts.IsZero() {
			ts = startOfDay
		}

		activities = append(activities, model.Activity{
			ID:        uuid.New().String(),
			Source:    model.ProviderGoogle,
			Title:     item.Summary,
			Body:      item.Description,
			Timestamp: ts,
			CreatedAt: now,
		})
	}

	return activities, nil
}
