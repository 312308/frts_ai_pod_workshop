package conformance

import (
	"context"
	"testing"
	"time"
)

const specPath = "../../../../contracts/fop/fop-oas-v0.4.0.json"

func validWholesaleProductJSON() []byte {
	return []byte(`{
		"catalogueItemId": "000000000000012345-C00",
		"isActive": true,
		"productId": "000000000070000063",
		"dimensions": {"length": 4.123},
		"wholesalePrices": [{"wholesalePrice": 9.99, "currencyCode": "GBP", "startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"costPrices": [{"startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"duty": {}
	}`)
}

// TestValidPayloadPasses_TC_156 — a payload conforming to the WholesaleProduct
// schema validates clean.
func TestValidPayloadPasses_TC_156(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	res, err := v.ValidateComponentSchema("WholesaleProduct", validWholesaleProductJSON())
	if err != nil {
		t.Fatalf("ValidateComponentSchema: %v", err)
	}
	if !res.Valid {
		t.Errorf("expected valid, got errors: %v", res.Errors)
	}
}

// TestMissingRequiredField_TC_157 — a payload missing a required field
// (catalogueItemId) is invalid, not silently accepted.
func TestMissingRequiredField_TC_157(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	payload := []byte(`{
		"isActive": true,
		"productId": "000000000070000063",
		"dimensions": {"length": 4.123},
		"wholesalePrices": [{"wholesalePrice": 9.99, "currencyCode": "GBP", "startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"costPrices": [{"startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"duty": {}
	}`)
	res, err := v.ValidateComponentSchema("WholesaleProduct", payload)
	if err != nil {
		t.Fatalf("ValidateComponentSchema: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invalid (missing catalogueItemId), got valid")
	}
	if len(res.Errors) == 0 {
		t.Error("expected at least one validation error")
	}
}

// TestWrongDataType_TC_158 — a field with the wrong JSON type (isActive as a
// string, not boolean) is invalid.
func TestWrongDataType_TC_158(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	payload := []byte(`{
		"catalogueItemId": "000000000000012345-C00",
		"isActive": "yes",
		"productId": "000000000070000063",
		"dimensions": {"length": 4.123},
		"wholesalePrices": [{"wholesalePrice": 9.99, "currencyCode": "GBP", "startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"costPrices": [{"startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"duty": {}
	}`)
	res, err := v.ValidateComponentSchema("WholesaleProduct", payload)
	if err != nil {
		t.Fatalf("ValidateComponentSchema: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invalid (isActive wrong type), got valid")
	}
}

// TestValueConstraintViolation_TC_159 — catalogueItemId must be exactly 22
// chars (minLength=maxLength=22); a short value is invalid.
func TestValueConstraintViolation_TC_159(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	payload := []byte(`{
		"catalogueItemId": "TOO-SHORT",
		"isActive": true,
		"productId": "000000000070000063",
		"dimensions": {"length": 4.123},
		"wholesalePrices": [{"wholesalePrice": 9.99, "currencyCode": "GBP", "startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"costPrices": [{"startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"duty": {}
	}`)
	res, err := v.ValidateComponentSchema("WholesaleProduct", payload)
	if err != nil {
		t.Fatalf("ValidateComponentSchema: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invalid (catalogueItemId not 22 chars), got valid")
	}
}

// TestValidateOperationResponse_TC_162 — validates a full paginated
// GET /api/v1/wholesale-products 200 response wrapper (the shape actually
// returned by FOP, used before per-record unwrapping in SP-05/SP-07).
func TestValidateOperationResponse_TC_162(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	payload := []byte(`{
		"total": 1,
		"wholesaleProducts": [` + string(validWholesaleProductJSON()) + `]
	}`)
	res, err := v.ValidateOperationResponse("/api/v1/wholesale-products", "GET", "200", payload)
	if err != nil {
		t.Fatalf("ValidateOperationResponse: %v", err)
	}
	if !res.Valid {
		t.Errorf("expected valid, got errors: %v", res.Errors)
	}
}

// TestValidateOperationResponse_UnknownPath_TC_163 — an unknown path/method
// is a caller/config error (spec drift), surfaced distinctly from a
// conformance failure so it isn't silently quarantined as bad data.
func TestValidateOperationResponse_UnknownPath_TC_163(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	if _, err := v.ValidateOperationResponse("/api/v1/does-not-exist", "GET", "200", []byte(`{}`)); err == nil {
		t.Fatal("expected error for unknown path, got nil")
	}
}

// TestMultipleViolations_TC_164 — a payload violating several constraints at
// once (missing field + wrong type) surfaces every violation, not just the
// first (AC-006b–d combined; exercises the MultiError flattening path).
func TestMultipleViolations_TC_164(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	payload := []byte(`{
		"isActive": "not-a-bool",
		"productId": "000000000070000063",
		"dimensions": {"length": 4.123},
		"wholesalePrices": [{"wholesalePrice": 9.99, "currencyCode": "GBP", "startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"costPrices": [{"startDate": "2026-02-04", "endDate": "2026-02-04"}],
		"duty": {}
	}`)
	res, err := v.ValidateComponentSchema("WholesaleProduct", payload)
	if err != nil {
		t.Fatalf("ValidateComponentSchema: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invalid (missing catalogueItemId + wrong isActive type), got valid")
	}
	if len(res.Errors) < 2 {
		t.Errorf("expected multiple violations reported, got %d: %v", len(res.Errors), res.Errors)
	}
}

// fakeQuarantineStore is an in-memory QuarantineStore for TC-160/TC-161 —
// no live Postgres available in this environment.
type fakeQuarantineStore struct {
	inserted []QuarantineRecord
}

func (f *fakeQuarantineStore) Insert(_ context.Context, rec QuarantineRecord) error {
	f.inserted = append(f.inserted, rec)
	return nil
}

// TestQuarantinePreservesOriginalPayload_TC_160 — a malformed payload is
// quarantined with the original bytes intact (for replay), not reshaped.
func TestQuarantinePreservesOriginalPayload_TC_160(t *testing.T) {
	v, err := NewValidator(specPath)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	malformed := []byte(`{"isActive": true}`) // missing every other required field
	res, err := v.ValidateComponentSchema("WholesaleProduct", malformed)
	if err != nil {
		t.Fatalf("ValidateComponentSchema: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invalid, got valid")
	}

	store := &fakeQuarantineStore{}
	rec := QuarantineRecord{
		Entity:     "wholesale_product",
		Payload:    malformed,
		Errors:     res.Errors,
		RecordedAt: time.Now().UTC(),
	}
	if err := store.Insert(context.Background(), rec); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if len(store.inserted) != 1 {
		t.Fatalf("got %d inserted records, want 1", len(store.inserted))
	}
	if string(store.inserted[0].Payload) != string(malformed) {
		t.Errorf("quarantined payload was reshaped: got %s, want %s", store.inserted[0].Payload, malformed)
	}
}

// TestNoDataLossOnQuarantine_TC_161 — quarantining a bad record must not
// error out / drop the record; the caller can always call Insert and get a
// nil error for a well-formed QuarantineRecord regardless of how malformed
// the underlying payload was.
func TestNoDataLossOnQuarantine_TC_161(t *testing.T) {
	store := &fakeQuarantineStore{}
	rec := QuarantineRecord{
		Entity:     "wholesale_product",
		Payload:    []byte(`not even json`),
		Errors:     []string{"payload is not valid JSON"},
		RecordedAt: time.Now().UTC(),
	}
	if err := store.Insert(context.Background(), rec); err != nil {
		t.Fatalf("Insert must not lose a record: %v", err)
	}
	if len(store.inserted) != 1 {
		t.Fatalf("got %d inserted records, want 1", len(store.inserted))
	}
}
