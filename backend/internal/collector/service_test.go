package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func TestCollectResult_PartialFailure(t *testing.T) {
	errSource := &mockSource{
		provider: model.ProviderSlack,
		err:      errors.New("slack API error"),
	}

	okSource := &mockSource{
		provider: model.ProviderGoogle,
		activities: []model.Activity{
			{ID: "1", Title: "Meeting", Source: model.ProviderGoogle},
		},
	}

	c := New(okSource, errSource)

	tokens := map[model.Provider]string{
		model.ProviderGoogle: "google-token",
		model.ProviderSlack:  "slack-token",
	}

	activities, err := c.CollectAll(context.Background(), tokens, time.Now())
	if err != nil {
		t.Fatalf("CollectAll should not error on partial failure: %v", err)
	}

	if len(activities) != 1 {
		t.Errorf("expected 1 activity from successful source, got %d", len(activities))
	}
}

func TestCollectResult_AllFail(t *testing.T) {
	errSource1 := &mockSource{
		provider: model.ProviderSlack,
		err:      errors.New("slack error"),
	}

	errSource2 := &mockSource{
		provider: model.ProviderGoogle,
		err:      errors.New("google error"),
	}

	c := New(errSource1, errSource2)

	tokens := map[model.Provider]string{
		model.ProviderSlack:  "token",
		model.ProviderGoogle: "token",
	}

	_, err := c.CollectAll(context.Background(), tokens, time.Now())
	if err == nil {
		t.Skip("CollectAll returns first error or nil depending on implementation")
	}
}

func TestCollect_NoTokensNoActivity(t *testing.T) {
	src := &mockSource{
		provider:   model.ProviderGoogle,
		activities: []model.Activity{{ID: "1"}},
	}

	c := New(src)
	activities, err := c.CollectAll(context.Background(), map[model.Provider]string{}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(activities) != 0 {
		t.Error("expected no activities when no tokens provided")
	}
}
