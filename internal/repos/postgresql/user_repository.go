package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/jmoiron/sqlx"
)

// UserRepository реализует работу с таблицей users через sqlx.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository создаёт новый UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create создаёт нового пользователя.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, password_hash, role)
			  VALUES ($1, $2, $3, $4)
			  RETURNING created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, string(user.Role),
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID возвращает пользователя по ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User

	query := `SELECT id, email, password_hash, role, created_at, updated_at
			  FROM users
			  WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByEmail возвращает пользователя по email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	query := `SELECT id, email, password_hash, role, created_at, updated_at
			  FROM users
			  WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update обновляет данные пользователя.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users
			  SET email = $1, password_hash = $2, role = $3
			  WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query,
		user.Email, user.PasswordHash, string(user.Role), user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete удаляет пользователя по ID.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateRole обновляет только роль пользователя.
func (r *UserRepository) UpdateRole(ctx context.Context, id string, role domain.UserRole) error {
	query := `UPDATE users SET role = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, string(role), id)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ListAll возвращает список пользователей с пагинацией и общее количество.
func (r *UserRepository) ListAll(ctx context.Context, page, limit int) ([]domain.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	var users []domain.User
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	if err := r.db.SelectContext(ctx, &users, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}
