-- +goose Up
-- +goose StatementBegin

-- =============================================
-- EXTENSIONS
-- =============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================
-- ENUM TYPES
-- =============================================
CREATE TYPE user_role AS ENUM ('user', 'moderator', 'admin');
CREATE TYPE market_status AS ENUM ('ACTIVE', 'CLOSED', 'RESOLVED');
CREATE TYPE order_status AS ENUM ('PENDING', 'FILLED', 'PARTIALLY_FILLED', 'CANCELLED');

-- =============================================
-- USERS
-- =============================================
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email      VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role       user_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);

-- =============================================
-- REFRESH TOKENS (для JWT refresh flow)
-- =============================================
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at) WHERE revoked = false;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS market_status;
DROP TYPE IF EXISTS user_role;
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
