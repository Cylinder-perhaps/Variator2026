-- +goose Up
-- +goose StatementBegin

-- =============================================
-- MARKETS (события для предсказаний)
-- =============================================
CREATE TABLE markets (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title            VARCHAR(200) NOT NULL,
    description      TEXT,
    outcomes         JSONB NOT NULL DEFAULT '[]'::jsonb,
    status           market_status NOT NULL DEFAULT 'ACTIVE',
    category         VARCHAR(100),
    deadline         TIMESTAMPTZ NOT NULL,
    resolved_outcome VARCHAR(200),
    evidence_url     TEXT,
    created_by       UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_outcomes_not_empty CHECK (jsonb_array_length(outcomes) >= 2),
    CONSTRAINT chk_deadline_future CHECK (deadline > created_at)
);

CREATE INDEX idx_markets_status ON markets (status);
CREATE INDEX idx_markets_deadline ON markets (deadline);
CREATE INDEX idx_markets_created_at ON markets (created_at DESC);
CREATE INDEX idx_markets_category ON markets (category) WHERE category IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS markets;
-- +goose StatementEnd
