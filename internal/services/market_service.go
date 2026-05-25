package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
	"github.com/google/uuid"
)

// MarketService реализует бизнес-логику рынков.
type MarketService struct {
	markets   repos.MarketRepository
	orders    repos.OrderRepository
	balances  repos.BalanceRepository
	positions repos.PositionRepository
	trades    repos.TradeRepository
}

// NewMarketService создаёт новый MarketService.
func NewMarketService(
	markets repos.MarketRepository,
	orders repos.OrderRepository,
	balances repos.BalanceRepository,
	positions repos.PositionRepository,
	trades repos.TradeRepository,
) *MarketService {
	return &MarketService{
		markets:   markets,
		orders:    orders,
		balances:  balances,
		positions: positions,
		trades:    trades,
	}
}

// ListMarkets возвращает список рынков с пагинацией.
func (s *MarketService) ListMarkets(ctx context.Context, status *string, sortBy string, page, limit int) ([]domain.Market, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if sortBy == "" {
		sortBy = "created_at"
	}

	filter := domain.MarketFilter{
		Status: status,
		SortBy: sortBy,
		Page:   page,
		Limit:  limit,
	}

	markets, total, err := s.markets.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list markets: %w", err)
	}

	for i := range markets {
		pools, err := s.positions.GetPoolsByMarketID(ctx, markets[i].ID)
		if err == nil && len(pools) > 0 {
			markets[i].Pools = pools
		}
	}

	return markets, total, nil
}

// GetMarketDetails возвращает детали рынка по ID.
	market, err := s.markets.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pools, err := s.positions.GetPoolsByMarketID(ctx, market.ID)
	if err == nil && len(pools) > 0 {
		market.Pools = pools
	}

	return market, nil
}

// CreateMarket создаёт новый рынок (только admin).
func (s *MarketService) CreateMarket(ctx context.Context, title string, description *string, outcomes []string, deadline time.Time, category *string, createdBy string) (*domain.Market, error) {
	market := &domain.Market{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Outcomes:    domain.OutcomesJSON(outcomes),
		Status:      domain.MarketStatusActive,
		Category:    category,
		Deadline:    deadline,
		CreatedBy:   &createdBy,
	}

	if err := s.markets.Create(ctx, market); err != nil {
		return nil, fmt.Errorf("failed to create market: %w", err)
	}

	return market, nil
}

// ResolveMarket разрешает рынок, закрывает ордера и начисляет выигрыши.
func (s *MarketService) ResolveMarket(ctx context.Context, marketID, winningOutcome string, evidenceURL *string) (*domain.Market, float64, error) {
	market, err := s.markets.GetByID(ctx, marketID)
	if err != nil {
		return nil, 0, err
	}

	if market.Status != domain.MarketStatusActive && market.Status != domain.MarketStatusClosed {
		return nil, 0, domain.ErrConflict
	}

	// Проверяем, что исход валиден.
	validOutcome := false
	for _, o := range market.Outcomes {
		if o == winningOutcome {
			validOutcome = true
			break
		}
	}
	if !validOutcome {
		return nil, 0, fmt.Errorf("%w: invalid outcome '%s'", domain.ErrInvalidRequest, winningOutcome)
	}

	// Отменяем все PENDING ордера и возвращаем средства.
	pendingOrders, err := s.orders.GetPendingByMarket(ctx, marketID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get pending orders: %w", err)
	}

	for _, order := range pendingOrders {
		if err := s.orders.UpdateStatus(ctx, order.ID, domain.OrderStatusCancelled); err != nil {
			return nil, 0, fmt.Errorf("failed to cancel order %s: %w", order.ID, err)
		}
		// Возвращаем заблокированные средства.
		if err := s.balances.UpdateBalance(ctx, order.UserID, order.AmountPaid, -order.AmountPaid); err != nil {
			// Ошибка ErrInsufficientBalance здесь не должна возникать, но обрабатываем.
			if !errors.Is(err, domain.ErrInsufficientBalance) {
				return nil, 0, fmt.Errorf("failed to refund order %s: %w", order.ID, err)
			}
		}
	}

	// Обновляем статус рынка.
	market.Status = domain.MarketStatusResolved
	market.ResolvedOutcome = &winningOutcome
	market.EvidenceURL = evidenceURL

	if err := s.markets.Update(ctx, market); err != nil {
		return nil, 0, fmt.Errorf("failed to update market: %w", err)
	}

	// Начисление выигрышей по пулам (Parimutuel System)
	totalPayout := 0.0

	positions, err := s.positions.GetByMarketID(ctx, marketID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get positions for payout: %w", err)
	}

	totalPool := 0.0
	winningPool := 0.0

	for _, pos := range positions {
		investment := pos.Quantity * pos.AvgCost
		totalPool += investment
		if pos.Outcome == winningOutcome {
			winningPool += investment
		}
	}

	for _, pos := range positions {
		investment := pos.Quantity * pos.AvgCost

		var payout float64
		if winningPool == 0 {
			// Refund: никто не угадал, возвращаем деньги всем участникам
			payout = investment
		} else if pos.Outcome == winningOutcome {
			// Выплата = (Ставка пользователя / Пул победителей) * Общий Пул
			payout = (investment / winningPool) * totalPool
		} else {
			payout = 0
		}

		if payout > 0 {
			if err := s.balances.UpdateBalance(ctx, pos.UserID, payout, 0); err != nil {
				// В реальной системе здесь должна быть транзакционность
				return nil, 0, fmt.Errorf("failed to process payout for user %s: %w", pos.UserID, err)
			}
			totalPayout += payout
		}
	}

	return market, totalPayout, nil
}
