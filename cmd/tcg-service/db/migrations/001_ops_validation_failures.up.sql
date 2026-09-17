-- SP-02: ops.validation_failures — quarantine store for conformance
-- failures (BRD FR-019 AC-006e). Additive-only per stack-conventions-golang-react.md.

CREATE SCHEMA IF NOT EXISTS ops;

CREATE TABLE ops.validation_failures (
    id                BIGSERIAL PRIMARY KEY,
    entity            TEXT NOT NULL,
    payload           JSONB NOT NULL,
    validation_errors JSONB NOT NULL,
    recorded_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_validation_failures_entity_recorded
    ON ops.validation_failures (entity, recorded_at);
