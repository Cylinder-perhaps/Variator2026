-- +goose Up
-- +goose StatementBegin

-- =============================================
-- SEED DEFAULT ADMIN USER
-- =============================================
-- Вставляем админа, если email 'admin@variator.io' еще не существует.
-- Пароль "Admin123!" хешируется с использованием встроенной в pgcrypto функции crypt с параметром gen_salt('bf', 10), 
-- что соответствует стоимости bcrypt = 10, используемой в AuthConfig по умолчанию (BcryptCost: 10)
INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
VALUES (
    '10000000-0000-0000-0000-000000000000',
    'admin@variator.io',
    crypt('Admin123!', gen_salt('bf', 10)),
    'admin',
    now(),
    now()
)
ON CONFLICT (email) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE email = 'admin@variator.io' AND role = 'admin';
-- +goose StatementEnd
