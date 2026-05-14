package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/jmoiron/sqlx"
)

// BalanceRepository реализует работу с таблицей balances через sqlx.
type BalanceRepository struct {
	db *sqlx.DB
}

// NewBalanceRepository создаёт новый BalanceRepository.
func NewBalanceRepository(db *sqlx.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// GetByUserID возвращает баланс пользователя.
func (r *BalanceRepository) GetByUserID(ctx context.Context, userID string) (*domain.Balance, error) {
	var balance domain.Balance

	query := `SELECT id, user_id, total, available, blocked_in_orders, updated_at
			  FROM balances
			  WHERE user_id = $1`

	err := r.db.GetContext(ctx, &balance, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &balance, nil
}

// UpdateBalance атомарно обновляет доступный и заблокированный баланс.
// availableDelta и blockedDelta могут быть отрицательными для списания.
func (r *BalanceRepository) UpdateBalance(ctx context.Context, userID string, availableDelta, blockedDelta float64) error {
	query := `UPDATE balances
			  SET available = available + $1,
			      blocked_in_orders = blocked_in_orders + $2,
			      total = (available + $1) + (blocked_in_orders + $2)
			  WHERE user_id = $3
			    AND (available + $1) >= 0
			    AND (blocked_in_orders + $2) >= 0`

	result, err := r.db.ExecContext(ctx, query, availableDelta, blockedDelta, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrInsufficientBalance
	}

	return nil
}
