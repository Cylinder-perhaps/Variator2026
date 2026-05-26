package mocks

import (
	"context"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository — мок для repos.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, id string, role domain.UserRole) error {
	args := m.Called(ctx, id, role)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ListAll(ctx context.Context, page, limit int) ([]domain.User, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

// MockRefreshTokenRepository — мок для repos.RefreshTokenRepository
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.RefreshToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockMarketRepository — мок для repos.MarketRepository
type MockMarketRepository struct {
	mock.Mock
}

func (m *MockMarketRepository) Create(ctx context.Context, market *domain.Market) error {
	args := m.Called(ctx, market)
	return args.Error(0)
}

func (m *MockMarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Market), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMarketRepository) GetByExternalID(ctx context.Context, externalID, source string) (*domain.Market, error) {
	args := m.Called(ctx, externalID, source)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Market), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMarketRepository) List(ctx context.Context, filter domain.MarketFilter) ([]domain.Market, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Market), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockMarketRepository) ListByExternalSource(ctx context.Context, source string) ([]domain.Market, error) {
	args := m.Called(ctx, source)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Market), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMarketRepository) Update(ctx context.Context, market *domain.Market) error {
	args := m.Called(ctx, market)
	return args.Error(0)
}
