package domain

import "time"

// UserRole определяет роль пользователя в системе.
type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleAdmin     UserRole = "admin"
	RoleModerator UserRole = "moderator"
)

// User представляет пользователя в системе.
type User struct {
	ID           string   `db:"id"`
	Email        string   `db:"email"`
	PasswordHash string   `db:"password_hash"`
	Role         UserRole `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}