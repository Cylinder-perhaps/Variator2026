package domain

import (
	"encoding/json"
	"time"
)

// MarketStatus определяет статус рынка.
type MarketStatus string

const (
	MarketStatusActive   MarketStatus = "ACTIVE"
	MarketStatusClosed   MarketStatus = "CLOSED"
	MarketStatusResolved MarketStatus = "RESOLVED"
)

// OutcomesJSON — вспомогательный тип для JSONB массива исходов.
type OutcomesJSON []string

// Scan реализует интерфейс sql.Scanner для чтения JSONB из PostgreSQL.
func (o *OutcomesJSON) Scan(src interface{}) error {
	if src == nil {
		*o = []string{}
		return nil
	}

	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, o)
	case string:
		return json.Unmarshal([]byte(v), o)
	default:
		return json.Unmarshal(src.([]byte), o)
	}
}

// Market представляет рынок предсказаний.
type Market struct {
	ID              string       `db:"id"`
	Title           string       `db:"title"`
	Description     *string      `db:"description"`
	Outcomes        OutcomesJSON `db:"outcomes"`
	Status          MarketStatus `db:"status"`
	Category        *string      `db:"category"`
	Deadline        time.Time    `db:"deadline"`
	ResolvedOutcome *string      `db:"resolved_outcome"`
	EvidenceURL     *string      `db:"evidence_url"`
	CreatedBy       *string      `db:"created_by"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
}
