// Package watermark maintains the per-entity delta-query watermark and
// 7-day guard.
//
// Source: BRD FR-022 ("Maintain per-entity watermark... Advance watermark
// only on successful upsert completion. Fail loudly... if watermark exceeds
// 7-day window"); BR-001 ("Delta API queries may not exceed 7-day window");
// BR-005 ("Watermarks advance only on successful upsert completion. Failed
// activities leave watermark unchanged; replay on next run.").
package watermark

import (
	"fmt"
	"time"
)

// MaxDeltaWindow is the maximum age a watermark may reach before the guard
// trips (BR-001: 7 days).
const MaxDeltaWindow = 7 * 24 * time.Hour

// Watermark is the delta-query marker for one entity.
type Watermark struct {
	Entity             string
	LastSuccessfulPoll time.Time
	LastAPITimestamp   time.Time
}

// SinceParam returns the ISO-8601 `since` value to use for the next delta
// query against FOP.
func (w Watermark) SinceParam() string {
	return w.LastSuccessfulPoll.UTC().Format(time.RFC3339)
}

// GuardError identifies a 7-day-window breach so callers can alert loudly
// and reference the re-baseline runbook (AC-007d, AC-007e), rather than
// treating it as an ordinary error.
type GuardError struct {
	Entity  string
	Age     time.Duration
	Runbook string
}

func (e *GuardError) Error() string {
	return fmt.Sprintf(
		"watermark: entity %q last successful poll is %s old, exceeds 7-day guard — re-baseline required, see %s",
		e.Entity, e.Age.Round(time.Minute), e.Runbook,
	)
}

// DefaultRunbookRef is the doc link surfaced in GuardError (AC-007e).
const DefaultRunbookRef = "iac/watermark-rebase-runbook.md"

// CheckGuard returns a *GuardError if now - w.LastSuccessfulPoll exceeds
// MaxDeltaWindow (BR-001, AC-007d). A nil return means the watermark is
// within the safe delta-query window.
func CheckGuard(w Watermark, now time.Time) error {
	age := now.Sub(w.LastSuccessfulPoll)
	if age > MaxDeltaWindow {
		return &GuardError{Entity: w.Entity, Age: age, Runbook: DefaultRunbookRef}
	}
	return nil
}
