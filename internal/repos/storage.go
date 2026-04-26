package repos

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/config"
	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos/postgresql"
)

type Storage struct {
	db        *sql.DB
	Users     UserRepository
	Markets   MarketRepository
	Balances  BalanceRepository
	Orders    OrderRepository
	Positions PositionRepository
	Trades    TradeRepository
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

type MarketRepository interface {
	Create(ctx context.Context, market *domain.Market) error
	GetByID(ctx context.Context, id string) (*domain.Market, error)
	GetAll(ctx context.Context) ([]domain.Market, error)
	Update(ctx context.Context, market *domain.Market) error
}

type BalanceRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Balance, error)
	Update(ctx context.Context, balance *domain.Balance) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByUserID(ctx context.Context, userId string) ([]domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id string) error
}

type PositionRepository interface {
	GetByUserMarketOutcome(ctx context.Context, userID, marketID, outcome string) (*domain.Position, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Position, error)
	Create(ctx context.Context, position *domain.Position) error
	Update(ctx context.Context, position *domain.Position) error
	Delete(ctx context.Context, id string) error
}

type TradeRepository interface {
	Create(ctx context.Context, trade *domain.Trade) error
	GetByUserID(ctx context.Context, userID string) ([]domain.Trade, error)
	GetByOrderID(ctx context.Context, orderID string) ([]domain.Trade, error)
	Update(ctx context.Context, trade *domain.Trade) error
	Delete(ctx context.Context, id string) error
}

func NewStorage(ctx context.Context, cfg *config.Config) (*Storage, error) {
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{
		db:        db,
		Users:     postgresql.NewPostgresUserRepository(db),
		Markets:   postgresql.NewPostgresMarketRepository(db),
		Balances:  postgresql.NewPostgresBalanceRepository(db),
		Orders:    postgresql.NewPostgresOrderRepository(db),
		Positions: postgresql.NewPostgresPositionRepository(db),
		Trades:    postgresql.NewPostgresTradeRepository(db),
	}

	return storage, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
