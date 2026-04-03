package collector

import (
	"context"
	"testing"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

type mockSource struct {
	provider   model.Provider
	activities []model.Activity
	err        error
}

func (m *mockSource) Provider() model.Provider {
	return m.provider
}

func (m *mockSource) Collect(_ context.Context, _ string, _ time.Time) ([]model.Activity, error) {
	return m.activities, m.err
}

func TestCollector_CollectAll(t *testing.T) {
	tests := []struct {
		name      string
		sources   []Source
		tokens    map[model.Provider]string
		wantCount int
		wantErr   bool
	}{
		{
			name: "collects from all sources with tokens",
			sources: []Source{
				&mockSource{
					provider: model.ProviderGoogle,
					activities: []model.Activity{
						{ID: "1", Title: "Meeting"},
					},
				},
				&mockSource{
					provider: model.ProviderSlack,
					activities: []model.Activity{
						{ID: "2", Title: "Message"},
						{ID: "3", Title: "Reply"},
					},
				},
			},
			tokens: map[model.Provider]string{
				model.ProviderGoogle: "token1",
				model.ProviderSlack:  "token2",
			},
			wantCount: 3,
		},
		{
			name: "skips sources without tokens",
			sources: []Source{
				&mockSource{
					provider: model.ProviderGoogle,
					activities: []model.Activity{
						{ID: "1", Title: "Meeting"},
					},
				},
				&mockSource{
					provider: model.ProviderSlack,
					activities: []model.Activity{
						{ID: "2", Title: "Message"},
					},
				},
			},
			tokens: map[model.Provider]string{
				model.ProviderGoogle: "token1",
			},
			wantCount: 1,
		},
		{
			name:      "no sources returns empty",
			sources:   []Source{},
			tokens:    map[model.Provider]string{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.sources...)
			activities, err := c.CollectAll(context.Background(), tt.tokens, time.Now())

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(activities) != tt.wantCount {
				t.Errorf("expected %d activities, got %d", tt.wantCount, len(activities))
			}
		})
	}
}
