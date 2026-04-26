-- +goose Up
-- +goose StatementBegin

-- =============================================
-- ORDERS (ордера пользователей)
-- =============================================
CREATE TABLE orders (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    market_id   UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    outcome     VARCHAR(200) NOT NULL,
    quantity    DECIMAL(18,4) NOT NULL,
    price       DECIMAL(18,4) NOT NULL,
    amount_paid DECIMAL(18,2) NOT NULL,
    status      order_status NOT NULL DEFAULT 'PENDING',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_order_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_order_price_range CHECK (price > 0 AND price < 1),
    CONSTRAINT chk_order_amount_positive CHECK (amount_paid > 0)
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_market_id ON orders (market_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_user_status ON orders (user_id, status);
CREATE INDEX idx_orders_market_status ON orders (market_id, status);
CREATE INDEX idx_orders_created_at ON orders (created_at DESC);

-- =============================================
-- POSITIONS (открытые позиции пользователей)
-- =============================================
CREATE TABLE positions (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    market_id     UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    outcome       VARCHAR(200) NOT NULL,
    quantity      DECIMAL(18,4) NOT NULL DEFAULT 0,
    avg_cost      DECIMAL(18,4) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_position_user_market_outcome UNIQUE (user_id, market_id, outcome),
    CONSTRAINT chk_position_quantity_non_negative CHECK (quantity >= 0)
);

CREATE INDEX idx_positions_user_id ON positions (user_id);
CREATE INDEX idx_positions_market_id ON positions (market_id);
CREATE INDEX idx_positions_user_market ON positions (user_id, market_id);

-- =============================================
-- TRADES (история исполненных сделок)
-- =============================================
CREATE TABLE trades (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id     UUID REFERENCES orders(id) ON DELETE SET NULL,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    market_id    UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    outcome      VARCHAR(200) NOT NULL,
    quantity     DECIMAL(18,4) NOT NULL,
    price        DECIMAL(18,4) NOT NULL,
    pnl          DECIMAL(18,2) NOT NULL DEFAULT 0,
    executed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_trade_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_trade_price_range CHECK (price > 0 AND price <= 1)
);

CREATE INDEX idx_trades_user_id ON trades (user_id);
CREATE INDEX idx_trades_market_id ON trades (market_id);
CREATE INDEX idx_trades_user_market ON trades (user_id, market_id);
CREATE INDEX idx_trades_executed_at ON trades (executed_at DESC);
CREATE INDEX idx_trades_order_id ON trades (order_id) WHERE order_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
