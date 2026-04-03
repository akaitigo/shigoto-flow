package collector

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/akaitigo/shigoto-flow/backend/internal/auth"
	"github.com/akaitigo/shigoto-flow/backend/internal/model"
	"github.com/akaitigo/shigoto-flow/backend/internal/repository"
)

type Service struct {
	repo      *repository.Repository
	collector *Collector
	encryptor *auth.TokenEncryptor
}

func NewService(repo *repository.Repository, collector *Collector, encryptor *auth.TokenEncryptor) *Service {
	return &Service{
		repo:      repo,
		collector: collector,
		encryptor: encryptor,
	}
}

type CollectResult struct {
	Collected int
	Errors    []SourceError
}

type SourceError struct {
	Provider model.Provider
	Err      error
}

func (s *Service) CollectForUser(ctx context.Context, userID string, date time.Time) (*CollectResult, error) {
	dataSources, err := s.repo.ListDataSourcesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list data sources: %w", err)
	}

	if len(dataSources) == 0 {
		return &CollectResult{Collected: 0}, nil
	}

	tokens := make(map[model.Provider]string)
	for _, ds := range dataSources {
		decrypted, err := s.encryptor.Decrypt(ds.AccessToken)
		if err != nil {
			slog.Warn("failed to decrypt token, skipping",
				"provider", ds.Provider,
				"error", err,
			)
			continue
		}
		tokens[ds.Provider] = decrypted
	}

	type sourceResult struct {
		activities []model.Activity
		provider   model.Provider
		err        error
	}

	results := make(chan sourceResult, len(s.collector.sources))
	g, gCtx := errgroup.WithContext(ctx)

	for _, src := range s.collector.sources {
		token, ok := tokens[src.Provider()]
		if !ok {
			continue
		}

		g.Go(func() error {
			activities, err := src.Collect(gCtx, token, date)
			results <- sourceResult{
				activities: activities,
				provider:   src.Provider(),
				err:        err,
			}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(results)
	}()

	result := &CollectResult{}
	for sr := range results {
		if sr.err != nil {
			slog.Warn("collection failed for source",
				"provider", sr.provider,
				"error", sr.err,
			)
			result.Errors = append(result.Errors, SourceError{
				Provider: sr.provider,
				Err:      sr.err,
			})
			continue
		}

		for i := range sr.activities {
			sr.activities[i].UserID = userID
			if err := s.repo.CreateActivity(ctx, &sr.activities[i]); err != nil {
				slog.Error("failed to save activity",
					"provider", sr.provider,
					"error", err,
				)
				continue
			}
			result.Collected++
		}
	}

	return result, nil
}
