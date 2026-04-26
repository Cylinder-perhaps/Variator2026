package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
)

type PostgresTradeRepository struct {
	db *sql.DB
}

func NewPostgresTradeRepository(db *sql.DB) *PostgresTradeRepository {
	return &PostgresTradeRepository{db: db}
}

func (r *PostgresTradeRepository) Create(ctx context.Context, trade *domain.Trade) error {
	query := `INSERT INTO trades (id, order_id, user_id, market_id, outcome, quantity, price, pnl, executed_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query, trade.ID, trade.OrderID, trade.UserID, trade.MarketID, trade.Outcome, trade.Quantity, trade.Price, trade.PNL, trade.ExecutedAt)
	if err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}
	return nil
}

func (r *PostgresTradeRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Trade, error) {
	query := `SELECT id, order_id, user_id, market_id, outcome, quantity, price, pnl, executed_at
			  FROM trades 
			  WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by user ID: %w", err)
	}
	defer rows.Close()

	var trades []domain.Trade
	for rows.Next() {
		var trade domain.Trade
		err := rows.Scan(
			&trade.ID,
			&trade.OrderID,
			&trade.UserID,
			&trade.MarketID,
			&trade.Outcome,
			&trade.Quantity,
			&trade.Price,
			&trade.PNL,
			&trade.ExecutedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, trade)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over trades: %w", err)
	}
	return trades, nil
}

func (r *PostgresTradeRepository) GetByOrderID(ctx context.Context, orderID string) ([]domain.Trade, error) {
	query := `SELECT id, order_id, user_id, market_id, outcome, quantity, price, pnl, executed_at
			  FROM trades 
			  WHERE order_id = $1`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by order ID: %w", err)
	}
	defer rows.Close()

	var trades []domain.Trade
	for rows.Next() {
		var trade domain.Trade
		err := rows.Scan(
			&trade.ID,
			&trade.OrderID,
			&trade.UserID,
			&trade.MarketID,
			&trade.Outcome,
			&trade.Quantity,
			&trade.Price,
			&trade.PNL,
			&trade.ExecutedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, trade)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over trades: %w", err)
	}
	return trades, nil
}

func (r *PostgresTradeRepository) Update(ctx context.Context, trade *domain.Trade) error {
	query := `UPDATE trades
			  SET order_id = $1, user_id = $2, market_id = $3, outcome = $4, quantity = $5, price = $6, pnl = $7, executed_at = $8
			  WHERE id = $9`
	result, err := r.db.ExecContext(ctx, query, trade.OrderID, trade.UserID, trade.MarketID, trade.Outcome, trade.Quantity, trade.Price, trade.PNL, trade.ExecutedAt, trade.ID)
	if err != nil {
		return fmt.Errorf("failed to update trade: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("trade not found")
	}
	return nil
}

func (r *PostgresTradeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM trades WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete trade: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("trade not found")
	}
	return nil
}
