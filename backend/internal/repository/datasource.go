package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (r *Repository) UpsertDataSource(ctx context.Context, ds *model.DataSource) error {
	query := `
		INSERT INTO data_sources (id, user_id, provider, access_token, refresh_token, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, provider) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		ds.ID, ds.UserID, ds.Provider,
		ds.AccessToken, ds.RefreshToken, ds.ExpiresAt,
		ds.CreatedAt, ds.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert data source: %w", err)
	}
	return nil
}

func (r *Repository) GetDataSource(ctx context.Context, userID string, provider model.Provider) (*model.DataSource, error) {
	query := `
		SELECT id, user_id, provider, access_token, refresh_token, expires_at, created_at, updated_at
		FROM data_sources WHERE user_id = $1 AND provider = $2
	`
	var ds model.DataSource
	err := r.db.QueryRowContext(ctx, query, userID, provider).Scan(
		&ds.ID, &ds.UserID, &ds.Provider,
		&ds.AccessToken, &ds.RefreshToken, &ds.ExpiresAt,
		&ds.CreatedAt, &ds.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get data source: %w", err)
	}
	return &ds, nil
}

func (r *Repository) ListDataSourcesByUser(ctx context.Context, userID string) ([]model.DataSource, error) {
	query := `
		SELECT id, user_id, provider, access_token, refresh_token, expires_at, created_at, updated_at
		FROM data_sources WHERE user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list data sources: %w", err)
	}
	defer rows.Close()

	var sources []model.DataSource
	for rows.Next() {
		var ds model.DataSource
		if err := rows.Scan(
			&ds.ID, &ds.UserID, &ds.Provider,
			&ds.AccessToken, &ds.RefreshToken, &ds.ExpiresAt,
			&ds.CreatedAt, &ds.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan data source: %w", err)
		}
		sources = append(sources, ds)
	}
	return sources, rows.Err()
}

func (r *Repository) DeleteDataSource(ctx context.Context, userID string, provider model.Provider) error {
	query := `DELETE FROM data_sources WHERE user_id = $1 AND provider = $2`
	_, err := r.db.ExecContext(ctx, query, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to delete data source: %w", err)
	}
	return nil
}
