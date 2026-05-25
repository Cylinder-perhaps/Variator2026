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

// MarketRepository реализует работу с таблицей markets через sqlx.
type MarketRepository struct {
	db *sqlx.DB
}

// NewMarketRepository создаёт новый MarketRepository.
func NewMarketRepository(db *sqlx.DB) *MarketRepository {
	return &MarketRepository{db: db}
}

// Create создаёт новый рынок.
func (r *MarketRepository) Create(ctx context.Context, market *domain.Market) error {
	query := `INSERT INTO markets (id, title, description, outcomes, status, category, deadline, created_by, external_id, external_source)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			  RETURNING created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		market.ID, market.Title, market.Description,
		market.Outcomes, string(market.Status),
		market.Category, market.Deadline, market.CreatedBy,
		market.ExternalID, market.ExternalSource,
	).Scan(&market.CreatedAt, &market.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create market: %w", err)
	}

	return nil
}

// GetByID возвращает рынок по ID.
func (r *MarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	var market domain.Market

	query := `SELECT id, title, description, outcomes, status, category,
	                 deadline, resolved_outcome, evidence_url, created_by,
	                 external_id, external_source, created_at, updated_at
			  FROM markets
			  WHERE id = $1`

	err := r.db.GetContext(ctx, &market, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get market by ID: %w", err)
	}

	return &market, nil
}

// List возвращает список рынков с фильтрацией и пагинацией.
func (r *MarketRepository) List(ctx context.Context, filter domain.MarketFilter) ([]domain.Market, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Считаем общее кол-во.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM markets %s", where)

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count markets: %w", err)
	}

	// Определяем порядок сортировки.
	orderBy := "created_at DESC"
	switch filter.SortBy {
	case "deadline":
		orderBy = "deadline ASC"
	case "created_at":
		orderBy = "created_at DESC"
	}

	offset := (filter.Page - 1) * filter.Limit

	dataQuery := fmt.Sprintf(`
		SELECT id, title, description, outcomes, status, category,
		       deadline, resolved_outcome, evidence_url, created_by,
		       created_at, updated_at
		FROM markets
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	var markets []domain.Market
	err = r.db.SelectContext(ctx, &markets, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list markets: %w", err)
	}

	if markets == nil {
		markets = []domain.Market{}
	}

	return markets, total, nil
}

// Update обновляет данные рынка.
func (r *MarketRepository) Update(ctx context.Context, market *domain.Market) error {
	query := `UPDATE markets
			  SET title = $1, description = $2, outcomes = $3, status = $4,
			      category = $5, deadline = $6, resolved_outcome = $7, evidence_url = $8,
			      external_id = $9, external_source = $10, updated_at = now()
			  WHERE id = $11`

	result, err := r.db.ExecContext(ctx, query,
		market.Title, market.Description, market.Outcomes,
		string(market.Status), market.Category, market.Deadline,
		market.ResolvedOutcome, market.EvidenceURL,
		market.ExternalID, market.ExternalSource, market.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update market: %w", err)
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

// GetByExternalID возвращает рынок по внешнему ID и источнику.
func (r *MarketRepository) GetByExternalID(ctx context.Context, externalID, source string) (*domain.Market, error) {
	var market domain.Market

	query := `SELECT id, title, description, outcomes, status, category,
	                 deadline, resolved_outcome, evidence_url, created_by,
	                 external_id, external_source, created_at, updated_at
			  FROM markets
			  WHERE external_id = $1 AND external_source = $2`

	err := r.db.GetContext(ctx, &market, query, externalID, source)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get market by external ID: %w", err)
	}

	return &market, nil
}

// ListByExternalSource возвращает все активные рынки с указанным внешним источником.
func (r *MarketRepository) ListByExternalSource(ctx context.Context, source string) ([]domain.Market, error) {
	var markets []domain.Market

	query := `SELECT id, title, description, outcomes, status, category,
	                 deadline, resolved_outcome, evidence_url, created_by,
	                 external_id, external_source, created_at, updated_at
			  FROM markets
			  WHERE external_source = $1 AND status = 'ACTIVE'
			  ORDER BY created_at DESC`

	if err := r.db.SelectContext(ctx, &markets, query, source); err != nil {
		return nil, fmt.Errorf("failed to list markets by source: %w", err)
	}

	return markets, nil
}
