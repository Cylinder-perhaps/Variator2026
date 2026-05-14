package domain

import "time"

// Trade представляет исполненную сделку.
type Trade struct {
	ID         string  `db:"id"`
	OrderID    *string `db:"order_id"`
	UserID     string  `db:"user_id"`
	MarketID   string  `db:"market_id"`
	Outcome    string  `db:"outcome"`
	Quantity   float64 `db:"quantity"`
	Price      float64 `db:"price"`
	PNL        float64 `db:"pnl"`
	ExecutedAt time.Time `db:"executed_at"`
}
