package postgresql

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/jmoiron/sqlx"
)

// TradeRepository реализует работу с таблицей trades через sqlx.
type TradeRepository struct {
	db *sqlx.DB
}

// NewTradeRepository создаёт новый TradeRepository.
func NewTradeRepository(db *sqlx.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

// Create создаёт новую сделку.
func (r *TradeRepository) Create(ctx context.Context, trade *domain.Trade) error {
	query := `INSERT INTO trades (id, order_id, user_id, market_id, outcome, quantity, price, pnl, executed_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		trade.ID, trade.OrderID, trade.UserID, trade.MarketID,
		trade.Outcome, trade.Quantity, trade.Price, trade.PNL, trade.ExecutedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}

	return nil
}

// List возвращает список сделок пользователя с фильтрацией и пагинацией.
func (r *TradeRepository) List(ctx context.Context, userID string, filter domain.TradeFilter) ([]domain.Trade, int, error) {
	conditions := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.MarketID != nil && *filter.MarketID != "" {
		conditions = append(conditions, fmt.Sprintf("market_id = $%d", argIdx))
		args = append(args, *filter.MarketID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Считаем общее кол-во.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM trades %s", where)

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count trades: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit

	dataQuery := fmt.Sprintf(`
		SELECT id, order_id, user_id, market_id, outcome, quantity, price, pnl, executed_at
		FROM trades
		%s
		ORDER BY executed_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	var trades []domain.Trade
	err = r.db.SelectContext(ctx, &trades, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list trades: %w", err)
	}

	if trades == nil {
		trades = []domain.Trade{}
	}

	return trades, total, nil
}
