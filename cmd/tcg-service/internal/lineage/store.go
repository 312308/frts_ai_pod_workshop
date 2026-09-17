package lineage

import (
	"context"
	"database/sql"
	"fmt"
)

// Store is append-only by design: it exposes Insert and Query only — no
// Update/Delete method exists, so "lineage is append-only" (BR-007) is a
// compile-time property of the interface, not just a runtime convention.
type Store interface {
	Insert(ctx context.Context, rec Record) error
	Query(ctx context.Context, entity, key string) ([]Record, error)
}

// PostgresStore implements Store against ops.lineage (migration
// 003_ops_lineage.up.sql). Untestable without a live Postgres here.
type PostgresStore struct {
	DB *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{DB: db} }

func (s *PostgresStore) Insert(ctx context.Context, rec Record) error {
	if err := rec.Validate(); err != nil {
		return fmt.Errorf("lineage: insert: %w", err)
	}
	const q = `
		INSERT INTO ops.lineage
			(entity, record_key, source_type, page_number, api_timestamp, file_name, sftp_path, line_number, manifest_hash, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := s.DB.ExecContext(ctx, q,
		rec.Entity, rec.RecordKey, rec.SourceType, rec.PageNumber, rec.APITimestamp,
		rec.FileName, rec.SFTPPath, rec.LineNumber, rec.ManifestHash, rec.RecordedAt)
	if err != nil {
		return fmt.Errorf("lineage: insert entity %q key %q: %w", rec.Entity, rec.RecordKey, err)
	}
	return nil
}

// Query answers "where is my record?" (AC-009e) for one entity+key.
func (s *PostgresStore) Query(ctx context.Context, entity, key string) ([]Record, error) {
	const q = `
		SELECT entity, record_key, source_type, page_number, api_timestamp,
		       file_name, sftp_path, line_number, manifest_hash, recorded_at
		FROM ops.lineage
		WHERE entity = $1 AND record_key = $2
		ORDER BY recorded_at`
	rows, err := s.DB.QueryContext(ctx, q, entity, key)
	if err != nil {
		return nil, fmt.Errorf("lineage: query entity %q key %q: %w", entity, key, err)
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.Entity, &r.RecordKey, &r.SourceType, &r.PageNumber, &r.APITimestamp,
			&r.FileName, &r.SFTPPath, &r.LineNumber, &r.ManifestHash, &r.RecordedAt); err != nil {
			return nil, fmt.Errorf("lineage: scan row: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
