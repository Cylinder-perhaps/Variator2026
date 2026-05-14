package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/jmoiron/sqlx"
)

// RefreshTokenRepository реализует работу с таблицей refresh_tokens через sqlx.
type RefreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository создаёт новый RefreshTokenRepository.
func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create создаёт новый refresh-токен.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
			  VALUES ($1, $2, $3, $4)
			  RETURNING created_at`

	err := r.db.QueryRowxContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash возвращает refresh-токен по хешу.
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken

	query := `SELECT id, user_id, token_hash, expires_at, revoked, created_at
			  FROM refresh_tokens
			  WHERE token_hash = $1`

	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

// RevokeByUserID отзывает все refresh-токены пользователя.
func (r *RefreshTokenRepository) RevokeByUserID(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke tokens by user ID: %w", err)
	}

	return nil
}

// RevokeByTokenHash отзывает конкретный refresh-токен.
func (r *RefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`

	_, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}
