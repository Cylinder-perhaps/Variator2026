package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `INSERT INTO orders (id, user_id, market_id, outcome, quantity, price, amount_paid, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())`
	_, err := r.db.ExecContext(ctx, query, order.ID, order.UserID, order.MarketID, order.Outcome, order.Quantity, order.Price, order.AmountPaid, string(order.Status))
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `SELECT id, user_id, market_id, outcome, quantity, price, amount_paid, status, created_at, updated_at
			  FROM orders 
			  WHERE id = $1`
	order := &domain.Order{}
	var status string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.MarketID,
		&order.Outcome,
		&order.Quantity,
		&order.Price,
		&order.AmountPaid,
		&status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}
	order.Status = domain.OrderStatus(status)
	return order, nil
}

func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userId string) ([]domain.Order, error) {
	query := `SELECT id, user_id, market_id, outcome, quantity, price, amount_paid, status, created_at, updated_at
			  FROM orders
			  WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		var status string
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.MarketID,
			&order.Outcome,
			&order.Quantity,
			&order.Price,
			&order.AmountPaid,
			&status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		order.Status = domain.OrderStatus(status)
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `UPDATE orders
			  SET quantity = $1, price = $2, amount_paid = $3, status = $4, updated_at = now()
			  WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, order.Quantity, order.Price, order.AmountPaid, string(order.Status), order.ID)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}

func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM orders WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}
