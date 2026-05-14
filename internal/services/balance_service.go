package services

import (
	"context"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
)

// BalanceService реализует бизнес-логику баланса.
type BalanceService struct {
	balances repos.BalanceRepository
}

// NewBalanceService создаёт новый BalanceService.
func NewBalanceService(balances repos.BalanceRepository) *BalanceService {
	return &BalanceService{balances: balances}
}

// GetBalance возвращает баланс пользователя.
func (s *BalanceService) GetBalance(ctx context.Context, userID string) (*domain.Balance, error) {
	balance, err := s.balances.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}
