-- +goose Up
-- +goose StatementBegin

-- =============================================
-- BALANCES (счёт пользователя)
-- =============================================
CREATE TABLE balances (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    total            DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    available        DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    blocked_in_orders DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_total_non_negative CHECK (total >= 0),
    CONSTRAINT chk_available_non_negative CHECK (available >= 0),
    CONSTRAINT chk_blocked_non_negative CHECK (blocked_in_orders >= 0),
    CONSTRAINT chk_total_equals_sum CHECK (total = available + blocked_in_orders)
);

CREATE INDEX idx_balances_user_id ON balances (user_id);

-- Автоматически создаём баланс при регистрации пользователя
-- Начальный баланс — 1000.00 (тестовый депозит для MVP)
CREATE OR REPLACE FUNCTION create_user_balance()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO balances (user_id, total, available, blocked_in_orders)
    VALUES (NEW.id, 1000.00, 1000.00, 0.00);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_create_user_balance
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION create_user_balance();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_create_user_balance ON users;
DROP FUNCTION IF EXISTS create_user_balance();
DROP TABLE IF EXISTS balances;
-- +goose StatementEnd
