package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/akaitigo/shigoto-flow/backend/internal/model"
)

func (r *Repository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, name, created_at, updated_at FROM users WHERE email = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET email = $1, name = $2, updated_at = $3 WHERE id = $4`
	user.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, user.Email, user.Name, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
