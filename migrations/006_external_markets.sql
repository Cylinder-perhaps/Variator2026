-- +goose Up
-- +goose StatementBegin

-- =============================================
-- Добавляем поддержку внешних источников (Polymarket и др.)
-- =============================================
ALTER TABLE markets
    ADD COLUMN external_id     VARCHAR(100),
    ADD COLUMN external_source VARCHAR(50);

CREATE UNIQUE INDEX idx_markets_external
    ON markets (external_source, external_id)
    WHERE external_id IS NOT NULL;

COMMENT ON COLUMN markets.external_id IS 'ID рынка во внешнем источнике (например Polymarket)';
COMMENT ON COLUMN markets.external_source IS 'Источник: polymarket, predictit и т.д.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_markets_external;
ALTER TABLE markets
    DROP COLUMN IF EXISTS external_source,
    DROP COLUMN IF EXISTS external_id;
-- +goose StatementEnd
