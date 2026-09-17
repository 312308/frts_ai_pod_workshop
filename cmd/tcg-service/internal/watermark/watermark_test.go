package watermark

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeStore is an in-memory Store — no live Postgres in this environment.
type fakeStore struct {
	saved map[string]Watermark
}

func newFakeStore() *fakeStore { return &fakeStore{saved: map[string]Watermark{}} }

func (f *fakeStore) Get(_ context.Context, entity string) (Watermark, error) {
	w, ok := f.saved[entity]
	if !ok {
		return Watermark{}, errors.New("not found")
	}
	return w, nil
}

func (f *fakeStore) Advance(_ context.Context, w Watermark) error {
	f.saved[w.Entity] = w
	return nil
}

// TestGuardPassesWithinWindow_TC_165 — AC-007d positive: within 7 days, no error.
func TestGuardPassesWithinWindow_TC_165(t *testing.T) {
	now := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	w := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: now.Add(-6 * 24 * time.Hour)}
	if err := CheckGuard(w, now); err != nil {
		t.Errorf("expected nil (within window), got %v", err)
	}
}

// TestGuardFailsLoudlyBeyondWindow_TC_166 — AC-007d/e negative: >7 days is a
// loud, identifiable failure citing the re-baseline runbook.
func TestGuardFailsLoudlyBeyondWindow_TC_166(t *testing.T) {
	now := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	w := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: now.Add(-8 * 24 * time.Hour)}
	err := CheckGuard(w, now)
	if err == nil {
		t.Fatal("expected guard error, got nil")
	}
	var guardErr *GuardError
	if !errors.As(err, &guardErr) {
		t.Fatalf("expected *GuardError, got %T", err)
	}
	if guardErr.Entity != "wholesale_product" {
		t.Errorf("GuardError.Entity = %q, want wholesale_product", guardErr.Entity)
	}
	if guardErr.Runbook == "" {
		t.Error("GuardError must reference a re-baseline runbook (AC-007e)")
	}
	if msg := err.Error(); msg == "" {
		t.Error("GuardError.Error() must produce a non-empty, identifiable message")
	}
}

// TestAdvanceOnSuccess_TC_167 — BR-005 positive: watermark advances only
// after a successful upsert.
func TestAdvanceOnSuccess_TC_167(t *testing.T) {
	store := newFakeStore()
	current := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)}
	next := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)}

	got, err := AdvanceOnSuccess(context.Background(), store, current, next, nil)
	if err != nil {
		t.Fatalf("AdvanceOnSuccess: %v", err)
	}
	if !got.LastSuccessfulPoll.Equal(next.LastSuccessfulPoll) {
		t.Errorf("watermark not advanced: got %v, want %v", got.LastSuccessfulPoll, next.LastSuccessfulPoll)
	}
	saved, _ := store.Get(context.Background(), "wholesale_product")
	if !saved.LastSuccessfulPoll.Equal(next.LastSuccessfulPoll) {
		t.Error("store was not updated to the new watermark")
	}
}

// TestWatermarkUnchangedOnFailure_TC_168 — BR-005/AC-007c negative: a failed
// upsert must leave the watermark exactly as it was, for replay.
func TestWatermarkUnchangedOnFailure_TC_168(t *testing.T) {
	store := newFakeStore()
	current := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)}
	next := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)}
	_ = store.Advance(context.Background(), current) // seed prior state

	got, err := AdvanceOnSuccess(context.Background(), store, current, next, errors.New("upsert failed"))
	if err != nil {
		t.Fatalf("AdvanceOnSuccess: %v", err)
	}
	if !got.LastSuccessfulPoll.Equal(current.LastSuccessfulPoll) {
		t.Errorf("watermark changed on failure: got %v, want unchanged %v", got.LastSuccessfulPoll, current.LastSuccessfulPoll)
	}
	saved, _ := store.Get(context.Background(), "wholesale_product")
	if !saved.LastSuccessfulPoll.Equal(current.LastSuccessfulPoll) {
		t.Error("store was mutated despite upsert failure — replay safety broken")
	}
}

// TestSinceParamFormat_TC_169 — the `since` value sent to FOP is a strict
// ISO-8601/RFC3339 timestamp (BRD FR-025 "since ISO-date").
func TestSinceParamFormat_TC_169(t *testing.T) {
	w := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: time.Date(2026, 9, 10, 12, 30, 0, 0, time.UTC)}
	got := w.SinceParam()
	want := "2026-09-10T12:30:00Z"
	if got != want {
		t.Errorf("SinceParam() = %q, want %q", got, want)
	}
}

// TestGuardExactlyAtBoundary_TC_170 — exactly 7 days old must still pass
// (guard trips only when the window is exceeded, not merely reached).
func TestGuardExactlyAtBoundary_TC_170(t *testing.T) {
	now := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
	w := Watermark{Entity: "wholesale_product", LastSuccessfulPoll: now.Add(-MaxDeltaWindow)}
	if err := CheckGuard(w, now); err != nil {
		t.Errorf("expected nil at exactly 7 days, got %v", err)
	}
}
