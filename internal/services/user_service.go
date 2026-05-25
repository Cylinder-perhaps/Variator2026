package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
)

// UserService реализует бизнес-логику управления пользователями.
type UserService struct {
	users repos.UserRepository
}

// NewUserService создаёт новый UserService.
func NewUserService(users repos.UserRepository) *UserService {
	return &UserService{
		users: users,
	}
}

// UsersListResult — результат списка пользователей с пагинацией.
type UsersListResult struct {
	Items []domain.User
	Total int
}

// ListUsers возвращает список пользователей (с пагинацией).
func (s *UserService) ListUsers(ctx context.Context, page, limit int) (*UsersListResult, error) {
	users, total, err := s.users.ListAll(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return &UsersListResult{
		Items: users,
		Total: total,
	}, nil
}

// UpdateUserRole изменяет роль пользователя.
func (s *UserService) UpdateUserRole(ctx context.Context, adminID string, targetUserID string, newRole domain.UserRole) error {
	// Нельзя менять роль самому себе.
	if adminID == targetUserID {
		return fmt.Errorf("%w: нельзя изменить роль самому себе", domain.ErrInvalidRequest)
	}

	// Только роль "moderator" и "user" можно назначать через API (admin назначается вручную или через миграцию).
	if newRole != domain.RoleUser && newRole != domain.RoleModerator {
		return fmt.Errorf("%w: недопустимая роль (только user или moderator)", domain.ErrInvalidRequest)
	}

	// Проверяем существование пользователя.
	user, err := s.users.GetByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Запрещаем изменять роль других админов через API для безопасности.
	if user.Role == domain.RoleAdmin {
		return fmt.Errorf("%w: нельзя изменить роль администратора", domain.ErrForbidden)
	}

	if err := s.users.UpdateRole(ctx, targetUserID, newRole); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}
