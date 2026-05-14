package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/jmoiron/sqlx"
)

// OrderRepository реализует работу с таблицей orders через sqlx.
type OrderRepository struct {
	db *sqlx.DB
}

// NewOrderRepository создаёт новый OrderRepository.
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create создаёт новый ордер.
func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `INSERT INTO orders (id, user_id, market_id, outcome, quantity, price, amount_paid, status)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			  RETURNING created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		order.ID, order.UserID, order.MarketID, order.Outcome,
		order.Quantity, order.Price, order.AmountPaid, string(order.Status),
	).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID возвращает ордер по ID.
func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var order domain.Order

	query := `SELECT id, user_id, market_id, outcome, quantity, price, amount_paid, status, created_at, updated_at
			  FROM orders
			  WHERE id = $1`

	err := r.db.GetContext(ctx, &order, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	return &order, nil
}

// List возвращает список ордеров пользователя с фильтрацией и пагинацией.
func (r *OrderRepository) List(ctx context.Context, userID string, filter domain.OrderFilter) ([]domain.Order, int, error) {
	conditions := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Считаем общее кол-во.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where)

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit

	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, market_id, outcome, quantity, price, amount_paid, status, created_at, updated_at
		FROM orders
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	var orders []domain.Order
	err = r.db.SelectContext(ctx, &orders, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	if orders == nil {
		orders = []domain.Order{}
	}

	return orders, total, nil
}

// UpdateStatus обновляет статус ордера.
func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetPendingByMarket возвращает все PENDING-ордера для данного рынка.
func (r *OrderRepository) GetPendingByMarket(ctx context.Context, marketID string) ([]domain.Order, error) {
	var orders []domain.Order

	query := `SELECT id, user_id, market_id, outcome, quantity, price, amount_paid, status, created_at, updated_at
			  FROM orders
			  WHERE market_id = $1 AND status = 'PENDING'
			  ORDER BY created_at ASC`

	err := r.db.SelectContext(ctx, &orders, query, marketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}

	if orders == nil {
		orders = []domain.Order{}
	}

	return orders, nil
}
