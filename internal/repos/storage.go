package repos

import (
	"context"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos/postgresql"
	"github.com/jmoiron/sqlx"

	// PostgreSQL driver.
	_ "github.com/lib/pq"
)

// Storage объединяет все репозитории.
type Storage struct {
	db            *sqlx.DB
	Users         UserRepository
	Markets       MarketRepository
	Balances      BalanceRepository
	Orders        OrderRepository
	Positions     PositionRepository
	Trades        TradeRepository
	RefreshTokens RefreshTokenRepository
}

// UserRepository описывает методы работы с пользователями.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateRole(ctx context.Context, id string, role domain.UserRole) error
	Delete(ctx context.Context, id string) error
	ListAll(ctx context.Context, page, limit int) ([]domain.User, int, error)
}

// MarketRepository описывает методы работы с рынками.
type MarketRepository interface {
	Create(ctx context.Context, market *domain.Market) error
	GetByID(ctx context.Context, id string) (*domain.Market, error)
	GetByExternalID(ctx context.Context, externalID, source string) (*domain.Market, error)
	List(ctx context.Context, filter domain.MarketFilter) ([]domain.Market, int, error)
	ListByExternalSource(ctx context.Context, source string) ([]domain.Market, error)
	Update(ctx context.Context, market *domain.Market) error
}

// BalanceRepository описывает методы работы с балансами.
type BalanceRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Balance, error)
	UpdateBalance(ctx context.Context, userID string, availableDelta, blockedDelta float64) error
}

// OrderRepository описывает методы работы с ордерами.
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	List(ctx context.Context, userID string, filter domain.OrderFilter) ([]domain.Order, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	GetPendingByMarket(ctx context.Context, marketID string) ([]domain.Order, error)
}

// PositionRepository описывает методы работы с позициями.
type PositionRepository interface {
	GetByUserMarketOutcome(ctx context.Context, userID, marketID, outcome string) (*domain.Position, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Position, error)
	GetByMarketID(ctx context.Context, marketID string) ([]domain.Position, error)
	GetPoolsByMarketID(ctx context.Context, marketID string) (map[string]float64, error)
	Upsert(ctx context.Context, position *domain.Position) error
}

// TradeRepository описывает методы работы с сделками.
type TradeRepository interface {
	Create(ctx context.Context, trade *domain.Trade) error
	List(ctx context.Context, userID string, filter domain.TradeFilter) ([]domain.Trade, int, error)
}

// RefreshTokenRepository описывает методы работы с refresh-токенами.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID string) error
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
}

// NewStorage создаёт новое хранилище с подключением к PostgreSQL.
func NewStorage(dsn string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	storage := &Storage{
		db:            db,
		Users:         postgresql.NewUserRepository(db),
		Markets:       postgresql.NewMarketRepository(db),
		Balances:      postgresql.NewBalanceRepository(db),
		Orders:        postgresql.NewOrderRepository(db),
		Positions:     postgresql.NewPositionRepository(db),
		Trades:        postgresql.NewTradeRepository(db),
		RefreshTokens: postgresql.NewRefreshTokenRepository(db),
	}

	return storage, nil
}

// ConfigurePool устанавливает параметры пула подключений.
func (s *Storage) ConfigurePool(maxOpen, maxIdle int) {
	s.db.SetMaxOpenConns(maxOpen)
	s.db.SetMaxIdleConns(maxIdle)
}

// Close закрывает подключение к БД.
func (s *Storage) Close() error {
	return s.db.Close()
}

// DB возвращает *sqlx.DB для случаев, когда нужен прямой доступ (транзакции и т.д.).
func (s *Storage) DB() *sqlx.DB {
	return s.db
}
