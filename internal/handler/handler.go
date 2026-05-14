package handler

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	mw "github.com/Cylinder-perhaps/Variator2026/internal/middleware"
	"github.com/Cylinder-perhaps/Variator2026/internal/services"
	"github.com/Cylinder-perhaps/Variator2026/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// App реализует api.ServerInterface — все 13 эндпоинтов.
type App struct {
	Auth      *services.AuthService
	Markets   *services.MarketService
	Orders    *services.OrderService
	Balances  *services.BalanceService
	Positions *services.PositionService
	Trades    *services.TradeService
}

// --- Helpers ---

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, code api.ErrorResponseErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{
		Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Details *map[string]interface{}     `json:"details,omitempty"`
			Message string                     `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}

func paginationMeta(total, page, limit int) *api.PaginationMeta {
	pages := int(math.Ceil(float64(total) / float64(limit)))
	return &api.PaginationMeta{
		Page:    &page,
		PerPage: &limit,
		Total:   &total,
		Pages:   &pages,
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

// --- 1. Register ---

func (h *App) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, "Неверный формат запроса")
		return
	}

	email := string(req.Email)
	if !isValidEmail(email) {
		respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, "Email должен быть валидным")
		return
	}

	if !isValidPassword(req.Password) {
		respondError(w, http.StatusUnprocessableEntity, api.UNPROCESSABLEENTITY, "Пароль должен содержать минимум 8 символов, буквы и цифры")
		return
	}

	result, err := h.Auth.Register(r.Context(), email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			respondError(w, http.StatusConflict, api.CONFLICT, "Email уже существует в системе")
			return
		}
		log.Printf("[ERROR] RegisterUser failed: %v", err)
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	respondJSON(w, http.StatusCreated, toAuthResponse(result))
}

// --- 2. Login ---

func (h *App) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, "Неверный формат запроса")
		return
	}

	result, err := h.Auth.Login(r.Context(), string(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Email или пароль неверны")
			return
		}
		log.Printf("[ERROR] LoginUser failed: %v", err)
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	respondJSON(w, http.StatusOK, toAuthResponse(result))
}

// --- 3. Refresh ---

func (h *App) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req api.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, "Неверный формат запроса")
		return
	}

	accessToken, err := h.Auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) || errors.Is(err, domain.ErrTokenExpired) {
			respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Refresh токен истёк или невалиден")
			return
		}
		if errors.Is(err, domain.ErrTokenRevoked) {
			respondError(w, http.StatusForbidden, api.FORBIDDEN, "Токен отозван")
			return
		}
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	respondJSON(w, http.StatusOK, api.RefreshTokenResponse{AccessToken: accessToken})
}

// --- 4. Logout ---

func (h *App) LogoutUser(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	if err := h.Auth.Logout(r.Context(), userID); err != nil {
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- 5. List Markets ---

func (h *App) ListMarkets(w http.ResponseWriter, r *http.Request, params api.ListMarketsParams) {
	page := 1
	limit := 20
	if params.Page != nil {
		page = *params.Page
	}
	if params.PerPage != nil {
		limit = *params.PerPage
	}

	var status *string
	if params.Status != nil {
		s := string(*params.Status)
		status = &s
	}

	sortBy := "created_at"
	if params.SortBy != nil {
		sortBy = string(*params.SortBy)
	}

	markets, total, err := h.Markets.ListMarkets(r.Context(), status, sortBy, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	apiMarkets := make([]api.Market, 0, len(markets))
	for _, m := range markets {
		apiMarkets = append(apiMarkets, toAPIMarket(m))
	}

	respondJSON(w, http.StatusOK, api.MarketsResponse{
		Data: &apiMarkets,
		Meta: paginationMeta(total, page, limit),
	})
}

// --- 6. Get Market Details ---

func (h *App) GetMarketDetails(w http.ResponseWriter, r *http.Request, marketId openapi_types.UUID) {
	market, err := h.Markets.GetMarketDetails(r.Context(), marketId.String())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(w, http.StatusNotFound, api.NOTFOUND, "Рынок не найден")
			return
		}
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	respondJSON(w, http.StatusOK, toAPIMarket(*market))
}

// --- 7. Create Order ---

func (h *App) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	var req api.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, "Неверный формат запроса")
		return
	}

	if req.Quantity <= 0 {
		respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, "Количество должно быть положительным")
		return
	}

	order, err := h.Orders.CreateOrder(r.Context(), userID, req.MarketId.String(), req.Outcome, float64(req.Quantity))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, http.StatusNotFound, api.NOTFOUND, "Рынок не найден")
		case errors.Is(err, domain.ErrMarketNotActive):
			respondError(w, http.StatusConflict, api.CONFLICT, "Рынок не активен")
		case errors.Is(err, domain.ErrInsufficientBalance):
			respondError(w, http.StatusPaymentRequired, api.INSUFFICIENTBALANCE, "Недостаточно средств на счёте")
		case errors.Is(err, domain.ErrInvalidRequest):
			respondError(w, http.StatusUnprocessableEntity, api.UNPROCESSABLEENTITY, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		}
		return
	}

	respondJSON(w, http.StatusCreated, toAPIOrder(*order))
}

// --- 8. Cancel Order ---

func (h *App) CancelOrder(w http.ResponseWriter, r *http.Request, orderId openapi_types.UUID) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	err := h.Orders.CancelOrder(r.Context(), userID, orderId.String())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, http.StatusNotFound, api.NOTFOUND, "Ордер не найден")
		case errors.Is(err, domain.ErrForbidden):
			respondError(w, http.StatusForbidden, api.FORBIDDEN, "Это не ваш ордер")
		case errors.Is(err, domain.ErrOrderAlreadyFilled):
			respondError(w, http.StatusConflict, api.CONFLICT, "Ордер уже исполнен, отмена невозможна")
		default:
			respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- 9. Get My Orders ---

func (h *App) GetMyOrders(w http.ResponseWriter, r *http.Request, params api.GetMyOrdersParams) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	page := 1
	limit := 20
	if params.Page != nil {
		page = *params.Page
	}
	if params.PerPage != nil {
		limit = *params.PerPage
	}

	var status *string
	if params.Status != nil {
		s := string(*params.Status)
		status = &s
	}

	orders, total, err := h.Orders.GetMyOrders(r.Context(), userID, status, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	apiOrders := make([]api.Order, 0, len(orders))
	for _, o := range orders {
		apiOrders = append(apiOrders, toAPIOrder(o))
	}

	respondJSON(w, http.StatusOK, api.OrdersResponse{
		Data: &apiOrders,
		Meta: paginationMeta(total, page, limit),
	})
}

// --- 10. Get Positions ---

func (h *App) GetPositions(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	result, err := h.Positions.GetPositions(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	apiPositions := make([]api.Position, 0, len(result.Items))
	for _, p := range result.Items {
		mid := openapi_types.UUID{}
		mid.UnmarshalText([]byte(p.MarketID))

		apiPositions = append(apiPositions, api.Position{
			MarketId:     mid,
			MarketTitle:  p.MarketTitle,
			Outcome:      p.Outcome,
			Quantity:     float32(p.Quantity),
			AvgCost:      float32(p.AvgCost),
			CurrentValue: float32(p.CurrentValue),
			Pnl:          float32(p.PNL),
		})
	}

	tp := result.TotalPositions
	ti := float32(result.TotalInvested)
	tcv := float32(result.TotalCurrentValue)
	tpnl := float32(result.TotalPNL)

	respondJSON(w, http.StatusOK, api.PositionsResponse{
		Data: &apiPositions,
		Meta: &struct {
			TotalCurrentValue *float32 `json:"total_current_value,omitempty"`
			TotalInvested     *float32 `json:"total_invested,omitempty"`
			TotalPnl          *float32 `json:"total_pnl,omitempty"`
			TotalPositions    *int     `json:"total_positions,omitempty"`
		}{
			TotalPositions:    &tp,
			TotalInvested:     &ti,
			TotalCurrentValue: &tcv,
			TotalPnl:          &tpnl,
		},
	})
}

// --- 11. Get Balance ---

func (h *App) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	balance, err := h.Balances.GetBalance(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(w, http.StatusNotFound, api.NOTFOUND, "Аккаунт не найден")
			return
		}
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	respondJSON(w, http.StatusOK, api.Balance{
		Total:           float32(balance.Total),
		Available:       float32(balance.Available),
		BlockedInOrders: float32(balance.BlockedInOrders),
	})
}

// --- 12. Get My Trades ---

func (h *App) GetMyTrades(w http.ResponseWriter, r *http.Request, params api.GetMyTradesParams) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	page := 1
	limit := 20
	if params.Page != nil {
		page = *params.Page
	}
	if params.PerPage != nil {
		limit = *params.PerPage
	}

	var marketID *string
	if params.MarketId != nil {
		s := params.MarketId.String()
		marketID = &s
	}

	result, err := h.Trades.GetMyTrades(r.Context(), userID, marketID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		return
	}

	tradeItems := make([]struct {
		ExecutedAt  *time.Time          `json:"executed_at,omitempty"`
		Id          *openapi_types.UUID `json:"id,omitempty"`
		MarketId    *openapi_types.UUID `json:"market_id,omitempty"`
		MarketTitle *string             `json:"market_title,omitempty"`
		Outcome     *string             `json:"outcome,omitempty"`
		Pnl         *float32            `json:"pnl,omitempty"`
		Price       *float32            `json:"price,omitempty"`
		Quantity    *float32            `json:"quantity,omitempty"`
	}, 0, len(result.Items))

	for _, t := range result.Items {
		tid := openapi_types.UUID{}
		tid.UnmarshalText([]byte(t.ID))
		mid := openapi_types.UUID{}
		mid.UnmarshalText([]byte(t.MarketID))

		price := float32(t.Price)
		qty := float32(t.Quantity)
		pnl := float32(t.PNL)
		mt := t.MarketTitle
		outcome := t.Outcome
		ea := t.ExecutedAt

		tradeItems = append(tradeItems, struct {
			ExecutedAt  *time.Time          `json:"executed_at,omitempty"`
			Id          *openapi_types.UUID `json:"id,omitempty"`
			MarketId    *openapi_types.UUID `json:"market_id,omitempty"`
			MarketTitle *string             `json:"market_title,omitempty"`
			Outcome     *string             `json:"outcome,omitempty"`
			Pnl         *float32            `json:"pnl,omitempty"`
			Price       *float32            `json:"price,omitempty"`
			Quantity    *float32            `json:"quantity,omitempty"`
		}{
			Id: &tid, MarketId: &mid, MarketTitle: &mt,
			Outcome: &outcome, Quantity: &qty, Price: &price,
			Pnl: &pnl, ExecutedAt: &ea,
		})
	}

	respondJSON(w, http.StatusOK, api.TradesResponse{
		Data: &tradeItems,
		Meta: paginationMeta(result.Total, page, limit),
	})
}

// --- 13. Resolve Market ---

func (h *App) ResolveMarket(w http.ResponseWriter, r *http.Request, marketId openapi_types.UUID) {
	userID := mw.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, api.UNAUTHORIZED, "Не авторизован")
		return
	}

	role := mw.GetUserRole(r.Context())
	if role != "admin" && role != "moderator" {
		respondError(w, http.StatusForbidden, api.FORBIDDEN, "Только Admin и Moderator могут разрешить события")
		return
	}

	var req api.ResolveMarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, "Неверный формат запроса")
		return
	}

	market, totalPayout, err := h.Markets.ResolveMarket(r.Context(), marketId.String(), req.WinningOutcome, req.EvidenceUrl)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, http.StatusNotFound, api.NOTFOUND, "Рынок не найден")
		case errors.Is(err, domain.ErrConflict):
			respondError(w, http.StatusConflict, api.CONFLICT, "Этот рынок уже разрешён")
		case errors.Is(err, domain.ErrInvalidRequest):
			msg := err.Error()
			respondError(w, http.StatusBadRequest, api.INVALIDREQUEST, msg)
		default:
			respondError(w, http.StatusInternalServerError, api.INTERNALERROR, "Внутренняя ошибка сервера")
		}
		return
	}

	mid := openapi_types.UUID{}
	mid.UnmarshalText([]byte(market.ID))

	resolvedAt := time.Now()
	payout := float32(totalPayout)
	status := api.ResolveMarketResponseStatusRESOLVED
	wo := req.WinningOutcome

	respondJSON(w, http.StatusOK, api.ResolveMarketResponse{
		MarketId:       &mid,
		Status:         &status,
		WinningOutcome: &wo,
		TotalPayout:    &payout,
		ResolvedAt:     &resolvedAt,
	})
}

// --- Converters ---

func toAuthResponse(r *services.AuthResponse) api.AuthResponse {
	uid := openapi_types.UUID{}
	uid.UnmarshalText([]byte(r.User.ID))

	return api.AuthResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		User: api.User{
			Id:        uid,
			Email:     openapi_types.Email(r.User.Email),
			Role:      api.UserRole(r.User.Role),
			CreatedAt: r.User.CreatedAt,
		},
	}
}

func toAPIMarket(m domain.Market) api.Market {
	mid := openapi_types.UUID{}
	mid.UnmarshalText([]byte(m.ID))

	market := api.Market{
		Id:              mid,
		Title:           m.Title,
		Description:     m.Description,
		Outcomes:        []string(m.Outcomes),
		Status:          api.MarketStatus(m.Status),
		Deadline:        m.Deadline,
		ResolvedOutcome: m.ResolvedOutcome,
		CreatedAt:       m.CreatedAt,
	}

	if m.CreatedBy != nil {
		cbID := openapi_types.UUID{}
		cbID.UnmarshalText([]byte(*m.CreatedBy))
		market.CreatedBy = &cbID
	}

	return market
}

func toAPIOrder(o domain.Order) api.Order {
	oid := openapi_types.UUID{}
	oid.UnmarshalText([]byte(o.ID))
	mid := openapi_types.UUID{}
	mid.UnmarshalText([]byte(o.MarketID))

	return api.Order{
		Id:         oid,
		MarketId:   mid,
		Outcome:    o.Outcome,
		Quantity:   float32(o.Quantity),
		Price:      float32(o.Price),
		AmountPaid: float32(o.AmountPaid),
		Status:     api.OrderStatus(o.Status),
		CreatedAt:  o.CreatedAt,
	}
}

// Ensure App implements ServerInterface.
var _ api.ServerInterface = (*App)(nil)

// Подавляем предупреждение о неиспользуемом strings.
var _ = strings.NewReader
