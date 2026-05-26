package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/handler"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos/mocks"
	"github.com/Cylinder-perhaps/Variator2026/internal/services"
	"github.com/Cylinder-perhaps/Variator2026/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserTestApp(
	balanceRepo *mocks.MockBalanceRepository,
	positionRepo *mocks.MockPositionRepository,
	marketRepo *mocks.MockMarketRepository,
) *handler.App {
	balanceService := services.NewBalanceService(balanceRepo)
	positionService := services.NewPositionService(positionRepo, marketRepo)

	return &handler.App{
		Balances:  balanceService,
		Positions: positionService,
	}
}

func TestGetMyPositions_Success(t *testing.T) {
	// TC-10: Получение позиций (HTTP 200)
	balanceRepo := new(mocks.MockBalanceRepository)
	positionRepo := new(mocks.MockPositionRepository)
	marketRepo := new(mocks.MockMarketRepository)

	app := setupUserTestApp(balanceRepo, positionRepo, marketRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/positions", nil)
	// Эмуляция auth: пользователь user-1
	req = req.WithContext(domain.NewContextWithUser(req.Context(), &domain.User{ID: "user-1", Role: domain.RoleUser}))
	w := httptest.NewRecorder()

	positions := []domain.Position{
		{
			ID:       "pos-1",
			UserID:   "user-1",
			MarketID: "market-1",
			Outcome:  "Yes",
			Quantity: 10,
			AvgCost:  0.5,
		},
	}

	market := &domain.Market{
		ID:       "market-1",
		Title:    "Market 1",
		Status:   domain.MarketStatusActive,
		Outcomes: []string{"Yes", "No"},
	}

	positionRepo.On("GetByUserID", mock.Anything, "user-1").Return(positions, nil)
	marketRepo.On("GetByID", mock.Anything, "market-1").Return(market, nil)

	app.GetMyPositions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp api.PortfolioResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, *resp.Summary.TotalPositions)
	assert.Len(t, *resp.Positions, 1)

	positionRepo.AssertExpectations(t)
	marketRepo.AssertExpectations(t)
}

func TestGetMyBalance_Success(t *testing.T) {
	// TC-11: Баланс (HTTP 200)
	balanceRepo := new(mocks.MockBalanceRepository)
	positionRepo := new(mocks.MockPositionRepository)
	marketRepo := new(mocks.MockMarketRepository)

	app := setupUserTestApp(balanceRepo, positionRepo, marketRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/balance", nil)
	// Эмуляция auth: пользователь user-1
	req = req.WithContext(domain.NewContextWithUser(req.Context(), &domain.User{ID: "user-1", Role: domain.RoleUser}))
	w := httptest.NewRecorder()

	balance := &domain.Balance{
		UserID:    "user-1",
		Available: 100.0,
		Blocked:   20.0,
	}

	balanceRepo.On("GetByUserID", mock.Anything, "user-1").Return(balance, nil)

	app.GetMyBalance(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp api.BalanceResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, float32(100.0), *resp.Available)
	assert.Equal(t, float32(20.0), *resp.Blocked)

	balanceRepo.AssertExpectations(t)
}
