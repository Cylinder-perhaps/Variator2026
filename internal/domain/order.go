package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "PENDING"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
)

type Order struct {
	ID         string
	UserID     string
	MarketID   string
	Outcome    string
	Quantity   float64
	Price      float64
	AmountPaid float64
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
