package handler_test

import (
	"bytes"
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
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func setupOrderTestApp(
	orderRepo *mocks.MockOrderRepository,
	marketRepo *mocks.MockMarketRepository,
	balanceRepo *mocks.MockBalanceRepository,
	positionRepo *mocks.MockPositionRepository,
	tradeRepo *mocks.MockTradeRepository,
) *handler.App {
	orderService := services.NewOrderService(orderRepo, marketRepo, balanceRepo, positionRepo, tradeRepo)

	return &handler.App{
		Orders: orderService,
	}
}

func TestCreateOrder_Success(t *testing.T) {
	// TC-08: Создание ордера (HTTP 201)
	orderRepo := new(mocks.MockOrderRepository)
	marketRepo := new(mocks.MockMarketRepository)
	balanceRepo := new(mocks.MockBalanceRepository)
	positionRepo := new(mocks.MockPositionRepository)
	tradeRepo := new(mocks.MockTradeRepository)

	app := setupOrderTestApp(orderRepo, marketRepo, balanceRepo, positionRepo, tradeRepo)

	marketID := uuid.New()
	marketIDStr := marketID.String()

	reqBody := api.CreateOrderRequest{
		MarketId: marketID,
		Outcome:  "Yes",
		Quantity: 10,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	// Эмуляция того, что запрос прошел через auth middleware
	req = req.WithContext(domain.NewContextWithUser(req.Context(), &domain.User{ID: "user-1", Role: domain.RoleUser}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	market := &domain.Market{
		ID:       marketIDStr,
		Status:   domain.MarketStatusActive,
		Outcomes: []string{"Yes", "No"},
	}

	marketRepo.On("GetByID", mock.Anything, marketIDStr).Return(market, nil)
	balanceRepo.On("UpdateBalance", mock.Anything, "user-1", -5.0, 5.0).Return(nil)
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)
	positionRepo.On("GetByUserMarketOutcome", mock.Anything, "user-1", marketIDStr, "Yes").Return(nil, domain.ErrNotFound)
	positionRepo.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.Position")).Return(nil)
	tradeRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Trade")).Return(nil)
	balanceRepo.On("UpdateBalance", mock.Anything, "user-1", 0.0, -5.0).Return(nil)

	app.CreateOrder(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var resp api.Order
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, float32(10), resp.Quantity)

	marketRepo.AssertExpectations(t)
	balanceRepo.AssertExpectations(t)
	orderRepo.AssertExpectations(t)
	positionRepo.AssertExpectations(t)
	tradeRepo.AssertExpectations(t)
}

func TestCancelOrder_Forbidden(t *testing.T) {
	// TC-09: Отмена ордера (чужой ордер) (HTTP 403)
	orderRepo := new(mocks.MockOrderRepository)
	marketRepo := new(mocks.MockMarketRepository)
	balanceRepo := new(mocks.MockBalanceRepository)
	positionRepo := new(mocks.MockPositionRepository)
	tradeRepo := new(mocks.MockTradeRepository)

	app := setupOrderTestApp(orderRepo, marketRepo, balanceRepo, positionRepo, tradeRepo)

	orderID := uuid.New()
	orderIDStr := orderID.String()

	req := httptest.NewRequest(http.MethodDelete, "/api/orders/"+orderIDStr, nil)
	// Эмуляция auth: пользователь user-2 пытается отменить ордер пользователя user-1
	req = req.WithContext(domain.NewContextWithUser(req.Context(), &domain.User{ID: "user-2", Role: domain.RoleUser}))
	w := httptest.NewRecorder()

	existingOrder := &domain.Order{
		ID:     orderIDStr,
		UserID: "user-1",
		Status: domain.OrderStatusPending,
	}

	orderRepo.On("GetByID", mock.Anything, orderIDStr).Return(existingOrder, nil)

	app.CancelOrder(w, req, orderID)

	assert.Equal(t, http.StatusForbidden, w.Code)
	orderRepo.AssertExpectations(t)
}
