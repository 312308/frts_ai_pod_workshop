package watermark

import (
	"context"
	"database/sql"
	"fmt"
)

// Store persists and retrieves per-entity watermarks.
type Store interface {
	Get(ctx context.Context, entity string) (Watermark, error)
	// Advance updates the watermark. Callers must only invoke this after a
	// poll cycle's upserts have all succeeded (BR-005) — Store itself does
	// not enforce that; see AdvanceOnSuccess.
	Advance(ctx context.Context, w Watermark) error
}

// AdvanceOnSuccess is the single call site that encodes BR-005: the
// watermark is written only when upsertErr is nil. On failure, the store is
// left untouched so the next run replays from the same watermark
// (AC-007c). Returns the watermark actually in effect after this call.
func AdvanceOnSuccess(ctx context.Context, store Store, current Watermark, next Watermark, upsertErr error) (Watermark, error) {
	if upsertErr != nil {
		return current, nil // unchanged; replay on next run
	}
	if err := store.Advance(ctx, next); err != nil {
		return current, fmt.Errorf("watermark: advance entity %q: %w", next.Entity, err)
	}
	return next, nil
}

// PostgresStore implements Store against ops.watermark (migration
// 002_ops_watermark.up.sql). Untestable without a live Postgres in this
// environment — logic that can be unit-tested lives in AdvanceOnSuccess and
// CheckGuard above, independent of this type.
type PostgresStore struct {
	DB *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{DB: db} }

func (s *PostgresStore) Get(ctx context.Context, entity string) (Watermark, error) {
	var w Watermark
	w.Entity = entity
	const q = `SELECT last_successful_poll, last_api_timestamp FROM ops.watermark WHERE entity = $1`
	err := s.DB.QueryRowContext(ctx, q, entity).Scan(&w.LastSuccessfulPoll, &w.LastAPITimestamp)
	if err != nil {
		return Watermark{}, fmt.Errorf("watermark: get entity %q: %w", entity, err)
	}
	return w, nil
}

func (s *PostgresStore) Advance(ctx context.Context, w Watermark) error {
	const q = `
		INSERT INTO ops.watermark (entity, last_successful_poll, last_api_timestamp, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (entity) DO UPDATE SET
			last_successful_poll = EXCLUDED.last_successful_poll,
			last_api_timestamp = EXCLUDED.last_api_timestamp,
			updated_at = now()`
	if _, err := s.DB.ExecContext(ctx, q, w.Entity, w.LastSuccessfulPoll, w.LastAPITimestamp); err != nil {
		return fmt.Errorf("watermark: advance entity %q: %w", w.Entity, err)
	}
	return nil
}
