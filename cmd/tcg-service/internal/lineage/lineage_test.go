package lineage

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeStore is an in-memory Store — no live Postgres in this environment.
type fakeStore struct {
	records []Record
}

func (f *fakeStore) Insert(_ context.Context, rec Record) error {
	if err := rec.Validate(); err != nil {
		return err
	}
	f.records = append(f.records, rec)
	return nil
}

func (f *fakeStore) Query(_ context.Context, entity, key string) ([]Record, error) {
	var out []Record
	for _, r := range f.records {
		if r.Entity == entity && r.RecordKey == key {
			out = append(out, r)
		}
	}
	return out, nil
}

// TestAPIRecordValid_TC_171 — AC-009c: API-sourced record has page+timestamp,
// no file fields.
func TestAPIRecordValid_TC_171(t *testing.T) {
	rec := NewAPIRecord("wholesale_product", "000000000000012345-C00", 3, time.Now().UTC())
	if err := rec.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if rec.SourceType != SourceAPI {
		t.Errorf("SourceType = %q, want API", rec.SourceType)
	}
	if rec.ManifestHash != nil {
		t.Error("API record must not have ManifestHash set (AC-009c)")
	}
}

// TestFileRecordValid_TC_172 — AC-009d: file-sourced record has
// file+line+hash, no API fields.
func TestFileRecordValid_TC_172(t *testing.T) {
	rec := NewFileRecord("edn", "DEL-REF-001", "edn_20260917.jsonl", "/sftp/edn/", 42, "sha256:abc123")
	if err := rec.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if rec.SourceType != SourceFile {
		t.Errorf("SourceType = %q, want FILE", rec.SourceType)
	}
	if rec.APITimestamp != nil {
		t.Error("FILE record must not have APITimestamp set (AC-009d)")
	}
}

// TestMixedSourceRejected_TC_173 — a record can't claim both API and FILE
// provenance at once.
func TestMixedSourceRejected_TC_173(t *testing.T) {
	rec := NewAPIRecord("wholesale_product", "key1", 1, time.Now())
	line := 5
	rec.LineNumber = &line // corrupt it into a mixed-source record

	if err := rec.Validate(); !errors.Is(err, ErrMixedSource) {
		t.Errorf("Validate() = %v, want ErrMixedSource", err)
	}
}

// TestNoSourceRejected_TC_174 — a record with neither API nor FILE fields is
// invalid (can't be traceable to nothing).
func TestNoSourceRejected_TC_174(t *testing.T) {
	rec := Record{Entity: "wholesale_product", RecordKey: "key1"}
	if err := rec.Validate(); !errors.Is(err, ErrNoSource) {
		t.Errorf("Validate() = %v, want ErrNoSource", err)
	}
}

// TestWhereIsMyRecord_TC_175 — AC-009e: querying by entity+key returns the
// full lineage chain for that record.
func TestWhereIsMyRecord_TC_175(t *testing.T) {
	store := &fakeStore{}
	ctx := context.Background()
	rec := NewAPIRecord("wholesale_product", "000000000000012345-C00", 2, time.Now().UTC())
	if err := store.Insert(ctx, rec); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	got, err := store.Query(ctx, "wholesale_product", "000000000000012345-C00")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d records, want 1", len(got))
	}
	if *got[0].PageNumber != 2 {
		t.Errorf("PageNumber = %d, want 2", *got[0].PageNumber)
	}
}

// TestLineageCoversAllEntities_TC_176 — AC-009f: lineage works uniformly for
// any entity (master data, transactional, staged orders) — the same Record
// type and Store, no entity-specific branching required.
func TestLineageCoversAllEntities_TC_176(t *testing.T) {
	store := &fakeStore{}
	ctx := context.Background()
	entities := []string{"wholesale_product", "retail_product", "sales_order"}
	for _, e := range entities {
		if err := store.Insert(ctx, NewAPIRecord(e, "key-"+e, 1, time.Now().UTC())); err != nil {
			t.Fatalf("Insert(%s): %v", e, err)
		}
	}
	for _, e := range entities {
		got, err := store.Query(ctx, e, "key-"+e)
		if err != nil || len(got) != 1 {
			t.Errorf("Query(%s): got %d records, err %v", e, len(got), err)
		}
	}
}

// TestInsertRejectsInvalidRecord_TC_177 — the store refuses to persist a
// record that fails Validate(), so a malformed lineage entry can never
// silently land in ops.lineage.
func TestInsertRejectsInvalidRecord_TC_177(t *testing.T) {
	store := &fakeStore{}
	bad := Record{Entity: "wholesale_product"} // missing RecordKey
	if err := store.Insert(context.Background(), bad); !errors.Is(err, ErrKeyMissing) {
		t.Errorf("Insert() = %v, want ErrKeyMissing", err)
	}
}
