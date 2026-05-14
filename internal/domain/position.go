package domain

import "time"

// Position представляет открытую позицию пользователя.
type Position struct {
	ID        string  `db:"id"`
	UserID    string  `db:"user_id"`
	MarketID  string  `db:"market_id"`
	Outcome   string  `db:"outcome"`
	Quantity  float64 `db:"quantity"`
	AvgCost   float64 `db:"avg_cost"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
