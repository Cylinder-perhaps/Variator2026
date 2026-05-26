package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Cylinder-perhaps/Variator2026/internal/config"
	"github.com/Cylinder-perhaps/Variator2026/internal/domain"
	"github.com/Cylinder-perhaps/Variator2026/internal/handler"
	"github.com/Cylinder-perhaps/Variator2026/internal/repos/mocks"
	"github.com/Cylinder-perhaps/Variator2026/internal/services"
	"github.com/Cylinder-perhaps/Variator2026/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func setupTestApp(userRepo *mocks.MockUserRepository, refreshRepo *mocks.MockRefreshTokenRepository) *handler.App {
	cfg := config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 24 * time.Hour,
		},
		Auth: config.AuthConfig{
			BcryptCost:       4, // Используем минимальную стоимость для тестов
			MaxLoginAttempts: 10,
		},
	}

	authService := services.NewAuthService(userRepo, refreshRepo, cfg.JWT, cfg.Auth)

	return &handler.App{
		Auth: authService,
	}
}

func TestRegisterUser_Success(t *testing.T) {
	// TC-01: Успешная регистрация
	mockUserRepo := new(mocks.MockUserRepository)
	mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
	app := setupTestApp(mockUserRepo, mockRefreshRepo)

	reqBody := api.RegisterRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ожидаем, что пользователя с таким email нет
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, domain.ErrUserNotFound)
	// Ожидаем создание пользователя
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	// Ожидаем сохранение refresh-токена
	mockRefreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	app.RegisterUser(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUser_EmailExists(t *testing.T) {
	// TC-02: Регистрация с уже существующим email
	mockUserRepo := new(mocks.MockUserRepository)
	mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
	app := setupTestApp(mockUserRepo, mockRefreshRepo)

	reqBody := api.RegisterRequest{
		Email:    "existing@example.com",
		Password: "Password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	existingUser := &domain.User{Email: "existing@example.com"}
	mockUserRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

	app.RegisterUser(w, req)

	// Согласно ТЗ или логике, должно быть 409 Conflict
	assert.Equal(t, http.StatusConflict, w.Code)
	mockUserRepo.AssertExpectations(t)
}

func TestLoginUser_Success(t *testing.T) {
	// TC-03: Успешный вход пользователя
	mockUserRepo := new(mocks.MockUserRepository)
	mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
	app := setupTestApp(mockUserRepo, mockRefreshRepo)

	password := "Password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 4)

	existingUser := &domain.User{
		ID:           "user-1",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	reqBody := api.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)
	mockRefreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	app.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.AuthResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	mockUserRepo.AssertExpectations(t)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	// TC-04: Вход пользователя с неверным паролем
	mockUserRepo := new(mocks.MockUserRepository)
	mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
	app := setupTestApp(mockUserRepo, mockRefreshRepo)

	password := "Password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 4)

	existingUser := &domain.User{
		ID:           "user-1",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	reqBody := api.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	app.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUserRepo.AssertExpectations(t)
}
