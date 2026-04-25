package domain

import "time"

type UserRole string

const (
	RoleUser UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"
	RoleModerator UserRole = "MODERATOR"
)

type User struct{
	ID		string
	Email	string
	PasswordHash string
	Role	UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}