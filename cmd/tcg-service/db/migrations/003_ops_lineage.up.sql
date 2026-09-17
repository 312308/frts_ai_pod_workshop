-- SP-04: ops.lineage — append-only record-origin tracking (BRD FR-021, BR-007).

CREATE TABLE ops.lineage (
    id                 BIGSERIAL PRIMARY KEY,
    entity             TEXT NOT NULL,
    record_key         TEXT NOT NULL,
    source_type        TEXT NOT NULL CHECK (source_type IN ('API', 'FILE')),
    page_number        INTEGER,
    api_timestamp      TIMESTAMPTZ,
    file_name          TEXT,
    sftp_path          TEXT,
    line_number        INTEGER,
    manifest_hash      TEXT,
    recorded_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lineage_entity_key ON ops.lineage (entity, record_key);
