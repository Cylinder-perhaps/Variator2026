package services

import (
	"context"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
)

// PositionService реализует бизнес-логику позиций.
type PositionService struct {
	positions repos.PositionRepository
	markets   repos.MarketRepository
}

// NewPositionService создаёт новый PositionService.
func NewPositionService(
	positions repos.PositionRepository,
	markets repos.MarketRepository,
) *PositionService {
	return &PositionService{positions: positions, markets: markets}
}

// PositionItem — позиция с названием рынка.
type PositionItem struct {
	domain.Position
	MarketTitle  string
	CurrentValue float64
	PNL          float64
}

// PositionsResult — результат получения портфеля.
type PositionsResult struct {
	Items             []PositionItem
	TotalPositions    int
	TotalInvested     float64
	TotalCurrentValue float64
	TotalPNL          float64
}

// GetPositions возвращает портфель пользователя.
func (s *PositionService) GetPositions(ctx context.Context, userID string) (*PositionsResult, error) {
	positions, err := s.positions.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	result := &PositionsResult{Items: make([]PositionItem, 0, len(positions))}

	for _, pos := range positions {
		market, err := s.markets.GetByID(ctx, pos.MarketID)
		if err != nil {
			continue
		}

		invested := pos.AvgCost * pos.Quantity
		var currentValue float64

		if market.Status == domain.MarketStatusResolved {
			pools, err := s.positions.GetPoolsByMarketID(ctx, market.ID)
			if err == nil {
				totalPool := 0.0
				winningPool := 0.0
				for out, p := range pools {
					totalPool += p
					if market.ResolvedOutcome != nil && out == *market.ResolvedOutcome {
						winningPool += p
					}
				}

				if winningPool == 0 {
					currentValue = invested // Refund
				} else if market.ResolvedOutcome != nil && pos.Outcome == *market.ResolvedOutcome {
					currentValue = (invested / winningPool) * totalPool
				} else {
					currentValue = 0
				}
			}
		} else {
			currentPrice := 1.0 / float64(len(market.Outcomes))
			currentValue = currentPrice * pos.Quantity
		}

		pnl := currentValue - invested

		result.Items = append(result.Items, PositionItem{
			Position: pos, MarketTitle: market.Title,
			CurrentValue: currentValue, PNL: pnl,
		})
		result.TotalInvested += invested
		result.TotalCurrentValue += currentValue
		result.TotalPNL += pnl
	}

	result.TotalPositions = len(result.Items)
	return result, nil
}
