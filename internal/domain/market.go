package domain

import "time"

type MarketStatus string

const (
	MarketStatusActive   MarketStatus = "ACTIVE"
	MarketStatusClosed   MarketStatus = "CLOSED"
	MarketStatusResolved MarketStatus = "RESOLVED"
)

type Market struct {
	ID          string
	Title       string
	Description string
	Category    string
	Status      MarketStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
