package repository

import (
	"context"
	"errors"
	"time"

	"github.com/auction-studio/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, phone, location, rating, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	_, err := r.pool.Exec(ctx, query, u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.Phone, u.Location, u.Rating, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, phone, location, rating, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Phone, &u.Location, &u.Rating, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, phone, location, rating, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)
	`
	var u models.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Phone, &u.Location, &u.Rating, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens WHERE token_hash = $1 AND expires_at > NOW()
	`
	var t models.RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *UserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}
