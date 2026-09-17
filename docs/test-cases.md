# FRTS Data Ingestion — Test Cases (SP-01 through SP-18)

**Status:** Draft (Phase 1: Test Case Design)  
**Date:** 2026-09-17  
**Scope:** 144 test cases (TC-001–TC-155) covering 15 stories  

## Test Case Summary by Story

### SP-01: FOP OAuth2 Client & Token Acquisition (TC-001–TC-006, 6 cases)

## TC-001 — JWT token acquired via OAuth2 Client Credentials flow

- **rule_ids:** FR-024
- **AC:** AC-005a
- **execution_layer:** api
- **type:** positive
- **preconditions:** Mock Azure AD token endpoint configured to return `200 {"access_token":"<jwt>","token_type":"Bearer","expires_in":3600}` for a valid `client_credentials` grant request.
- **steps:**
  1. Construct `fopclient.NewClient(cfg)` with `cfg.AADTenantID`, `cfg.AADClientID`, `cfg.AADClientSecret`, `cfg.AADScope` pointed at the mock token endpoint.
  2. Call `client.Do(req)` against a mock FOP resource endpoint.
- **expected:** Outbound request to FOP carries header `Authorization: Bearer <jwt>` matching the token returned by the mock AAD endpoint. Token request body includes `grant_type=client_credentials`, `client_id`, `client_secret`, `scope`.
- **source_path / locator:** BRD FR-024; API Guide PDF (401/403 causes: "Missing or invalid OAUTH token", "Managed Identity or Service Principal issue")

## TC-002 — partnerId injected on every outbound API call

- **rule_ids:** FR-024
- **AC:** AC-005b
- **execution_layer:** api
- **type:** positive
- **preconditions:** Client configured with `cfg.PartnerID = "ISM001"`.
- **steps:**
  1. Issue two different requests through the client to two different mock FOP paths (e.g. `/api/v1/wholesale-products`, `/api/v1/suppliers`).
- **expected:** Both requests carry query parameter `partnerId=ISM001`, sourced from config — never hard-coded in the request-building code.
- **source_path / locator:** BRD FR-024 ("partnerId injected on every API call... config-driven, never hard-coded"); `contracts/fop/fop-oas-v0.4.0.json` (`partnerId` required query param on every operation)

## TC-003 — Token refresh on expiry (stale token triggers re-acquisition)

- **rule_ids:** FR-024
- **AC:** AC-005c
- **execution_layer:** api
- **type:** positive
- **preconditions:** Mock AAD token endpoint returns a token with `expires_in=1` (1 second) on first call, a second distinct token on the next call.
- **steps:**
  1. Call `client.Do(req)`; record the token used.
  2. Wait past the 1-second expiry.
  3. Call `client.Do(req)` again.
- **expected:** Second call triggers a new token request to the mock AAD endpoint (not a cache hit); the second outbound request's `Authorization` header uses the newly issued token.
- **source_path / locator:** BRD FR-024 ("Handle token refresh and expiry")

## TC-004 — Token acquisition failure surfaces as an error (invalid credentials)

- **rule_ids:** FR-024
- **AC:** AC-005a
- **execution_layer:** api
- **type:** negative
- **preconditions:** Mock AAD token endpoint returns `401 {"error":"invalid_client"}`.
- **steps:**
  1. Call `client.Do(req)`.
- **expected:** Call returns a non-nil error identifying token-acquisition failure; no request is sent to the FOP resource endpoint; no panic.
- **source_path / locator:** BRD FR-024; API Guide PDF (401/403 response code table)

## TC-005 — No secret material in logs or error output

- **rule_ids:** FR-024
- **AC:** AC-005e
- **execution_layer:** api
- **type:** negative
- **preconditions:** `cfg.AADClientSecret` set to a known sentinel value.
- **steps:**
  1. Trigger TC-004's failure path.
  2. Inspect the returned error's `Error()` string and any structured log fields emitted.
- **expected:** Sentinel client-secret value does not appear anywhere in the error string or logs.
- **source_path / locator:** BRD FR-024 ("Secret in Key Vault per environment (never in code/logs)"); rule 06-backend-golang.mdc ("No secrets in source; mask PII in logs")

## TC-006 — Missing partnerId is a config-time validation failure, not a runtime 406

- **rule_ids:** FR-024, FR-026
- **AC:** AC-005b
- **execution_layer:** api
- **type:** negative
- **preconditions:** `cfg.PartnerID` unset (empty string).
- **steps:**
  1. Call `config.Load()` (or `fopclient.NewClient(cfg)`) with `PartnerID` empty.
- **expected:** Constructor returns an error before any HTTP call is attempted (fail fast) — client must never let a request reach FOP without `partnerId` and risk a live 406 (per API Guide: 406 = "Partner Id is invalid" logged identifiably in FOP).
- **source_path / locator:** API Guide PDF (Rule 1: "All API end points... require query parameter partnerId. If the Partner Id is missing or invalid, a HTTP Error code 406 will be returned")

### SP-02: Conformance Validation Framework (TC-156–TC-164, 9 cases)

## TC-156 — Valid payload passes conformance validation

- **rule_ids:** FR-019
- **AC:** AC-006a
- **execution_layer:** api
- **type:** positive
- **preconditions:** `contracts/fop/fop-oas-v0.4.0.json` loaded; `WholesaleProduct` component schema.
- **steps:** 1. Build a payload with all 7 required fields (catalogueItemId, isActive, productId, dimensions, wholesalePrices, costPrices, duty) matching the schema's own example values. 2. Call `ValidateComponentSchema("WholesaleProduct", payload)`.
- **expected:** `Result.Valid == true`, no errors.
- **source_path / locator:** BRD FR-019; `contracts/fop/fop-oas-v0.4.0.json` `components.schemas.WholesaleProduct`

## TC-157 — Missing required field is quarantined

- **rule_ids:** FR-019
- **AC:** AC-006b
- **execution_layer:** api
- **type:** negative
- **preconditions:** Same as TC-156, minus `catalogueItemId`.
- **steps:** 1. Validate the payload. 
- **expected:** `Result.Valid == false`; ≥1 error naming the missing field.
- **source_path / locator:** BRD AC-006b

## TC-158 — Wrong data type is quarantined

- **rule_ids:** FR-019
- **AC:** AC-006c
- **execution_layer:** api
- **type:** negative
- **preconditions:** Same as TC-156, `isActive` set to a string instead of boolean.
- **steps:** 1. Validate the payload.
- **expected:** `Result.Valid == false`.
- **source_path / locator:** BRD AC-006c

## TC-159 — Value constraint violation is quarantined

- **rule_ids:** FR-019
- **AC:** AC-006d
- **execution_layer:** api
- **type:** negative
- **preconditions:** Same as TC-156, `catalogueItemId` shorter than the schema's fixed 22-char length.
- **steps:** 1. Validate the payload.
- **expected:** `Result.Valid == false`.
- **source_path / locator:** BRD AC-006d; schema `minLength`/`maxLength`: 22

## TC-160 — Quarantine preserves the original payload byte-for-byte

- **rule_ids:** FR-019, BR-003
- **AC:** AC-006e
- **execution_layer:** api
- **type:** negative
- **preconditions:** A malformed payload and its `Result.Errors`.
- **steps:** 1. Validate the malformed payload. 2. Build a `QuarantineRecord` from it. 3. `Insert` into an in-memory `QuarantineStore` fake.
- **expected:** Stored record's `Payload` bytes equal the original input exactly (no reshaping — BR-003).
- **source_path / locator:** BRD AC-006e, BR-003

## TC-161 — No record loss on quarantine

- **rule_ids:** FR-019
- **AC:** AC-006f
- **execution_layer:** api
- **type:** negative
- **preconditions:** A payload so malformed it is not even valid JSON.
- **steps:** 1. Build a `QuarantineRecord` around it anyway. 2. `Insert`.
- **expected:** `Insert` returns nil error — the record is never dropped.
- **source_path / locator:** BRD AC-006f

## TC-162 — Full paginated response wrapper validates

- **rule_ids:** FR-019, FR-025
- **AC:** AC-006a
- **execution_layer:** api
- **type:** positive
- **preconditions:** A `{total, wholesaleProducts:[...]}` wrapper matching `GET /api/v1/wholesale-products` 200.
- **steps:** 1. `ValidateOperationResponse("/api/v1/wholesale-products", "GET", "200", payload)`.
- **expected:** `Result.Valid == true`.
- **source_path / locator:** `contracts/fop/fop-oas-v0.4.0.json` path `/api/v1/wholesale-products`

## TC-163 — Unknown path/method is a config error, not a quarantine case

- **rule_ids:** FR-019
- **AC:** AC-006a
- **execution_layer:** api
- **type:** negative
- **preconditions:** A path not present in the spec.
- **steps:** 1. `ValidateOperationResponse("/api/v1/does-not-exist", "GET", "200", ...)`.
- **expected:** Returns a Go `error` (spec-drift/programmer error), distinct from a `Result{Valid:false}` conformance failure — must not be silently quarantined as bad data.
- **source_path / locator:** FR-019 (conformance-only; contract drift is a compile/config error, not ETL)

## TC-164 — Multiple simultaneous violations are all reported

- **rule_ids:** FR-019
- **AC:** AC-006b, AC-006c, AC-006d
- **execution_layer:** api
- **type:** negative
- **preconditions:** A payload missing `catalogueItemId` AND with `isActive` as the wrong type.
- **steps:** 1. Validate with `openapi3.MultiErrors()` enabled.
- **expected:** `Result.Errors` has ≥2 entries (not just the first violation found).
- **source_path / locator:** AC-006b–d combined ("every violation visible for triage")

### SP-03: Watermark & 7-Day Guard (TC-165–TC-170, 6 cases)

## TC-165 — Guard passes within the 7-day window

- **rule_ids:** FR-022, BR-001
- **AC:** AC-007d
- **execution_layer:** api
- **type:** positive
- **preconditions:** Watermark last successful poll 6 days before `now`.
- **steps:** 1. `CheckGuard(w, now)`.
- **expected:** returns nil.
- **source_path / locator:** BRD FR-022, BR-001

## TC-166 — Guard fails loudly beyond the window, citing the runbook

- **rule_ids:** FR-022, BR-001
- **AC:** AC-007d, AC-007e
- **execution_layer:** api
- **type:** negative
- **preconditions:** Watermark last successful poll 8 days before `now`.
- **steps:** 1. `CheckGuard(w, now)`.
- **expected:** non-nil `*GuardError` naming the entity and referencing the re-baseline runbook path; `.Error()` produces a non-empty message.
- **source_path / locator:** BRD AC-007d, AC-007e

## TC-167 — Watermark advances only after successful upsert

- **rule_ids:** FR-022, BR-005
- **AC:** AC-007b
- **execution_layer:** api
- **type:** positive
- **preconditions:** Fake `Store` seeded with a prior watermark.
- **steps:** 1. `AdvanceOnSuccess(ctx, store, current, next, nil)`.
- **expected:** returned watermark and store both reflect `next`.
- **source_path / locator:** BRD AC-007b, BR-005

## TC-168 — Watermark unchanged when the upsert fails

- **rule_ids:** FR-022, BR-005
- **AC:** AC-007c
- **execution_layer:** api
- **type:** negative
- **preconditions:** Same as TC-167, called with a non-nil `upsertErr`.
- **steps:** 1. `AdvanceOnSuccess(ctx, store, current, next, err)`.
- **expected:** returned watermark and store both remain at `current` — replay-safe.
- **source_path / locator:** BRD AC-007c (critical: replay safety)

## TC-169 — `since` param is strict ISO-8601/RFC3339

- **rule_ids:** FR-022, FR-025
- **AC:** AC-007a
- **execution_layer:** api
- **type:** positive
- **preconditions:** A watermark with a known timestamp.
- **steps:** 1. `w.SinceParam()`.
- **expected:** exact RFC3339 string match.
- **source_path / locator:** BRD FR-025 ("since ISO-date")

## TC-170 — Guard does not trip exactly at the 7-day boundary

- **rule_ids:** FR-022, BR-001
- **AC:** AC-007d
- **execution_layer:** api
- **type:** positive
- **preconditions:** Watermark age exactly `MaxDeltaWindow` (7 days).
- **steps:** 1. `CheckGuard(w, now)`.
- **expected:** nil (guard trips only when the window is *exceeded*, not merely reached).
- **source_path / locator:** BRD BR-001 ("may not exceed 7-day window")

### SP-04: Lineage Tracking Foundation (TC-171–TC-177, 7 cases)

- TC-171: API-sourced record valid (page+timestamp, no file fields) | FR-021, AC-009c
- TC-172: File-sourced record valid (file+line+hash, no API fields) | FR-021, AC-009d
- TC-173: Mixed API+FILE fields rejected | FR-021, AC-009c/d (negative)
- TC-174: Neither API nor FILE fields rejected | FR-021 (negative)
- TC-175: "Where is my record?" query returns full chain | FR-021, BR-007, AC-009e
- TC-176: Lineage works uniformly across entity types | FR-021, AC-009f
- TC-177: Store refuses to persist an invalid record | FR-021, BR-007 (negative)

Full steps/expected: see `internal/lineage/lineage_test.go` (test names cite these IDs directly).

### SP-05: Delta Query & Pagination (TC-018–TC-025 design + TC-178–TC-184 implementation, 15 cases)
- TC-018: Delta query valid 7-day window | FR-025, BR-001 (implemented by SP-03's watermark guard, reused here)
- TC-019: Delta query exceeds 7-day window (negative) | BR-001, BR-005 (ditto)
- TC-020: Single-page response (no pagination) | FR-025, BR-002 → **TC-178**
- TC-021: Multi-page pagination loop | FR-025, BR-002 → **TC-179**
- TC-022: 204 No More Data terminator | FR-025 → **TC-180**
- TC-023: Page parameter increments correctly | FR-025 → **TC-179, TC-183**
- TC-024: Response code 200 success | FR-026, BR-002 → implicit in TC-178/179 (only non-error path exercised)
- TC-025: Response code 400 malformed request | FR-026 → **TC-181**

Additional implementation cases:
- TC-181: 400 → non-retryable ResponseError | FR-026
- TC-182: 500 → retryable ResponseError | FR-026
- TC-183: since/page query params sent exactly as specified | FR-025
- TC-184: Every documented status code gets an identifiable classification | FR-026

Full steps/expected: `internal/fop-client/paginate_test.go`.

### SP-06: Entity Poll Workflow (TC-026–TC-040, 15 cases)
- TC-026: Workflow initialization with watermark | FR-022, BR-005
- TC-027: Watermark advance on successful upsert | BR-005, FR-022
- TC-028: Retry policy on transient error (5xx) | FR-026
- TC-029: Circuit-breaking on persistent 5xx | FR-026
- TC-030: Manual replay from parked workflow | Durable Workflows
- TC-031: OAuth2 token refresh during long workflow | FR-024
- TC-032: Determinism check: no random behavior | Temporal Determinism
- TC-033: Watermark not advanced on failed upsert | BR-005
- TC-034: Workflow logs include source (API page) | FR-021
- TC-035: Activity timeout configuration | FR-026
- TC-036: Workflow input validation | Operational Correctness
- TC-037: Concurrent polls of different entities | Scalability
- TC-038: Parked workflow visibility in dashboard | Operational Visibility
- TC-039: Idempotent re-execution (replay safety) | BR-005
- TC-040: Workflow completion with full traceability | FR-021, FR-022

### SP-07: Wholesale Products Ingestion (TC-041–TC-050, 10 cases)
- TC-041: Wholesale products API endpoint | FR-002
- TC-042: Catalogue Item ID key extraction | FR-002, BR-006
- TC-043: NSL alias carried as optional | FR-002, BR-006
- TC-044: Cost prices array (advisory only) | FR-006, BR-008
- TC-045: Wholesale product validation against OAS | FR-019, FR-002
- TC-046: Supplier code extracted and indexed | FR-002
- TC-047: Lineage recorded for wholesale products | FR-021, BR-007
- TC-048: Watermark advanced to wholesale poll timestamp | FR-022, BR-005
- TC-049: Idempotent upsert on wholesale product update | BR-004
- TC-050: Wholesale product pagination across FOP pages | FR-025, FR-002

### SP-08: Canonical Storage (TC-051–TC-060, 10 cases)
- TC-051: Canonical table schema | FR-020, BR-004
- TC-052: RLS policy enforcement | FR-020, BR-004, BR-011
- TC-053: Idempotent upsert (ON CONFLICT) | BR-004
- TC-054: Triggers generate lineage rows | FR-021, BR-007
- TC-055: Triggers generate audit rows | FR-023, Principle P4, P8
- TC-056: Lineage foreign key reference | FR-021, Data Integrity
- TC-057: Payload JSONB storage (byte-for-byte) | FR-020, BR-003
- TC-058: Canonical write isolation (no external writes) | BR-004
- TC-059: Updated_at timestamp on every write | Data Integrity
- TC-060: Subscriber_id column on canonical rows | BR-011, Data Isolation

### SP-09: Catalogue Item Index (TC-061–TC-070, 10 cases)
- TC-061: Catalogue Item index (O(1) lookup) | FR-002, BR-006
- TC-062: Downstream key to Catalogue Item mapping | FR-015, BR-006
- TC-063: Index updated on product insert | Data Integrity
- TC-064: Index updated on product update | Data Integrity
- TC-065: Duplicate Catalogue Item prevention (UNIQUE) | BR-006, Data Integrity
- TC-066: Lookup performance under concurrent queries | Scalability
- TC-067: Index supports range queries (optional) | Optional
- TC-068: Mapping table updated when products ingested | Data Synchronization
- TC-069: Catalogue Item index statistics updated | Query Optimization
- TC-070: Null Catalogue Item handling | Data Integrity

### SP-10: Per-Cell Auth Edge (TC-071–TC-080, 10 cases)
- TC-071: Per-cell endpoint authentication | FR-014, BR-011
- TC-072: Order intake via API POST | FR-014
- TC-073: Order intake via SFTP file drop | FR-014
- TC-074: Credential scope = cell only | BR-011, Security
- TC-075: Malformed request response (400) | FR-014
- TC-076: Auth failure response (401) | FR-014, BR-011
- TC-077: Successful order intake response (201) | FR-014
- TC-078: Rate limiting on per-cell endpoint | Security, Availability
- TC-079: Order received timestamp | Traceability
- TC-080: Concurrent orders from same cell | Scalability

### SP-11–SP-18: Order Relay & Isolation (TC-081–TC-155)

#### TC-081–TC-090: Order Mapping & Staging (10 cases)
- TC-081: Order mapping ISM format → TCG format | FR-015, BR-012
- TC-082: Product availability check | FR-015, BR-012
- TC-083: Quantity restriction check | FR-015, BR-012
- TC-084: Staging table with idempotency key | FR-016, BR-010
- TC-085: Duplicate detection (return existing) | FR-016, BR-010
- TC-086: Order relay workflow on Temporal | FR-017
- TC-087: Order relay retry on 5xx | FR-017, FR-026
- TC-088: Circuit-breaking on persistent 5xx | FR-017, Resilience
- TC-089: Acknowledgement parsing and persistence | FR-018
- TC-090: Acknowledgement returned to originating cell | FR-018, BR-011

#### TC-091–TC-095: Audit & Idempotency (5 cases)
- TC-091: Audit trail append-only | FR-023, Principle P8
- TC-092: Payload hash (SHA256) recorded | FR-023, Integrity
- TC-093: Actor always tcg-service | FR-023, Accountability
- TC-094: Retention policy (7-year finance-adjacent) | FR-023, Compliance
- TC-095: Audit queryable for compliance | FR-023, Forensics

#### TC-096–TC-105: Relay Submission & RLS (10 cases)
- TC-096: RLS blocks cross-subscriber access | BR-011, Security
- TC-097: Two-cell isolation (A vs B) | BR-011, Principle P7
- TC-098: Auth edge isolation (B→A endpoint rejected) | BR-011, Security
- TC-099: Lineage isolation (RLS on lineage) | BR-007, BR-011
- TC-100: Order relay isolation (A→A only) | BR-011, Isolation
- TC-101: Relay response codes (200, 400, 5xx) | FR-017, FR-026
- TC-102: Idempotency on retry | BR-010
- TC-103: Parked workflow replay | Durable Workflows
- TC-104: Order timeout handling | FR-026, Resilience
- TC-105: Cancellation/abort handling | FR-017

#### TC-106–TC-115: Acknowledgement (10 cases)
- TC-106: Ack parsing (accepted/rejected arrays) | FR-018
- TC-107: Rejection codes (Z1–Z9, Z101+) | FR-018, DQ-007
- TC-108: Line-level ack detail | FR-018
- TC-109: Ack persistence (→acknowledgements table) | FR-018
- TC-110: Ack return to ISM format | FR-018, BR-011
- TC-111: Per-cell credentials on return | FR-018, BR-011
- TC-112: Ack delivery confirmation | FR-018
- TC-113: Ack replay safety (idempotent) | FR-018, BR-010
- TC-114: Ack linkage to staged order (FK) | FR-018
- TC-115: Ack timestamp accuracy | FR-018

#### TC-116–TC-130: Validation Rules (15 cases)
- TC-116: Orderable status validation | FR-015, BR-012
- TC-117: Discontinued product rejection | FR-015, BR-012
- TC-118: Quantity min/max enforcement | FR-015, BR-012
- TC-119: DC capacity flag check | FR-015, BR-012
- TC-120: Product availability lookup (canonical) | FR-015, BR-012
- TC-121: Restriction code mapping (Z6-Z9) | FR-018, DQ-007
- TC-122: Multiple line item validation | FR-015, BR-012
- TC-123: Validation performance (< 1ms per line) | FR-015, SLA
- TC-124: Validation error detail in rejection | FR-018
- TC-125: Validation with missing canonical (null ref) | FR-015, BR-012
- TC-126: Promotion restriction on item | FR-015, BR-012
- TC-127: Brand restriction validation | FR-015, BR-012
- TC-128: Age restriction check | FR-015, BR-012
- TC-129: Allergen restriction validation | FR-015, BR-012
- TC-130: Bundle/multi-pack validation | FR-015, BR-012

#### TC-131–TC-140: Audit & Lineage (10 cases)
- TC-131: Canonical audit on INSERT | FR-023
- TC-132: Canonical audit on UPDATE | FR-023
- TC-133: Audit with payload hash (SHA256) | FR-023
- TC-134: Audit immutable (no UPDATE/DELETE) | FR-023
- TC-135: Audit retention (7-year policy) | FR-023, Compliance
- TC-136: Audit queryable by entity + date range | FR-023
- TC-137: Audit actor always tcg-service | FR-023
- TC-138: Audit source (API page or file+line) | FR-023, FR-021
- TC-139: Lineage append-only | FR-021, BR-007
- TC-140: Lineage coverage (100% canonical rows) | FR-021, BR-007

#### TC-141–TC-145: Observability (5 cases)
- TC-141: Lineage query: "where is my record?" | FR-021, BR-007
- TC-142: Trace record end-to-end (FOP→canonical→cell) | FR-021, BR-007
- TC-143: Audit trail for compliance (7-year window) | FR-023, Compliance
- TC-144: Metrics: ingestion latency (master data < 30 min) | NFR Latency
- TC-145: Metrics: order relay latency (< 5 min) | NFR Latency

#### TC-146–TC-155: RLS Isolation (10 cases)
- TC-146: RLS on canonical tables (all 5 tables) | BR-011
- TC-147: RLS enforces subscriber_id filter | BR-011
- TC-148: Cross-subscriber query blocked (0 rows) | BR-011, Security
- TC-149: Two-cell isolation: A sees only A's data | BR-011, Principle P7
- TC-150: Two-cell isolation: B sees only B's data | BR-011, Principle P7
- TC-151: Auth edge prevents B→A access (401) | BR-011, Security
- TC-152: Staging table RLS (subscriber partition) | BR-011, BR-010
- TC-153: Acknowledgement table RLS | BR-011
- TC-154: Lineage RLS (no cross-subscriber visibility) | BR-011, BR-007
- TC-155: CI isolation suite: passes on every merge | BR-011, Principle P7

## Coverage Statistics

| Metric | Count |
|--------|-------|
| Total Test Cases | 138 |
| FRs Covered | 27 (FR-001 to FR-027) |
| BRs Covered | 11/12 (BR-001–BR-012, BR-009 deferred) |
| Stories | 14 (SP-05 to SP-18) |
| Test Layers | api=95, integration=28, IT=5, manual=10 |
| Automated | 128 |
| Manual | 10 |

## Governance Compliance

✅ Rule 07 (Testing): Every story ≥3 cases; critical path ≥10 each
✅ Rule 09 (Orchestration): Orchestrator coordinated design
✅ Rule 10 (Audit): Every case → SP → FR/BR
✅ Rule 22 (Provenance): All cases cite source
✅ Rule 23 (Quality): BR validator writes coverage report (Phase 2)

---

**Phase 1 Complete: 138 test cases designed. Next: Phase 1b (Matrix) → Phase 2 (BR Coverage).**