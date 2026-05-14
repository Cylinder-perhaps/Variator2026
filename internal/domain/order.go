package domain

import "time"

// OrderStatus определяет статус ордера.
type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "PENDING"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
)

// Order представляет ордер пользователя.
type Order struct {
	ID         string      `db:"id"`
	UserID     string      `db:"user_id"`
	MarketID   string      `db:"market_id"`
	Outcome    string      `db:"outcome"`
	Quantity   float64     `db:"quantity"`
	Price      float64     `db:"price"`
	AmountPaid float64     `db:"amount_paid"`
	Status     OrderStatus `db:"status"`
	CreatedAt  time.Time   `db:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at"`
}
