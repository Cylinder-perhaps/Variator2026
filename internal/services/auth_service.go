package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Cylinder-perhaps/Variator2026/internal/config"
	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService реализует бизнес-логику аутентификации.
type AuthService struct {
	users         repos.UserRepository
	refreshTokens repos.RefreshTokenRepository
	jwtCfg        config.JWTConfig
	authCfg       config.AuthConfig
}

// AuthResponse — ответ аутентификации (register/login).
type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

// NewAuthService создаёт новый AuthService.
func NewAuthService(
	users repos.UserRepository,
	refreshTokens repos.RefreshTokenRepository,
	jwtCfg config.JWTConfig,
	authCfg config.AuthConfig,
) *AuthService {
	return &AuthService{
		users:         users,
		refreshTokens: refreshTokens,
		jwtCfg:        jwtCfg,
		authCfg:       authCfg,
	}
}

// Register регистрирует нового пользователя.
func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResponse, error) {
	// Проверяем, не занят ли email.
	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return nil, domain.ErrConflict
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	// Хешируем пароль.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.authCfg.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
		// Role: 	   domain.RoleAdmin, // TODO: по умолчанию всем юзерам роль "user", админов нужно создавать вручную через БД или отдельный эндпоинт
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токены.
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Login аутентифицирует пользователя по email и паролю.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrUnauthorized
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем пароль.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Генерируем токены.
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Refresh обновляет access-токен по refresh-токену.
func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (string, error) {
	tokenHash := hashToken(rawRefreshToken)

	rt, err := s.refreshTokens.GetByTokenHash(ctx, tokenHash)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrUnauthorized
	}
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	if rt.Revoked {
		return "", domain.ErrTokenRevoked
	}

	if time.Now().After(rt.ExpiresAt) {
		return "", domain.ErrTokenExpired
	}

	user, err := s.users.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

// Logout отзывает все refresh-токены пользователя.
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	if err := s.refreshTokens.RevokeByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}
	return nil
}

// generateAccessToken создаёт JWT access-токен.
func (s *AuthService) generateAccessToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": string(user.Role),
		"exp":  time.Now().Add(s.jwtCfg.AccessTTL).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// createRefreshToken создаёт и сохраняет refresh-токен.
func (s *AuthService) createRefreshToken(ctx context.Context, userID string) (string, error) {
	rawToken := uuid.New().String()
	tokenHash := hashToken(rawToken)

	rt := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.jwtCfg.RefreshTTL),
	}

	if err := s.refreshTokens.Create(ctx, rt); err != nil {
		return "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return rawToken, nil
}

// hashToken создаёт SHA-256 хеш токена.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
