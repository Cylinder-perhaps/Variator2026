package domain

import "time"

type Trade struct {
	ID         string
	OrderID    string
	UserID     string
	MarketID   string
	Outcome    string
	Quantity   float64
	Price      float64
	PNL        float64
	ExecutedAt time.Time
}
