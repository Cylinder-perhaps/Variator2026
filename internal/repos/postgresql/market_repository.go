package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
)

type PostgresMarketRepository struct {
	db *sql.DB
}

func NewPostgresMarketRepository(db *sql.DB) *PostgresMarketRepository {
	return &PostgresMarketRepository{db: db}
}

func (r *PostgresMarketRepository) Create(ctx context.Context, market *domain.Market) error {
	query := `INSERT INTO markets (id, title, description, category, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, now(), now())`
	_, err := r.db.ExecContext(ctx, query, market.ID, market.Title, market.Description, market.Category, string(market.Status))
	if err != nil {
		return fmt.Errorf("failed to create market: %w", err)
	}
	return nil
}

func (r *PostgresMarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `SELECT id, title, description, category, status, created_at, updated_at
			  FROM markets 
			  WHERE id = $1`
	market := &domain.Market{}
	var status string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&market.ID,
		&market.Title,
		&market.Description,
		&market.Category,
		&status,
		&market.CreatedAt,
		&market.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("market not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get market by ID: %w", err)
	}
	market.Status = domain.MarketStatus(status)
	return market, nil
}

func (r *PostgresMarketRepository) GetAll(ctx context.Context) ([]domain.Market, error) {
	query := `SELECT id, title, description, category, status, created_at, updated_at
			  FROM markets`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all markets: %w", err)
	}
	defer rows.Close()

	var markets []domain.Market
	for rows.Next() {
		var market domain.Market
		var status string
		err := rows.Scan(
			&market.ID,
			&market.Title,
			&market.Description,
			&market.Category,
			&status,
			&market.CreatedAt,
			&market.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan market: %w", err)
		}
		market.Status = domain.MarketStatus(status)
		markets = append(markets, market)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over markets: %w", err)
	}

	return markets, nil
}

func (r *PostgresMarketRepository) Update(ctx context.Context, market *domain.Market) error {
	query := `UPDATE markets
			  SET title = $1, description = $2, category = $3, status = $4, updated_at = now()
			  WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, market.Title, market.Description, market.Category, string(market.Status), market.ID)
	if err != nil {
		return fmt.Errorf("failed to update market: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("market not found")
	}
	return nil
}
