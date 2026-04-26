-- +goose Up
-- +goose StatementBegin

-- =============================================
-- Функция обновления updated_at
-- =============================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_markets_updated_at
    BEFORE UPDATE ON markets FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_balances_updated_at
    BEFORE UPDATE ON balances FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_positions_updated_at
    BEFORE UPDATE ON positions FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_positions_updated_at ON positions;
DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS trg_balances_updated_at ON balances;
DROP TRIGGER IF EXISTS trg_markets_updated_at ON markets;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at();
-- +goose StatementEnd
