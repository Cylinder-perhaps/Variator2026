package domain

import (
	"database/sql/driver"
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

// Value реализует интерфейс driver.Valuer для записи JSONB в PostgreSQL.
func (o OutcomesJSON) Value() (driver.Value, error) {
	if o == nil {
		return "[]", nil
	}
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return string(b), nil
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
	ExternalID      *string      `db:"external_id"`
	ExternalSource  *string      `db:"external_source"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
}
