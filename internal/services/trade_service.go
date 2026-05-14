package services

import (
	"context"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
)

// TradeService реализует бизнес-логику сделок.
type TradeService struct {
	trades  repos.TradeRepository
	markets repos.MarketRepository
}

// NewTradeService создаёт новый TradeService.
func NewTradeService(trades repos.TradeRepository, markets repos.MarketRepository) *TradeService {
	return &TradeService{trades: trades, markets: markets}
}

// TradeItem — сделка с названием рынка.
type TradeItem struct {
	domain.Trade
	MarketTitle string
}

// TradesResult — результат получения сделок.
type TradesResult struct {
	Items []TradeItem
	Total int
}

// GetMyTrades возвращает историю сделок пользователя.
func (s *TradeService) GetMyTrades(ctx context.Context, userID string, marketID *string, page, limit int) (*TradesResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := domain.TradeFilter{MarketID: marketID, Page: page, Limit: limit}

	trades, total, err := s.trades.List(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list trades: %w", err)
	}

	items := make([]TradeItem, 0, len(trades))
	for _, t := range trades {
		title := ""
		market, err := s.markets.GetByID(ctx, t.MarketID)
		if err == nil {
			title = market.Title
		}
		items = append(items, TradeItem{Trade: t, MarketTitle: title})
	}

	return &TradesResult{Items: items, Total: total}, nil
}