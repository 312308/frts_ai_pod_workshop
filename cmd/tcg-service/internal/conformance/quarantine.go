package conformance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// QuarantineRecord is a malformed payload set aside without data loss
// (AC-006e: "Quarantined record stored in ops.validation_failures with
// original payload for replay").
type QuarantineRecord struct {
	Entity     string
	Payload    []byte
	Errors     []string
	RecordedAt time.Time
}

// QuarantineStore persists quarantined records. An interface so validation
// logic (tested above, no DB) stays decoupled from the Postgres-specific
// implementation (untestable without a live database in this environment).
type QuarantineStore interface {
	Insert(ctx context.Context, rec QuarantineRecord) error
}

// PostgresQuarantineStore writes to ops.validation_failures (migration
// 001_ops_validation_failures.up.sql).
type PostgresQuarantineStore struct {
	DB *sql.DB
}

func NewPostgresQuarantineStore(db *sql.DB) *PostgresQuarantineStore {
	return &PostgresQuarantineStore{DB: db}
}

func (s *PostgresQuarantineStore) Insert(ctx context.Context, rec QuarantineRecord) error {
	errsJSON, err := json.Marshal(rec.Errors)
	if err != nil {
		return fmt.Errorf("conformance: marshal validation errors: %w", err)
	}
	const q = `
		INSERT INTO ops.validation_failures (entity, payload, validation_errors, recorded_at)
		VALUES ($1, $2, $3, $4)`
	if _, err := s.DB.ExecContext(ctx, q, rec.Entity, rec.Payload, errsJSON, rec.RecordedAt); err != nil {
		return fmt.Errorf("conformance: insert quarantine record for entity %q: %w", rec.Entity, err)
	}
	return nil
}
