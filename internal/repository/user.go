package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/mrshabel/chat/internal/model"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := `
        INSERT INTO users(username)
        VALUES ($1)
		RETURNING id, username, created_at, updated_at
    `
	if err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, created_at, updated_at 
		FROM users 
		WHERE username = $1
		`
	var user model.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, created_at, updated_at 
		FROM users 
		WHERE id = $1
		`
	var user model.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
