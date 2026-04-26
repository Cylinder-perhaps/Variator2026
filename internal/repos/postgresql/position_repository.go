package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
)

type PostgresPositionRepository struct {
	db *sql.DB
}

func NewPostgresPositionRepository(db *sql.DB) *PostgresPositionRepository {
	return &PostgresPositionRepository{db: db}
}

func (r *PostgresPositionRepository) Create(ctx context.Context, position *domain.Position) error {
	query := `INSERT INTO positions (id, user_id, market_id, outcome, quantity, avg_cost, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, now(), now())`
	_, err := r.db.ExecContext(ctx, query, position.ID, position.UserID, position.MarketID, position.Outcome, position.Quantity, position.AvgCost)
	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}
	return nil
}

func (r *PostgresPositionRepository) GetByUserMarketOutcome(ctx context.Context, userID, marketID, outcome string) (*domain.Position, error) {
	query := `SELECT id, user_id, market_id, outcome, quantity, avg_cost, created_at, updated_at
			  FROM positions 
			  WHERE user_id = $1 AND market_id = $2 AND outcome = $3`
	position := &domain.Position{}
	err := r.db.QueryRowContext(ctx, query, userID, marketID, outcome).Scan(
		&position.ID,
		&position.UserID,
		&position.MarketID,
		&position.Outcome,
		&position.Quantity,
		&position.AvgCost,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("position not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}
	return position, nil
}

// func (r *PostgresPositionRepository) GetByID(ctx context.Context, id string) (*domain.Position, error) {
// 	query := `SELECT id, user_id, market_id, outcome, quantity, avg_cost, created_at, updated_at
// 			  FROM positions 
// 			  WHERE id = $1`
// 	position := &domain.Position{}
// 	err := r.db.QueryRowContext(ctx, query, id).Scan(
// 		&position.ID,
// 		&position.UserID,
// 		&position.MarketID,
// 		&position.Outcome,
// 		&position.Quantity,
// 		&position.AvgCost,
// 		&position.CreatedAt,
// 		&position.UpdatedAt,
// 	)

// 	if err == sql.ErrNoRows {
// 		return nil, fmt.Errorf("position not found")
// 	}
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get position: %w", err)
// 	}
// 	return position, nil
// }

func (r *PostgresPositionRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Position, error) {
	query := `SELECT id, user_id, market_id, outcome, quantity, avg_cost, created_at, updated_at
			  FROM positions 
			  WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil{
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	defer rows.Close()

	var positions []domain.Position
	for rows.Next() {
		var position domain.Position
		err := rows.Scan(
			&position.ID,
			&position.UserID,
			&position.MarketID,
			&position.Outcome,
			&position.Quantity,
			&position.AvgCost,
			&position.CreatedAt,
			&position.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over positions: %w", err)

	}
	return positions, nil
}

func (r *PostgresPositionRepository) Update(ctx context.Context, position *domain.Position) error {
	query := `UPDATE positions
			  SET quantity = $1, avg_cost = $2, updated_at = now()
			  WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, position.Quantity, position.AvgCost, position.ID)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("position not found")
	}
	return nil
}

func (r *PostgresPositionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM positions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("position not found")
	}
	return nil
}
