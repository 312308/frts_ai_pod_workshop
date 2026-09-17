-- SP-03: ops.watermark — per-entity delta-query watermark + 7-day guard
-- (BRD FR-022, BR-001, BR-005).

CREATE TABLE ops.watermark (
    entity              TEXT PRIMARY KEY,
    last_successful_poll TIMESTAMPTZ NOT NULL,
    last_api_timestamp  TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
