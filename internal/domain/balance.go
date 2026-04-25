package domain

import (
	"time"
)

type Balance struct {
	UserID    string
	Amount    float64
	Currency  string
	UpdatedAt time.Time
	CreatedAt time.Time
}
