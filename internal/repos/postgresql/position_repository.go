package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/jmoiron/sqlx"
)

// PositionRepository реализует работу с таблицей positions через sqlx.
type PositionRepository struct {
	db *sqlx.DB
}

// NewPositionRepository создаёт новый PositionRepository.
func NewPositionRepository(db *sqlx.DB) *PositionRepository {
	return &PositionRepository{db: db}
}

// GetByUserMarketOutcome возвращает позицию по комбинации user+market+outcome.
func (r *PositionRepository) GetByUserMarketOutcome(ctx context.Context, userID, marketID, outcome string) (*domain.Position, error) {
	var position domain.Position

	query := `SELECT id, user_id, market_id, outcome, quantity, avg_cost, created_at, updated_at
			  FROM positions
			  WHERE user_id = $1 AND market_id = $2 AND outcome = $3`

	err := r.db.GetContext(ctx, &position, query, userID, marketID, outcome)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	return &position, nil
}

// GetByUserID возвращает все позиции пользователя.
func (r *PositionRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Position, error) {
	var positions []domain.Position

	query := `SELECT id, user_id, market_id, outcome, quantity, avg_cost, created_at, updated_at
			  FROM positions
			  WHERE user_id = $1 AND quantity > 0
			  ORDER BY updated_at DESC`

	err := r.db.SelectContext(ctx, &positions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	if positions == nil {
		positions = []domain.Position{}
	}

	return positions, nil
}

// GetByMarketID возвращает все позиции по рынку.
func (r *PositionRepository) GetByMarketID(ctx context.Context, marketID string) ([]domain.Position, error) {
	var positions []domain.Position

	query := `SELECT id, user_id, market_id, outcome, quantity, avg_cost, created_at, updated_at
			  FROM positions
			  WHERE market_id = $1 AND quantity > 0`

	err := r.db.SelectContext(ctx, &positions, query, marketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions by market: %w", err)
	}

	if positions == nil {
		positions = []domain.Position{}
	}

	return positions, nil
}

// Upsert создаёт или обновляет позицию (INSERT ... ON CONFLICT UPDATE).
func (r *PositionRepository) Upsert(ctx context.Context, position *domain.Position) error {
	query := `INSERT INTO positions (id, user_id, market_id, outcome, quantity, avg_cost)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  ON CONFLICT (user_id, market_id, outcome)
			  DO UPDATE SET quantity = $5, avg_cost = $6`

	_, err := r.db.ExecContext(ctx, query,
		position.ID, position.UserID, position.MarketID,
		position.Outcome, position.Quantity, position.AvgCost,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert position: %w", err)
	}

	return nil
}
