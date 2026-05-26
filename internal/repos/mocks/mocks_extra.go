package mocks

import (
	"context"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository — мок для repos.OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) List(ctx context.Context, userID string, filter domain.OrderFilter) ([]domain.Order, int, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Order), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) GetPendingByMarket(ctx context.Context, marketID string) ([]domain.Order, error) {
	args := m.Called(ctx, marketID)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockBalanceRepository — мок для repos.BalanceRepository
type MockBalanceRepository struct {
	mock.Mock
}

func (m *MockBalanceRepository) GetByUserID(ctx context.Context, userID string) (*domain.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Balance), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBalanceRepository) UpdateBalance(ctx context.Context, userID string, availableDelta, blockedDelta float64) error {
	args := m.Called(ctx, userID, availableDelta, blockedDelta)
	return args.Error(0)
}

// MockPositionRepository — мок для repos.PositionRepository
type MockPositionRepository struct {
	mock.Mock
}

func (m *MockPositionRepository) GetByUserMarketOutcome(ctx context.Context, userID, marketID, outcome string) (*domain.Position, error) {
	args := m.Called(ctx, userID, marketID, outcome)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Position), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPositionRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Position, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Position), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPositionRepository) GetByMarketID(ctx context.Context, marketID string) ([]domain.Position, error) {
	args := m.Called(ctx, marketID)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Position), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPositionRepository) GetPoolsByMarketID(ctx context.Context, marketID string) (map[string]float64, error) {
	args := m.Called(ctx, marketID)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]float64), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPositionRepository) Upsert(ctx context.Context, position *domain.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

// MockTradeRepository — мок для repos.TradeRepository
type MockTradeRepository struct {
	mock.Mock
}

func (m *MockTradeRepository) Create(ctx context.Context, trade *domain.Trade) error {
	args := m.Called(ctx, trade)
	return args.Error(0)
}

func (m *MockTradeRepository) List(ctx context.Context, userID string, filter domain.TradeFilter) ([]domain.Trade, int, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Trade), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}
