package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (r *Repository) CreateReport(ctx context.Context, report *model.Report) error {
	query := `
		INSERT INTO reports (id, user_id, type, template_id, content, date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.UserID, report.Type, report.TemplateID,
		report.Content, report.Date, report.Status,
		report.CreatedAt, report.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}
	return nil
}

func (r *Repository) GetReportByID(ctx context.Context, id string) (*model.Report, error) {
	query := `
		SELECT id, user_id, type, template_id, content, date, status, created_at, updated_at
		FROM reports WHERE id = $1
	`
	var report model.Report
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID, &report.UserID, &report.Type, &report.TemplateID,
		&report.Content, &report.Date, &report.Status,
		&report.CreatedAt, &report.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	return &report, nil
}

func (r *Repository) ListReportsByUser(ctx context.Context, userID string, reportType model.ReportType, limit, offset int) ([]model.Report, error) {
	query := `
		SELECT id, user_id, type, template_id, content, date, status, created_at, updated_at
		FROM reports WHERE user_id = $1 AND type = $2
		ORDER BY date DESC LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, userID, reportType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var reports []model.Report
	for rows.Next() {
		var report model.Report
		if err := rows.Scan(
			&report.ID, &report.UserID, &report.Type, &report.TemplateID,
			&report.Content, &report.Date, &report.Status,
			&report.CreatedAt, &report.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

// ListReportsByUserAndDateRange returns all reports of the given type whose
// date falls within [start, end). Unlike ListReportsByUser it is not capped by
// a fixed limit, so callers can retrieve an arbitrary historical window (e.g.
// a specific past week or month) rather than only the most recent reports.
func (r *Repository) ListReportsByUserAndDateRange(ctx context.Context, userID string, reportType model.ReportType, start, end time.Time) ([]model.Report, error) {
	query := `
		SELECT id, user_id, type, template_id, content, date, status, created_at, updated_at
		FROM reports
		WHERE user_id = $1 AND type = $2 AND date >= $3 AND date < $4
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, reportType, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports by date range: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var reports []model.Report
	for rows.Next() {
		var report model.Report
		if err := rows.Scan(
			&report.ID, &report.UserID, &report.Type, &report.TemplateID,
			&report.Content, &report.Date, &report.Status,
			&report.CreatedAt, &report.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

func (r *Repository) UpdateReportContent(ctx context.Context, id, content, status string) error {
	query := `UPDATE reports SET content = $1, status = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, content, status, id)
	if err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}
	return nil
}
