package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
	"github.com/google/uuid"
)

// OrderService реализует бизнес-логику ордеров.
type OrderService struct {
	orders    repos.OrderRepository
	markets   repos.MarketRepository
	balances  repos.BalanceRepository
	positions repos.PositionRepository
	trades    repos.TradeRepository
}

// NewOrderService создаёт новый OrderService.
func NewOrderService(
	orders repos.OrderRepository,
	markets repos.MarketRepository,
	balances repos.BalanceRepository,
	positions repos.PositionRepository,
	trades repos.TradeRepository,
) *OrderService {
	return &OrderService{
		orders:    orders,
		markets:   markets,
		balances:  balances,
		positions: positions,
		trades:    trades,
	}
}

// CreateOrder создаёт новый ордер на покупку акции.
func (s *OrderService) CreateOrder(ctx context.Context, userID, marketID, outcome string, quantity float64) (*domain.Order, error) {
	// Проверяем рынок.
	market, err := s.markets.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}

	if market.Status != domain.MarketStatusActive {
		return nil, domain.ErrMarketNotActive
	}

	// Проверяем, что исход валиден.
	validOutcome := false
	for _, o := range market.Outcomes {
		if o == outcome {
			validOutcome = true
			break
		}
	}
	if !validOutcome {
		return nil, fmt.Errorf("%w: invalid outcome '%s'", domain.ErrInvalidRequest, outcome)
	}

	// Рассчитываем цену (MVP: фиксированная цена = 1/кол-во исходов).
	price := 1.0 / float64(len(market.Outcomes))
	amountPaid := price * quantity

	// Блокируем средства.
	err = s.balances.UpdateBalance(ctx, userID, -amountPaid, amountPaid)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientBalance) {
			return nil, domain.ErrInsufficientBalance
		}
		return nil, fmt.Errorf("failed to block funds: %w", err)
	}

	order := &domain.Order{
		ID:         uuid.New().String(),
		UserID:     userID,
		MarketID:   marketID,
		Outcome:    outcome,
		Quantity:   quantity,
		Price:      price,
		AmountPaid: amountPaid,
		Status:     domain.OrderStatusFilled,
	}

	if err := s.orders.Create(ctx, order); err != nil {
		// Откатываем блокировку средств при ошибке.
		_ = s.balances.UpdateBalance(ctx, userID, amountPaid, -amountPaid)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Обновляем/создаём позицию.
	existingPos, err := s.positions.GetByUserMarketOutcome(ctx, userID, marketID, outcome)
	if errors.Is(err, domain.ErrNotFound) {
		// Создаём новую позицию.
		pos := &domain.Position{
			ID:       uuid.New().String(),
			UserID:   userID,
			MarketID: marketID,
			Outcome:  outcome,
			Quantity: quantity,
			AvgCost:  price,
		}
		if err := s.positions.Upsert(ctx, pos); err != nil {
			return nil, fmt.Errorf("failed to create position: %w", err)
		}
	} else if err == nil {
		// Обновляем существующую позицию (средневзвешенная стоимость).
		totalCost := existingPos.AvgCost*existingPos.Quantity + price*quantity
		newQuantity := existingPos.Quantity + quantity
		existingPos.AvgCost = totalCost / newQuantity
		existingPos.Quantity = newQuantity
		if err := s.positions.Upsert(ctx, existingPos); err != nil {
			return nil, fmt.Errorf("failed to update position: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to check position: %w", err)
	}

	// Создаём запись о сделке.
	trade := &domain.Trade{
		ID:         uuid.New().String(),
		OrderID:    &order.ID,
		UserID:     userID,
		MarketID:   marketID,
		Outcome:    outcome,
		Quantity:   quantity,
		Price:      price,
		PNL:        0,
		ExecutedAt: order.CreatedAt,
	}
	if err := s.trades.Create(ctx, trade); err != nil {
		return nil, fmt.Errorf("failed to create trade: %w", err)
	}

	// Переводим заблокированные средства в «израсходованные» (для заполненного ордера).
	err = s.balances.UpdateBalance(ctx, userID, 0, -amountPaid)
	if err != nil && !errors.Is(err, domain.ErrInsufficientBalance) {
		return nil, fmt.Errorf("failed to finalize balance: %w", err)
	}

	return order, nil
}

// CancelOrder отменяет ордер и возвращает средства.
func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID string) error {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Проверяем владельца.
	if order.UserID != userID {
		return domain.ErrForbidden
	}

	// Можно отменить только PENDING ордера.
	if order.Status != domain.OrderStatusPending {
		return domain.ErrOrderAlreadyFilled
	}

	// Обновляем статус.
	if err := s.orders.UpdateStatus(ctx, orderID, domain.OrderStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Возвращаем заблокированные средства.
	if err := s.balances.UpdateBalance(ctx, userID, order.AmountPaid, -order.AmountPaid); err != nil {
		return fmt.Errorf("failed to refund: %w", err)
	}

	return nil
}

// GetMyOrders возвращает список ордеров пользователя.
func (s *OrderService) GetMyOrders(ctx context.Context, userID string, status *string, page, limit int) ([]domain.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := domain.OrderFilter{
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	orders, total, err := s.orders.List(ctx, userID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, total, nil
}
