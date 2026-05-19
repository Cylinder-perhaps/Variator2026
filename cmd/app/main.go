package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Cylinder-perhaps/Variator2026/internal/config"
	"github.com/Cylinder-perhaps/Variator2026/internal/handler"
	mw "github.com/Cylinder-perhaps/Variator2026/internal/middleware"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
	"github.com/Cylinder-perhaps/Variator2026/internal/services"
	"github.com/Cylinder-perhaps/Variator2026/pkg/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// HealthResponse — ответ health-check эндпоинта.
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// --- Загрузка конфигурации ---
	cfg := config.MustLoad()

	log.Printf("🔧 Config loaded: server port=%d, db=%s:%d/%s",
		cfg.Server.Port, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	// --- Подключение к БД ---
	storage, err := repos.NewStorage(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer storage.Close()

	storage.ConfigurePool(cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	log.Println("✅ Connected to PostgreSQL")

	// --- Инициализация сервисов ---
	authService := services.NewAuthService(
		storage.Users, storage.RefreshTokens,
		cfg.JWT, cfg.Auth,
	)

	marketService := services.NewMarketService(
		storage.Markets, storage.Orders, storage.Balances,
		storage.Positions, storage.Trades,
	)

	orderService := services.NewOrderService(
		storage.Orders, storage.Markets, storage.Balances,
		storage.Positions, storage.Trades,
	)

	balanceService := services.NewBalanceService(storage.Balances)
	positionService := services.NewPositionService(storage.Positions, storage.Markets)
	tradeService := services.NewTradeService(storage.Trades, storage.Markets)

	// --- Инициализация handler ---
	appHandler := &handler.App{
		Auth:      authService,
		Markets:   marketService,
		Orders:    orderService,
		Balances:  balanceService,
		Positions: positionService,
		Trades:    tradeService,
	}

	// --- Роутер ---
	r := chi.NewRouter()

	// Глобальные middleware.
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(corsMiddleware)

	// Health-check эндпоинты (без авторизации).
	r.Get("/", healthHandler)
	r.Get("/health", healthHandler)
	r.Get("/api/health", healthHandler)

	// Публичные маршруты (без JWT).
	r.Group(func(r chi.Router) {
		// Auth эндпоинты.
		r.Post("/api/auth/register", appHandler.RegisterUser)
		r.Post("/api/auth/login", appHandler.LoginUser)
		r.Post("/api/auth/refresh", appHandler.RefreshToken)

		// Публичные рынки.
		r.Get("/api/markets", func(w http.ResponseWriter, req *http.Request) {
			params := parseListMarketsParams(req)
			appHandler.ListMarkets(w, req, params)
		})
		r.Get("/api/markets/{marketId}", func(w http.ResponseWriter, req *http.Request) {
			marketID, err := parseUUIDParam(req, "marketId")
			if err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Неверный формат UUID")
				return
			}
			appHandler.GetMarketDetails(w, req, marketID)
		})
	})

	// Защищённые маршруты (требуют JWT).
	r.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware(cfg.JWT.Secret))

		r.Post("/api/auth/logout", appHandler.LogoutUser)

		// Ордера.
		r.Post("/api/orders", appHandler.CreateOrder)
		r.Get("/api/orders", func(w http.ResponseWriter, req *http.Request) {
			params := parseGetMyOrdersParams(req)
			appHandler.GetMyOrders(w, req, params)
		})
		r.Delete("/api/orders/{orderId}", func(w http.ResponseWriter, req *http.Request) {
			orderID, err := parseUUIDParam(req, "orderId")
			if err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Неверный формат UUID")
				return
			}
			appHandler.CancelOrder(w, req, orderID)
		})

		// Портфель.
		r.Get("/api/positions", appHandler.GetPositions)
		r.Get("/api/balance", appHandler.GetBalance)
		r.Get("/api/trades", func(w http.ResponseWriter, req *http.Request) {
			params := parseGetMyTradesParams(req)
			appHandler.GetMyTrades(w, req, params)
		})

		// Admin-only маршруты.
		r.Route("/api/admin", func(r chi.Router) {
			r.Use(mw.RequireRole("admin", "moderator"))

			r.Post("/markets", appHandler.CreateMarket)
			r.Post("/markets/{marketId}/resolve", func(w http.ResponseWriter, req *http.Request) {
				marketID, err := parseUUIDParam(req, "marketId")
				if err != nil {
					writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Неверный формат UUID")
					return
				}
				appHandler.ResolveMarket(w, req, marketID)
			})
		})
	})

	// --- HTTP Server ---
	port := strconv.Itoa(cfg.Server.Port)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown.
	go func() {
		log.Printf("🚀 Variator API server started on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}

// --- Вспомогательные функции ---

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status:  "ok",
		Message: "Variator API v1 MVP is running",
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{"code": code, "message": message},
	})
}

func parseUUIDParam(r *http.Request, name string) (openapi_types.UUID, error) {
	raw := chi.URLParam(r, name)
	var uid openapi_types.UUID
	err := uid.UnmarshalText([]byte(raw))
	return uid, err
}

func parseListMarketsParams(r *http.Request) api.ListMarketsParams {
	params := api.ListMarketsParams{}

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.Page = &n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.PerPage = &n
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := api.ListMarketsParamsStatus(v)
		params.Status = &s
	}
	if v := r.URL.Query().Get("sort_by"); v != "" {
		s := api.ListMarketsParamsSortBy(v)
		params.SortBy = &s
	}

	return params
}

func parseGetMyOrdersParams(r *http.Request) api.GetMyOrdersParams {
	params := api.GetMyOrdersParams{}

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.Page = &n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.PerPage = &n
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := api.GetMyOrdersParamsStatus(v)
		params.Status = &s
	}

	return params
}

func parseGetMyTradesParams(r *http.Request) api.GetMyTradesParams {
	params := api.GetMyTradesParams{}

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.Page = &n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.PerPage = &n
		}
	}
	if v := r.URL.Query().Get("market_id"); v != "" {
		var uid openapi_types.UUID
		if err := uid.UnmarshalText([]byte(v)); err == nil {
			params.MarketId = &uid
		}
	}

	return params
}
