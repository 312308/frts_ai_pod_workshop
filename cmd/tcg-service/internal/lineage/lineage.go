// Package lineage maintains record-origin tracking so operators can answer
// "where is my record?" — API page or file+line, never both, never neither.
//
// Source: BRD FR-021 ("Maintain lineage for every persisted record: API
// page source... OR file source... Store in ops.lineage table"); BR-007
// ("All canonical inbound records must be traceable to source").
package lineage

import (
	"errors"
	"time"
)

type SourceType string

const (
	SourceAPI  SourceType = "API"
	SourceFile SourceType = "FILE"
)

// Record is one append-only lineage entry. Exactly one of the API fields
// (PageNumber, APITimestamp) or the FILE fields (FileName, LineNumber,
// ManifestHash) is populated, matching AC-009c/AC-009d — never both, never
// neither.
type Record struct {
	Entity     string
	RecordKey  string
	SourceType SourceType

	PageNumber   *int
	APITimestamp *time.Time

	FileName     *string
	SFTPPath     *string
	LineNumber   *int
	ManifestHash *string

	RecordedAt time.Time
}

// NewAPIRecord builds a lineage Record for an API-sourced upsert (AC-009c).
func NewAPIRecord(entity, key string, page int, apiTimestamp time.Time) Record {
	return Record{
		Entity:       entity,
		RecordKey:    key,
		SourceType:   SourceAPI,
		PageNumber:   &page,
		APITimestamp: &apiTimestamp,
		RecordedAt:   time.Now().UTC(),
	}
}

// NewFileRecord builds a lineage Record for a file-sourced upsert (AC-009d).
func NewFileRecord(entity, key, fileName, sftpPath string, lineNumber int, manifestHash string) Record {
	return Record{
		Entity:       entity,
		RecordKey:    key,
		SourceType:   SourceFile,
		FileName:     &fileName,
		SFTPPath:     &sftpPath,
		LineNumber:   &lineNumber,
		ManifestHash: &manifestHash,
		RecordedAt:   time.Now().UTC(),
	}
}

var (
	ErrMixedSource   = errors.New("lineage: record has both API and FILE fields set")
	ErrNoSource      = errors.New("lineage: record has neither API nor FILE fields set")
	ErrEntityMissing = errors.New("lineage: entity is required")
	ErrKeyMissing    = errors.New("lineage: record key is required")
)

// Validate enforces the "API or FILE, never both, never neither" invariant
// (AC-009c, AC-009d) plus non-empty identity fields.
func (r Record) Validate() error {
	if r.Entity == "" {
		return ErrEntityMissing
	}
	if r.RecordKey == "" {
		return ErrKeyMissing
	}
	hasAPI := r.PageNumber != nil || r.APITimestamp != nil
	hasFile := r.FileName != nil || r.LineNumber != nil || r.ManifestHash != nil
	if hasAPI && hasFile {
		return ErrMixedSource
	}
	if !hasAPI && !hasFile {
		return ErrNoSource
	}
	return nil
}
