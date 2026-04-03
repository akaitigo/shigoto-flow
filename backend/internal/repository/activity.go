package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (r *Repository) CreateActivity(ctx context.Context, activity *model.Activity) error {
	query := `
		INSERT INTO activities (id, user_id, source, title, body, timestamp, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		activity.ID, activity.UserID, activity.Source,
		activity.Title, activity.Body, activity.Timestamp,
		activity.Metadata, activity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	return nil
}

func (r *Repository) ListActivitiesByUserAndDate(ctx context.Context, userID string, date time.Time) ([]model.Activity, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, user_id, source, title, body, timestamp, metadata, created_at
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3
		ORDER BY timestamp ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}
	defer rows.Close()

	var activities []model.Activity
	for rows.Next() {
		var a model.Activity
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Source,
			&a.Title, &a.Body, &a.Timestamp,
			&a.Metadata, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, a)
	}
	return activities, rows.Err()
}

func (r *Repository) ListActivitiesByUserAndRange(ctx context.Context, userID string, start, end time.Time) ([]model.Activity, error) {
	query := `
		SELECT id, user_id, source, title, body, timestamp, metadata, created_at
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3
		ORDER BY timestamp ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities by range: %w", err)
	}
	defer rows.Close()

	var activities []model.Activity
	for rows.Next() {
		var a model.Activity
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Source,
			&a.Title, &a.Body, &a.Timestamp,
			&a.Metadata, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, a)
	}
	return activities, rows.Err()
}
