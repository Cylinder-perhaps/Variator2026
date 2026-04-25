package domain

import "time"

type Position struct {
	ID        string
	UserID    string
	MarketID  string
	Outcome   string
	Quantity  float64
	AvgCost   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
