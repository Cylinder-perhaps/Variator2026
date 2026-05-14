package domain

import "errors"

// Типизированные доменные ошибки для маппинга в HTTP-коды.
var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrMarketNotActive     = errors.New("market is not active")
	ErrOrderAlreadyFilled  = errors.New("order already filled")
	ErrTokenExpired        = errors.New("token expired")
	ErrTokenRevoked        = errors.New("token revoked")
)
