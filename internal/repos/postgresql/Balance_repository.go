package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
)


type PostgresBalanceRepository struct {
	db *sql.DB
}


func NewPostgresBalanceRepository(db *sql.DB) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{db: db}
}

func (r *PostgresBalanceRepository) GetByUserID(ctx context.Context, userID string) (*domain.Balance, error) {
	query := `SELECT user_id, amount, currency, created_at, updated_at
			  FROM balances 
			  WHERE user_id = $1`
	
	balance := &domain.Balance{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.Amount,
		&balance.Currency,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("balance not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

func (r *PostgresBalanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	query := `UPDATE balances 
			  SET amount = $1, currency = $2, updated_at = now()
			  WHERE user_id = $3`
	result, err := r.db.ExecContext(ctx, query, balance.Amount, balance.Currency, balance.UserID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("balance not found")
	}
	return nil
}