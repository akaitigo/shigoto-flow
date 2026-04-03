package collector

import (
	"context"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

type Source interface {
	Provider() model.Provider
	Collect(ctx context.Context, accessToken string, date time.Time) ([]model.Activity, error)
}

type Collector struct {
	sources []Source
}

func New(sources ...Source) *Collector {
	return &Collector{sources: sources}
}

func (c *Collector) CollectAll(ctx context.Context, tokens map[model.Provider]string, date time.Time) ([]model.Activity, error) {
	var allActivities []model.Activity

	for _, src := range c.sources {
		token, ok := tokens[src.Provider()]
		if !ok {
			continue
		}

		activities, err := src.Collect(ctx, token, date)
		if err != nil {
			return nil, err
		}

		allActivities = append(allActivities, activities...)
	}

	return allActivities, nil
}
