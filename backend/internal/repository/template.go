package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (r *Repository) CreateTemplate(ctx context.Context, tmpl *model.Template) error {
	sectionsJSON, err := json.Marshal(tmpl.Sections)
	if err != nil {
		return fmt.Errorf("failed to marshal sections: %w", err)
	}

	query := `
		INSERT INTO templates (id, user_id, name, type, sections, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.db.ExecContext(ctx, query,
		tmpl.ID, tmpl.UserID, tmpl.Name, tmpl.Type,
		sectionsJSON, tmpl.IsDefault, tmpl.CreatedAt, tmpl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	return nil
}

func (r *Repository) GetDefaultTemplate(ctx context.Context, userID string, reportType model.ReportType) (*model.Template, error) {
	query := `
		SELECT id, user_id, name, type, sections, is_default, created_at, updated_at
		FROM templates WHERE user_id = $1 AND type = $2 AND is_default = true
		LIMIT 1
	`
	var tmpl model.Template
	var sectionsJSON []byte
	err := r.db.QueryRowContext(ctx, query, userID, reportType).Scan(
		&tmpl.ID, &tmpl.UserID, &tmpl.Name, &tmpl.Type,
		&sectionsJSON, &tmpl.IsDefault, &tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default template: %w", err)
	}

	if err := json.Unmarshal(sectionsJSON, &tmpl.Sections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sections: %w", err)
	}
	return &tmpl, nil
}

func (r *Repository) ListTemplatesByUser(ctx context.Context, userID string) ([]model.Template, error) {
	query := `
		SELECT id, user_id, name, type, sections, is_default, created_at, updated_at
		FROM templates WHERE user_id = $1 ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []model.Template
	for rows.Next() {
		var tmpl model.Template
		var sectionsJSON []byte
		if err := rows.Scan(
			&tmpl.ID, &tmpl.UserID, &tmpl.Name, &tmpl.Type,
			&sectionsJSON, &tmpl.IsDefault, &tmpl.CreatedAt, &tmpl.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		if err := json.Unmarshal(sectionsJSON, &tmpl.Sections); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sections: %w", err)
		}
		templates = append(templates, tmpl)
	}
	return templates, rows.Err()
}
