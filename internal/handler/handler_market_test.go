package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/handler"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos/mocks"
	"github.com/Cylinder-perhaps/Variator2026/internal/services"
	"github.com/Cylinder-perhaps/Variator2026/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/google/uuid"
)

func setupMarketTestApp(
	marketRepo *mocks.MockMarketRepository,
	orderRepo *mocks.MockOrderRepository,
	balanceRepo *mocks.MockBalanceRepository,
	positionRepo *mocks.MockPositionRepository,
	tradeRepo *mocks.MockTradeRepository,
) *handler.App {
	marketService := services.NewMarketService(marketRepo, orderRepo, balanceRepo, positionRepo, tradeRepo)

	return &handler.App{
		Markets: marketService,
	}
}

func TestListMarkets_Success(t *testing.T) {
	// TC-07: Просмотр списка рынков (HTTP 200)
	marketRepo := new(mocks.MockMarketRepository)
	orderRepo := new(mocks.MockOrderRepository)
	balanceRepo := new(mocks.MockBalanceRepository)
	positionRepo := new(mocks.MockPositionRepository)
	tradeRepo := new(mocks.MockTradeRepository)

	app := setupMarketTestApp(marketRepo, orderRepo, balanceRepo, positionRepo, tradeRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/markets", nil)
	w := httptest.NewRecorder()

	markets := []domain.Market{
		{ID: "market-1", Title: "Market 1", Status: domain.MarketStatusActive},
	}

	marketRepo.On("List", mock.Anything, mock.AnythingOfType("domain.MarketFilter")).Return(markets, 1, nil)
	positionRepo.On("GetPoolsByMarketID", mock.Anything, "market-1").Return(map[string]float64{"Yes": 100}, nil)

	// В реальном коде oapi-codegen передает параметры в хендлер. 
	// Мы эмулируем вызов ListMarkets.
	app.ListMarkets(w, req, api.ListMarketsParams{})

	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp api.MarketListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Market 1", resp.Data[0].Title)

	marketRepo.AssertExpectations(t)
	positionRepo.AssertExpectations(t)
}

func TestResolveMarket_Success(t *testing.T) {
	// TC-12: Разрешение рынка (Admin) (HTTP 200)
	marketRepo := new(mocks.MockMarketRepository)
	orderRepo := new(mocks.MockOrderRepository)
	balanceRepo := new(mocks.MockBalanceRepository)
	positionRepo := new(mocks.MockPositionRepository)
	tradeRepo := new(mocks.MockTradeRepository)

	app := setupMarketTestApp(marketRepo, orderRepo, balanceRepo, positionRepo, tradeRepo)

	marketID := uuid.New()
	marketIDStr := marketID.String()

	reqBody := api.ResolveMarketRequest{
		WinningOutcome: "Yes",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/markets/"+marketIDStr+"/resolve", bytes.NewReader(body))
	// Эмуляция того, что запрос прошел через auth middleware
	req = req.WithContext(domain.NewContextWithUser(req.Context(), &domain.User{ID: "admin-1", Role: domain.RoleAdmin}))
	
	w := httptest.NewRecorder()

	market := &domain.Market{
		ID:       marketIDStr,
		Status:   domain.MarketStatusActive,
		Outcomes: []string{"Yes", "No"},
	}

	marketRepo.On("GetByID", mock.Anything, marketIDStr).Return(market, nil)
	orderRepo.On("GetPendingByMarket", mock.Anything, marketIDStr).Return([]domain.Order{}, nil)
	marketRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Market")).Return(nil)
	positionRepo.On("GetByMarketID", mock.Anything, marketIDStr).Return([]domain.Position{}, nil)

	app.ResolveMarket(w, req, marketID)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp api.ResolveMarketResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "Yes", resp.ResolvedOutcome)

	marketRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
	positionRepo.AssertExpectations(t)
}
