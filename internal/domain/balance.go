package domain

import "time"

// Balance представляет баланс счёта пользователя.
type Balance struct {
	ID              string  `db:"id"`
	UserID          string  `db:"user_id"`
	Total           float64 `db:"total"`
	Available       float64 `db:"available"`
	BlockedInOrders float64 `db:"blocked_in_orders"`
	UpdatedAt       time.Time `db:"updated_at"`
}
