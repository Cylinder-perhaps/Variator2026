package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
	"github.com/Cylinder-perhaps/Variator2026/internal/services"
)

const (
	polymarketAPI     = "https://gamma-api.polymarket.com"
	polymarketSource  = "polymarket"
	syncInterval      = 5 * time.Minute
	httpClientTimeout = 10 * time.Second
)

// PolymarketSync — фоновый worker для синхронизации рынков с Polymarket.
type PolymarketSync struct {
	markets       repos.MarketRepository
	marketService *services.MarketService
	client        *http.Client
}

// NewPolymarketSync создаёт новый worker для синхронизации с Polymarket.
func NewPolymarketSync(markets repos.MarketRepository, marketService *services.MarketService) *PolymarketSync {
	return &PolymarketSync{
		markets:       markets,
		marketService: marketService,
		client: &http.Client{
			Timeout: httpClientTimeout,
		},
	}
}

// polymarketMarket — структура ответа Polymarket API.
type polymarketMarket struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	Description   string `json:"description"`
	Outcomes      string `json:"outcomes"`      // JSON-строка: '["Yes", "No"]'
	OutcomePrices string `json:"outcomePrices"` // JSON-строка: '["0.54", "0.46"]'
	EndDate       string `json:"endDate"`
	Active        bool   `json:"active"`
	Closed        bool   `json:"closed"`
}

// Start запускает периодическую синхронизацию.
func (s *PolymarketSync) Start(ctx context.Context) {
	log.Printf("🔄 Polymarket Sync worker started (interval: %s)", syncInterval)

	// Первый запуск сразу.
	s.sync(ctx)

	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Polymarket Sync worker stopped")
			return
		case <-ticker.C:
			s.sync(ctx)
		}
	}
}

// sync выполняет один цикл синхронизации.
func (s *PolymarketSync) sync(ctx context.Context) {
	// Получаем все наши ACTIVE рынки с source=polymarket.
	ourMarkets, err := s.markets.ListByExternalSource(ctx, polymarketSource)
	if err != nil {
		log.Printf("⚠️  Polymarket Sync: ошибка получения рынков: %v", err)
		return
	}

	if len(ourMarkets) == 0 {
		return
	}

	resolved := 0
	for _, our := range ourMarkets {
		if our.ExternalID == nil {
			continue
		}

		polyMarket, err := s.fetchPolymarketMarket(ctx, *our.ExternalID)
		if err != nil {
			log.Printf("⚠️  Polymarket Sync: ошибка для market %s (ext: %s): %v", our.ID, *our.ExternalID, err)
			continue
		}

		// Проверяем, закрыт ли рынок на Polymarket.
		if !polyMarket.Closed {
			continue
		}

		// Определяем победивший исход.
		winningOutcome := s.detectWinningOutcome(polyMarket)
		if winningOutcome == "" {
			log.Printf("⚠️  Polymarket Sync: не удалось определить исход для market %s (ext: %s)", our.ID, *our.ExternalID)
			continue
		}

		// Формируем evidence URL.
		evidenceURL := fmt.Sprintf("https://polymarket.com/event/%s", polyMarket.ID)

		// Вызываем ResolveMarket через сервис (бизнес-логика: закрытие ордеров, начисления).
		_, _, err = s.marketService.ResolveMarket(ctx, our.ID, winningOutcome, &evidenceURL)
		if err != nil {
			log.Printf("⚠️  Polymarket Sync: ошибка resolve market %s: %v", our.ID, err)
			continue
		}

		resolved++
		log.Printf("✅ Polymarket Sync: рынок «%s» resolved → %s", our.Title, winningOutcome)
	}

	if resolved > 0 {
		log.Printf("🔄 Polymarket Sync: resolved %d рынков", resolved)
	}
}

// fetchPolymarketMarket получает данные рынка из Polymarket API.
func (s *PolymarketSync) fetchPolymarketMarket(ctx context.Context, externalID string) (*polymarketMarket, error) {
	url := fmt.Sprintf("%s/markets/%s", polymarketAPI, externalID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var market polymarketMarket
	if err := json.Unmarshal(body, &market); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &market, nil
}

// detectWinningOutcome определяет победивший исход по ценам.
// Если цена исхода = 1.0 (или ≥ 0.95) — это победитель.
func (s *PolymarketSync) detectWinningOutcome(m *polymarketMarket) string {
	var outcomes []string
	if err := json.Unmarshal([]byte(m.Outcomes), &outcomes); err != nil {
		return ""
	}

	var prices []string
	if err := json.Unmarshal([]byte(m.OutcomePrices), &prices); err != nil {
		return ""
	}

	if len(outcomes) != len(prices) {
		return ""
	}

	for i, priceStr := range prices {
		// Парсим цену.
		var price float64
		if _, err := fmt.Sscanf(priceStr, "%f", &price); err != nil {
			continue
		}
		// Если цена ≥ 0.95 — это победитель (Polymarket ставит 1.0 для resolved).
		if price >= 0.95 {
			return outcomes[i]
		}
	}

	return ""
}

// ImportMarkets импортирует новые рынки из Polymarket (для seed-скрипта).
func (s *PolymarketSync) ImportMarkets(ctx context.Context, limit int, adminUserID string) (int, error) {
	url := fmt.Sprintf("%s/markets?limit=%d&active=true&closed=false", polymarketAPI, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch markets: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read body: %w", err)
	}

	var polyMarkets []polymarketMarket
	if err := json.Unmarshal(body, &polyMarkets); err != nil {
		return 0, fmt.Errorf("failed to unmarshal: %w", err)
	}

	imported := 0
	source := polymarketSource
	for _, pm := range polyMarkets {
		// Проверяем, не импортирован ли уже.
		_, err := s.markets.GetByExternalID(ctx, pm.ID, source)
		if err == nil {
			continue // Уже существует.
		}
		if !errors.Is(err, domain.ErrNotFound) {
			log.Printf("⚠️  Import: ошибка проверки market %s: %v", pm.ID, err)
			continue
		}

		// Парсим outcomes.
		var outcomes []string
		if err := json.Unmarshal([]byte(pm.Outcomes), &outcomes); err != nil {
			outcomes = []string{"Yes", "No"}
		}

		// Парсим deadline.
		deadline, err := time.Parse(time.RFC3339, pm.EndDate)
		if err != nil {
			continue
		}

		if deadline.Before(time.Now()) {
			continue
		}

		// Обрезаем описание.
		desc := pm.Description
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}

		title := pm.Question
		if len(title) > 200 {
			title = title[:200]
		}

		category := "polymarket"
		extID := pm.ID

		market, err := s.marketService.CreateMarket(ctx, title, &desc, outcomes, deadline, &category, adminUserID)
		if err != nil {
			log.Printf("⚠️  Import: ошибка создания рынка '%s': %v", title, err)
			continue
		}

		// Обновляем external_id.
		market.ExternalID = &extID
		market.ExternalSource = &source
		if err := s.markets.Update(ctx, market); err != nil {
			log.Printf("⚠️  Import: ошибка обновления external_id для market %s: %v", market.ID, err)
		}

		imported++
		log.Printf("📥 Imported: %s (poly: %s)", strings.TrimSpace(title), extID)
	}

	return imported, nil
}
