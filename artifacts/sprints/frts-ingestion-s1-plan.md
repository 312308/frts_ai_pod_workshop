# FRTS Data Ingestion Layer — Sprint Plan (frts-ingestion-s1)

**Status:** ✅ **Approved** (2026-09-16 by stakeholder)

**Sprint ID:** `frts-ingestion-s1`  
**Sprint Goal:** Walking Skeleton Stage 1 — Prototype Wholesale Product E2E + Order Relay + Lineage  
**Sprint Duration:** 2 weeks (14 calendar days: 2026-09-16 to 2026-09-30)  
**Module:** FRTS Data Ingestion Layer (tcg-service)  
**Base Directory:** `C:\frts\AI-COE-Workshop-Starter-Kit\starter-kits\claude`  
**Source:** BRD `artifacts/brd/data-ingestion-layer-BRD-draft.md` (Status: ✅ Accepted, 2026-09-16)

**BRD Scope Mapped:**
- 27 Functional Requirements (FR-001 through FR-027)
- 12 Business Rules (BR-001 through BR-012)
- 10 Non-Functional Requirements

---

## 1. Sprint Backlog (18 implement-api Stories)

All stories tagged `implement-api` for agent routing. No UI stories in scope; all backend integration & workflow.

### Dependency Order: Foundation → Integration → Master Data → Order Relay → Testing

---

### **FOUNDATION LAYER (Stories 1–4)**

#### **SP-01: Implement FOP OAuth2 Client & Token Acquisition**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-024, BR-001
- **Description:** Generate FOP client wrapper from committed OAS 3.0 spec using oapi-codegen. Implement OAuth2 Client Credentials flow. Acquire JWT token from TCG per environment. Inject partnerId on every API call. Handle token refresh and expiry.
- **Acceptance Criteria:**
  - AC-005a: JWT token acquired via OAuth2 Client Credentials
  - AC-005b: partnerId injected on all outbound API calls (config-driven, never hard-coded)
  - AC-005c: Token refresh on expiry; stale token triggers re-acquisition
  - AC-005d: Declared-egress-IP routing honored (integration subnet NAT)
  - AC-005e: Secret in Key Vault per environment (never in code/logs)
- **Oracle:** Unit test (mock OAuth2 token endpoint), Contract test (oapi-codegen validates generated client), Integration test (token flow round-trip)
- **Depends On:** — (Foundation; no upstream dependencies)
- **Blocks:** SP-05 (Delta Query handler), SP-07 (Wholesale Products API)
- **Status:** **Ready** (no blockers)
- **Notes:** Mocks-first approach (FOP mock in compose stack); swap to real endpoints at week 6. Use `contracts/fop-oas-v0.4.0.yaml` as committed spec.

---

#### **SP-02: Implement Conformance Validation Framework**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-019, BR-003
- **Description:** Implement JSON Schema validation for all inbound payloads against committed FOP OAS 3.0 spec and FRTS Data Dictionary. Validate required fields, data types, value constraints (e.g., partnerId format). Quarantine malformed payloads in ops.validation_failures table without data loss.
- **Acceptance Criteria:**
  - AC-006a: Inbound payload validated against OAS 3.0 JSON Schema
  - AC-006b: Required fields checked; missing field → quarantine
  - AC-006c: Data types checked; type mismatch → quarantine
  - AC-006d: Value constraints checked (e.g., partnerId format, response code precedence)
  - AC-006e: Quarantined record stored in ops.validation_failures with original payload for replay
  - AC-006f: No record loss; quarantine visible in ops dashboard
- **Oracle:** Unit test (JSON Schema validation), Golden-file test (fixture → expected quarantine row), Integration test (validate malformed payload), Contract test (oapi-codegen conformance)
- **Depends On:** — (Foundation)
- **Blocks:** SP-07 (Wholesale API), SP-08 (Canonical store), All master data ingestion stories
- **Status:** **Ready**
- **Notes:** Supports contract churn as compile error (oapi-codegen + generated validation structs in contracts/ directory). Re-validation after spec upgrade is a query, not migration.

---

#### **SP-03: Implement Watermark & 7-Day Guard**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-022, BR-001, BR-005
- **Description:** Maintain per-entity watermark: last_successful_poll timestamp, last_api_timestamp, 7-day guard check. Advance watermark only on successful upsert completion. Fail loudly (alert + runbook reference) if watermark exceeds 7-day window. Prevent silent gaps. Implement re-baseline runbook reference (weekly full SFTP feed or single-record endpoints).
- **Acceptance Criteria:**
  - AC-007a: Watermark table: (entity, last_successful_poll, last_api_timestamp, 7day_guard_check)
  - AC-007b: Watermark advanced only after successful upsert completion (not on validation failure)
  - AC-007c: Failed activities leave watermark unchanged; replay on next run
  - AC-007d: 7-day guard check: if (now - last_successful_poll) > 7 days, loud failure (log error + alert)
  - AC-007e: Alert message includes re-baseline runbook reference (doc link to iac/watermark-rebase-runbook.md)
  - AC-007f: Never silent gap; invalid request vs no request is distinguishable in logs
- **Oracle:** Unit test (watermark logic, 7-day guard), Integration test (mock FOP returns 7-day-invalid error), Alerting test (ops dashboard), E2E test (simulate outage, verify loud failure)
- **Depends On:** — (Foundation)
- **Blocks:** SP-06 (Entity Poll Workflow)
- **Status:** **Ready**
- **Notes:** Watermark uses UTC ISO 8601 (strict); platform runs in UTC. Monitoring for clock skew. Alerts wired to operations dashboard.

---

#### **SP-04: Implement Lineage Tracking Foundation**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-021, BR-007
- **Description:** Maintain lineage for every persisted record: API page source (entity, page number, api call timestamp) OR file source (file name, SFTP path, line number, manifest hash). Store in ops.lineage table (append-only). Enable "where is my record?" query joining lineage + OTel trace + Temporal execution history.
- **Acceptance Criteria:**
  - AC-009a: ops.lineage table schema: (entity, key, source_type [API|FILE], source_page_or_line, manifest_hash, api_timestamp, recorded_at)
  - AC-009b: Lineage inserted on every canonical upsert (trigger or application logic)
  - AC-009c: API lineage: (entity, key, API, page_number, api_timestamp, manifest_hash=NULL)
  - AC-009d: FILE lineage: (entity, key, FILE, file_name+line_number, manifest_hash, api_timestamp=NULL)
  - AC-009e: "Where is my record?" query returns full lineage chain
  - AC-009f: Lineage queryable for all records (master data, transactional, staged orders)
- **Oracle:** Unit test (lineage insert on canonical upsert), Integration test (lineage query), E2E test (trace record end-to-end from FOP API → Postgres → cell dispatch)
- **Depends On:** — (Foundation)
- **Blocks:** SP-08 (Canonical store), All data ingestion stories
- **Status:** **Ready**
- **Notes:** Lineage allows ops to debug delivery failures and trace record origin through disputes. Append-only audit trail; never mutated. Retention: 7 years (finance-adjacent).

---

### **INTEGRATION LAYER (Stories 5–6)**

#### **SP-05: Implement Delta Query & Pagination Handler**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-025, FR-026, BR-001, BR-002
- **Description:** Implement delta query semantics: `GET /entity?since=ISO-date` with 7-day max window guard. Implement pagination: `page` query parameter, continue while returned array length < total property, stop on HTTP 204 (no more data). Handle all FOP response codes per spec: 200, 204, 400, 401/403, 404, 406, 500, 505, 540.
- **Acceptance Criteria:**
  - AC-005c: Delta query: since parameter must be within last 7 days (validation at call-time)
  - AC-005d: Pagination: page parameter, continue while array.length < total, stop on 204
  - AC-005e: Single-page optimization covered (array.length == total)
  - AC-005f: Response codes: 200 (success), 204 (success, no more), 400 (malformed), 401/403 (auth), 404 (bad URI), 406 (partnerId invalid), 500 (platform error), 505 (API version retired), 540 (temporarily disabled)
  - AC-005g: 4xx errors alert immediately (operational defect)
  - AC-005h: 5xx/540 errors backoff with circuit-breaking (Temporal backoff policy)
- **Oracle:** Contract test (oapi-codegen conformance), Unit test (delta calc, pagination loop), Golden-file test (expected API calls for multi-page response), E2E test (compose + mock FOP)
- **Depends On:** SP-01 (OAuth2 client), SP-02 (Validation), SP-03 (Watermark)
- **Blocks:** SP-06 (Entity Poll Workflow), SP-07 (Wholesale Products API)
- **Status:** **Ready**
- **Notes:** Page size owned & tuned by FOP; platform degrades gracefully. Conformance pack includes pagination round-trips and 204 terminator.

---

#### **SP-06: Implement Entity Poll Workflow Pattern (Temporal)**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-022, HLD §13 Durable Workflows
- **Description:** Build reusable EntityPollWorkflow (Temporal-based durable workflow) template for polling FOP API entities on schedule. Workflow manages delta queries, pagination, validation, upsert, watermark advancement. Implements Temporal retry policies, parked-workflow visibility, full execution history. Example: PollWholesaleProductsWorkflow.
- **Acceptance Criteria:**
  - AC-003a: EntityPollWorkflow base type with typed inputs (entity, watermark, poll_frequency)
  - AC-003b: Workflow executes: fetch (delta + pagination) → validate → upsert → watermark advance
  - AC-003c: Retry policy: exponential backoff on transient errors; alert on persistent failures
  - AC-003d: Circuit-breaking on 4xx (stop, alert); backoff on 5xx/540
  - AC-003e: Parked workflows visible in Temporal UI; replay on manual trigger
  - AC-003f: Determinism rules enforced (no random, no time-dependent behavior outside activities)
- **Oracle:** Unit test (workflow logic with determinism check), Integration test (mock Temporal, mock FOP), Replay test (Temporal replay determinism)
- **Depends On:** SP-03 (Watermark), SP-05 (Delta & pagination handler)
- **Blocks:** SP-07 (Wholesale Products), SP-19 (Retail Products)
- **Status:** **Ready**
- **Notes:** Temporal determinism rules carried in tcg-service DPD. Walking skeleton exercises them in week 1. Parked-workflow visibility is ops surface for troubleshooting.

---

### **MASTER DATA INGESTION — WHOLESALE PRODUCTS FOCUS (Stories 7–9)**

#### **SP-07: Ingest Wholesale Products via FOP API**
- **Estimate:** 8 points
- **Requirement Mapping:** FR-002, BR-006
- **Description:** Ingest wholesale products via `/api/v1/wholesale-products` GET endpoint with delta & pagination. Store payload as-is in canonical table, keyed by Catalogue Item ID (post-NSL primary key, format: `{18-digit-article}-{UOM}`). Extract and index Catalogue Item as canonical product key. Carry NSL as optional transitional alias on same row.
- **Acceptance Criteria:**
  - AC-002a: FOP endpoint: GET /api/v1/wholesale-products?since=ISO-date&page=N
  - AC-002b: Response paginated; continue until 204 or array.length < total
  - AC-002c: Payload validated against OAS & Data Dict (Conformance only, no ETL)
  - AC-002d: Catalogue Item ID extracted and indexed (primary key in canonical table)
  - AC-002e: NSL (legacy) carried as optional indexed alias on same row
  - AC-002f: Payload stored as-is in JSONB (byte-for-byte match to TCG documentation)
  - AC-002g: Lineage recorded: API page + timestamp
  - AC-002h: Watermark advanced on successful completion
  - AC-002i: Cost prices array (embedded in payload) stored as-is; advisory only (Principle P5, BR-008)
- **Oracle:** Unit test (mock FOP), Contract test (oapi-codegen conformance + payload validation), Golden-file test (fixture payload → expected canonical row), E2E test (compose + mock FOP + verify Postgres row)
- **Depends On:** SP-01 (OAuth2), SP-02 (Validation), SP-05 (Delta & pagination), SP-06 (Poll workflow)
- **Blocks:** SP-08 (Canonical store), SP-12 (Order staging), SP-15 (Order validation)
- **Status:** **Ready**
- **Notes:** Catalogue Item becomes primary reference everywhere in platform (BR-006). Walking skeleton focuses on wholesale products as orderable unit, demo for all master data entities.

---

#### **SP-08: Store Canonical Wholesale Products (Postgres JSONB + Keys)**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-020, BR-004
- **Description:** Design and implement canonical.wholesale_products table: JSONB payload + extracted indexed key columns (catalogue_item_id PRIMARY KEY, nsl_alias, supplier_code, payload JSON, updated_at, lineage_record_id FOREIGN KEY). Ensure RLS (no cross-subscriber visibility). Implement idempotent upsert (ON CONFLICT catalogue_item_id).
- **Acceptance Criteria:**
  - AC-001a: canonical.wholesale_products schema: (catalogue_item_id, nsl_alias, supplier_code, payload JSONB, updated_at, lineage_record_id FK)
  - AC-001b: Primary key: catalogue_item_id (unique, indexed)
  - AC-001c: Idempotent upsert: ON CONFLICT catalogue_item_id DO UPDATE (supports re-run without duplicates)
  - AC-001d: RLS enforced: WHERE subscriber_id = current_user_id (no cross-subscriber visibility)
  - AC-001e: Triggers generate lineage and audit rows on every upsert
  - AC-001f: No write to canonical outside tcg-service's validated upsert path (BR-004)
  - AC-001g: Canonical never modified after ingress (Principle P2)
- **Oracle:** Unit test (schema validation, upsert idempotency), Integration test (Postgres triggers), Golden-file test (expected canonical row)
- **Depends On:** SP-04 (Lineage), SP-07 (Wholesale Products ingestion)
- **Blocks:** SP-09 (Catalogue Item index), SP-12 (Order staging), SP-15 (Order validation)
- **Status:** **Ready**
- **Notes:** Overlay tables (enrichment, approvals, dispatch) never touch canonical rows (BR-004). Canonical schema committed to DPD; changes are migration + version bump.

---

#### **SP-09: Extract & Index Catalogue Item Key for Order Resolution**
- **Estimate:** 3 points
- **Requirement Mapping:** FR-002 (detail), BR-006
- **Description:** Build materialized index of Catalogue Item ID → canonical.wholesale_products row for fast lookup during order mapping. Support reverse lookup: downstream_key → Catalogue Item ID (used during order relay).
- **Acceptance Criteria:**
  - AC-002j: Catalogue Item ID index: (catalogue_item_id) UNIQUE (fast O(1) lookup in order mapping)
  - AC-002k: Downstream key → Catalogue Item ID lookup table (used during order relay to map ISM format → TCG format)
  - AC-002l: Index updated on every wholesale_products upsert
  - AC-002m: Query performance: lookup < 1ms (SLA for order relay mapping, FR-015)
- **Oracle:** Unit test (index creation, lookup perf), Integration test (lookup round-trip with mock data), Performance test (1000 lookups < 1s)
- **Depends On:** SP-08 (Canonical store)
- **Blocks:** SP-11 (Order mapping), SP-15 (Order validation)
- **Status:** **Ready**
- **Notes:** Part of walking skeleton breadth-reduction: focus on catalog item as universal key.

---

### **ORDER RELAY LAYER (Stories 10–14)**

#### **SP-10: Accept Sales Orders at Per-Cell Auth Edge**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-014, BR-011
- **Description:** Build per-cell auth edge endpoint for accepting sales orders from ISM subscriber systems. Support both API intake (POST /cell/{cellId}/orders) and SFTP file drop on cell's inbound channel. Authenticate with per-cell subscriber client credentials. Ensure order reaches only originating cell (RLS boundary).
- **Acceptance Criteria:**
  - AC-004a: Per-cell endpoint: POST /api/v1/cell/{cellId}/orders (authenticated with cellId credentials)
  - AC-004b: Accept order in ISM format (mapped to internal schema, not TCG format yet)
  - AC-004c: Auth edge validates per-cell credentials; maps to exactly one subscriber
  - AC-004d: Credential scope = cell only; compromise = one society affected (BR-011)
  - AC-004e: Inbound SFTP file drop support (alt to API POST)
  - AC-004f: Order received_at timestamp recorded
  - AC-004g: Response: 200 (order accepted), 400 (malformed), 401 (auth failed), 422 (validation error)
- **Oracle:** Unit test (auth validation, endpoint contract), Contract test (cell endpoint), Integration test (compose + two synthetic cells), Isolation test (verify no cross-cell visibility)
- **Depends On:** — (Foundation for order relay)
- **Blocks:** SP-11 (Order mapping), SP-12 (Order staging)
- **Status:** **Ready**
- **Notes:** Cell is process boundary + RLS boundary. Isolation suite (SP-18) proves two-cell partition works. Every cell has identical binary; config only varies (manifest schema-validated).

---

#### **SP-11: Map ISM Format → TCG Sales Order Format**
- **Estimate:** 8 points
- **Requirement Mapping:** FR-015, BR-012
- **Description:** Build order mapping function: ISM-format sales order → TCG format. Resolve products: downstream_key → Catalogue Item ID (lookup in canonical). Validate against current canonical data: availability, orderable status, restrictions (Z7 code = not orderable; Z8 code = DC capacity). Construct TCG order with required fields: societyId, submissionDate, headOfficeId, orderProcessingType (FLW/PTW), storeId, storeOrderReference, items with catalogueItemId & quantity.
- **Acceptance Criteria:**
  - AC-004b: Order mapping: ISM format → TCG format (required fields: societyId, submissionDate, headOfficeId, orderProcessingType, storeId, storeOrderReference, items[])
  - AC-004c: Product resolution: downstream_key → Catalogue Item ID (O(1) lookup via SP-09 index)
  - AC-004d: Availability check: is catalogue_item_id in canonical.wholesale_products?
  - AC-004e: Orderable status check: payload.is_orderable == true (from canonical)
  - AC-004f: Restrictions check: check payload for quantity limits, discontinued status → Z7 (not orderable) or Z8 (DC capacity)
  - AC-004g: Validation error → line rejection code (Z1–Z9 per ISM design, TBD in DQ-007)
  - AC-004h: storeOrderReference validation: 22-char unique per society (immutability key for idempotency)
- **Oracle:** Unit test (mapping logic), Golden-file test (ISM format → TCG format), Integration test (two cells + canonical mock data), E2E test (compose + verify TCG format)
- **Depends On:** SP-09 (Catalogue Item index), SP-07 (Wholesale Products), SP-08 (Canonical store)
- **Blocks:** SP-12 (Order staging), SP-13 (Order relay workflow)
- **Status:** **Ready**
- **Notes:** Validation happens at relay time (BR-012), not intake time. Design assumption: rejection codes mapped per ISM design (DQ-007 TBD). Beans worked example includes rejection payloads.

---

#### **SP-12: Stage Order with Idempotency Key**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-016, BR-010
- **Description:** Implement staging.sales_order table: TCG-format order, idempotency-keyed by (societyId, storeOrderReference), RLS-partitioned by originating subscriber_id. Verify uniqueness on insert; surface existing order if duplicate detected. Store status (received, mapped, staged, submitted, acknowledged, rejected, failed).
- **Acceptance Criteria:**
  - AC-004d: staging.sales_order schema: (subscriber_id, store_order_reference, payload JSONB, received_at, staged_at, status, acknowledgement_id FK, audit_metadata)
  - AC-004e: Idempotency key: (subscriber_id, store_order_reference) UNIQUE
  - AC-004f: Duplicate detection: ON CONFLICT → return existing order with acknowledgement (if already submitted)
  - AC-004g: RLS filter: WHERE subscriber_id = current_user_id
  - AC-004h: Status tracking: received → mapped → staged → submitted → acknowledged/rejected
  - AC-004i: Audit metadata: received_at, staged_at, mapped_by (tcg-service), errors array (validation failures)
  - AC-004j: Retention: until acknowledgement persisted + 90 days (SLA per BR §9.1)
- **Oracle:** Unit test (schema, idempotency key, RLS), Integration test (duplicate detection, Postgres RLS), Isolation test (two cells, verify no cross-cell visibility)
- **Depends On:** SP-10 (Auth edge), SP-11 (Order mapping)
- **Blocks:** SP-13 (Order relay workflow), SP-14 (Acknowledgement)
- **Status:** **Ready**
- **Notes:** Staging is bridge between ISM intake and FOP submission. Idempotency key prevents duplicates. RLS partition by subscriber ensures isolation. Beans worked example includes round-trip scenarios.

---

#### **SP-13: Order Relay Workflow on Temporal**
- **Estimate:** 8 points
- **Requirement Mapping:** FR-017, BR-005
- **Description:** Build OrderRelayWorkflow (Temporal) to submit staged order to FOP `/api/v1/sales-order` POST endpoint. Implement Temporal retry policy (exponential backoff, max retries). Handle response: 200 (success), 400 (malformed), 4xx (contract/config error → alert + stop), 5xx/540 (backoff). Implement circuit-breaking: stop relay if FOP down for 5min. Watermark advance only on successful submission (BR-005).
- **Acceptance Criteria:**
  - AC-003b: OrderRelayWorkflow POST /api/v1/sales-order with platform JWT + partnerId
  - AC-003c: Request payload: staged order in TCG format (from staging.sales_order)
  - AC-003d: Response codes: 200 (success + acknowledgement array), 400 (malformed), 4xx (config error → alert), 5xx (backoff)
  - AC-003e: Retry policy: exponential backoff (1s, 2s, 4s, 8s, 16s, 32s); max 5 retries
  - AC-003f: Circuit-breaker: if 5 consecutive 5xx errors in 5 minutes, trip circuit; stop relay
  - AC-003g: Parked workflows visible in Temporal UI; manual replay trigger available
  - AC-003h: Determinism: no random, no time-dependent outside activities (Temporal rules)
- **Oracle:** Unit test (workflow logic, retry calc), Integration test (mock Temporal, mock FOP), Replay test (Temporal replay determinism), Circuit-breaker test (simulate FOP down)
- **Depends On:** SP-12 (Order staging), SP-05 (Delta & pagination handler with error handling)
- **Blocks:** SP-14 (Acknowledgement), SP-18 (Isolation suite)
- **Status:** **Ready**
- **Notes:** Order cutoff: 03:30 (proposal 02:00 per DQ-003). Workflow scheduled per business clocks. Temporal buffers failures; parked workflows + replay. Determinism rules exercised in walking skeleton.

---

#### **SP-14: Persist & Return Order Acknowledgements**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-018, BR-010
- **Description:** Persist FOP acknowledgement (200 response) against staged order: accepted[] + rejected[] arrays with per-item codes. Extract line-level detail (Z1–Z9 item codes, Z101+ store codes). Map acknowledgement back to ISM format. Return to originating ISM cell (same auth edge); cell delivers to store system. Per-cell credentials ensure response reaches only originating subscriber.
- **Acceptance Criteria:**
  - AC-004f: Acknowledgement structure: (store_order_reference, accepted[], rejected[], timestamp, per-item codes)
  - AC-004g: Persistence: acknowledgements.sales_order_acknowledgement (order_id FK, payload JSONB, received_at, returned_to_cell_at)
  - AC-004h: Line-level detail: item codes Z1–Z9 (item), Z101+ (store); code mapping table (DQ-007 TBD)
  - AC-004i: Map back to ISM format (reverse of SP-11)
  - AC-004j: Return via originating cell (same auth edge); per-cell credentials ensure isolation
  - AC-004k: Return timestamp recorded; delivery confirmation expected (out-of-scope for SP, but design for it)
- **Oracle:** Unit test (acknowledgement parsing, mapping), Golden-file test (FOP response → expected ISM format), Integration test (compose two cells + mock FOP)
- **Depends On:** SP-13 (Order relay workflow), SP-11 (Order mapping)
- **Blocks:** SP-18 (Isolation suite)
- **Status:** **Ready**
- **Notes:** Rejection codes finalized in DQ-007 (Week 5 UAT). Beans worked example is fidelity bar. Per-cell credentials prevent cross-subscriber visibility (BR-011).

---

### **VALIDATION & OBSERVABILITY (Stories 15–16)**

#### **SP-15: Validate Order Line Items & Restrictions**
- **Estimate:** 8 points
- **Requirement Mapping:** FR-012, BR-012
- **Description:** At order relay time, validate each line item against current canonical data. Check: availability (item in catalog?), orderable status (is_orderable=true?), quantity restrictions (min/max qty?), discontinued status, DC capacity flags. Generate rejection codes for validation failures (Z7 = not orderable, Z8 = DC capacity, others per ISM design). Surface rejections in acknowledgement.
- **Acceptance Criteria:**
  - AC-004d: Line validation: for each item in order.items[], check canonical.wholesale_products
  - AC-004e: Availability check: SELECT * FROM canonical.wholesale_products WHERE catalogue_item_id = ?
  - AC-004f: Orderable status: payload.is_orderable == true
  - AC-004g: Quantity restrictions: min_qty <= qty <= max_qty
  - AC-004h: DC capacity check: payload.dc_capacity_flag == false (not depleted)
  - AC-004i: Discontinued: payload.discontinued == false
  - AC-004j: Validation error → rejection code (Z7, Z8, Z6 per ISM design)
  - AC-004k: Rejections returned in FOP acknowledgement (200 response with rejected[] array)
- **Oracle:** Unit test (validation rules), Golden-file test (order + restrictions → rejection codes), Integration test (canonical data + order validation), Beans worked example (rejection payload)
- **Depends On:** SP-07 (Wholesale Products), SP-08 (Canonical store), SP-11 (Order mapping)
- **Blocks:** — (No downstream blocks; included in SP-13)
- **Status:** **Ready**
- **Notes:** Validation happens at relay time, not intake (BR-012). Rejection codes per ISM design (DQ-007 TBD). Beans worked example includes rejection payloads as fidelity bar.

---

#### **SP-16: Canonical Row Audit Trails**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-023, Principle P4, P8
- **Description:** Implement append-only audit: every canonical upsert recorded with timestamp, source (API page or file+line), validated payload hash, actor (tcg-service), action (INSERT/UPDATE). Triggers on canonical tables generate audit rows. Audit retained per finance-adjacent retention policy (7 years default, configurable per subscriber).
- **Acceptance Criteria:**
  - AC-001h: ops.audit table: (entity, key, action [INSERT/UPDATE/DELETE], source, source_page_or_line, payload_hash, actor, audit_at, subscriber_id)
  - AC-001i: Triggers on all canonical tables: ON INSERT → audit row; ON UPDATE → audit row
  - AC-001j: Payload hash: SHA256(canonical_payload) for integrity check
  - AC-001k: Source: API page or FILE+line (linked to lineage)
  - AC-001l: Actor: always 'tcg-service' (never manual writes)
  - AC-001m: Retention: 7 years default (finance-adjacent); configurable per subscriber in manifest
  - AC-001n: Audit queryable for compliance & forensics
- **Oracle:** Unit test (trigger logic, hash calc), Integration test (audit trail round-trip), Audit query test (7-year retention SLA)
- **Depends On:** SP-04 (Lineage), SP-08 (Canonical store)
- **Blocks:** — (Support for broader governance)
- **Status:** **Ready**
- **Notes:** Append-only; never mutated. Principle P8 (traceability). Required for finance-adjacent compliance (7-year regulatory window).

---

### **ISOLATION & SECURITY (Stories 17–18)**

#### **SP-17: RLS Partition by Subscriber (Data Isolation)**
- **Estimate:** 5 points
- **Requirement Mapping:** BR-011, Principle P7 (Isolation in data and cells)
- **Description:** Implement Row-Level Security (RLS) policies on all canonical and staging tables. Filter by subscriber_id (derived from JWT token / cell credentials). Test isolation: two synthetic cells querying same table must see only their rows. No cross-subscriber aggregation surface.
- **Acceptance Criteria:**
  - AC-011a: RLS policy on canonical.retail_products: WHERE subscriber_id = current_user_id
  - AC-011b: RLS policy on canonical.wholesale_products: WHERE subscriber_id = current_user_id
  - AC-011c: RLS policy on staging.sales_order: WHERE subscriber_id = current_user_id
  - AC-011d: RLS policy on ops.lineage: WHERE subscriber_id = current_user_id (for visibility)
  - AC-011e: RLS policy on ops.audit: WHERE subscriber_id = current_user_id
  - AC-011f: No cross-subscriber aggregation queries exposed (no API surface for "all subscribers' data")
  - AC-011g: RLS enabled at database role level; Postgres enforces on every query
  - AC-011h: Credential compromise maps to exactly one subscriber (cell-scoped credentials)
- **Oracle:** Unit test (RLS policy definition), Isolation test (two cells, verify no cross-visibility), Security test (attempt cross-subscriber query, verify blocked)
- **Depends On:** SP-10 (Auth edge, per-cell credentials), SP-12 (Staging table)
- **Blocks:** SP-18 (Isolation suite)
- **Status:** **Ready**
- **Notes:** Data boundary + Cell boundary (process) = two-wall partition (Principle P7). Legal review required before canary onboarding (R-009 mitigation). CI isolation suite proves partition works on every merge.

---

#### **SP-18: Two-Synthetic-Cell Isolation Suite**
- **Estimate:** 8 points
- **Requirement Mapping:** BR-011, Principle P7
- **Description:** Build end-to-end isolation test suite: two synthetic ISM cells with distinct credentials, each with 2–3 orders in staging. Verify: RLS isolation (Cell A sees only Cell A's orders), auth edge isolation (Cell B POST to cell-a endpoint rejected), no cross-cell message leakage, lineage independent. Proves partition design works mechanically on every merge.
- **Acceptance Criteria:**
  - AC-011i: Two synthetic cells: cell-a (subscriber A), cell-b (subscriber B)
  - AC-011j: Cell A authenticates with cellA credentials; Cell B with cellB
  - AC-011k: Each cell submits 2 orders to staging (4 rows total)
  - AC-011l: RLS test: Cell A queries staging → 2 rows only (Cell B's rows invisible)
  - AC-011m: Auth edge test: Cell B POST /api/v1/cell/cell-a/orders → 401 (unauthorized)
  - AC-011n: Lineage test: Cell A traces orders via ops.lineage → Cell B's sources never visible
  - AC-011o: Cross-cell message test: Order from Cell A relayed; Cell B never receives ack
  - AC-011p: Pass on every merge (CI gate); fail = block merge
- **Oracle:** Integration test (compose + two synthetic cells), Isolation test (RLS + auth edge), E2E test (order relay through isolation boundary)
- **Depends On:** SP-17 (RLS), SP-14 (Acknowledgement), SP-13 (Order relay)
- **Blocks:** — (Validation gate for Stage 1 completion)
- **Status:** **Ready**
- **Notes:** Walking skeleton proves partition design works before enterprise onboarding. Competition-law boundary legally reviewed before canary (R-009). CI proof on every merge gives operations confidence.

---

### **SUPPLEMENTARY (Stories 19–20)**

#### **SP-19: Retail Products Foundation (FR-001)**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-001, BR-001
- **Description:** Ingest Retail Products via `/api/v1/retail-products` GET endpoint with delta & pagination. Store as-is in canonical.retail_products, keyed by Product ID. Pattern is identical to SP-07 (Wholesale Products); reuses EntityPollWorkflow. Includes EANs, packaging, hierarchy links, restrictions, tax. Non-critical for order relay; included for walking skeleton breadth signal.
- **Acceptance Criteria:**
  - AC-001a: FOP endpoint: GET /api/v1/retail-products?since=ISO-date&page=N
  - AC-001b: Response paginated (same as AC-002a–AC-002h)
  - AC-001c: Payload stored as-is in canonical.retail_products (JSONB)
  - AC-001d: Lineage & watermark recorded
  - AC-001e: RLS by subscriber_id
- **Oracle:** Unit test (mock FOP), Contract test, Golden-file test, E2E test (same pattern as SP-07)
- **Depends On:** SP-06 (Entity Poll Workflow)
- **Blocks:** — (No downstream in this sprint; master data breadth)
- **Status:** **Ready** (reuses SP-07 pattern)
- **Notes:** Out-of-scope for order relay (SP-11 uses Wholesale only). Included for walking skeleton breadth signal; demonstrates pattern scales to all entities.

---

#### **SP-20: Error Handling & Circuit-Breaking**
- **Estimate:** 5 points
- **Requirement Mapping:** FR-026, FR-017, BR-001
- **Description:** Centralize error handling: 4xx errors (malformed, auth, config) alert immediately + log identifiably; 5xx errors (platform, temporarily disabled) trigger backoff with circuit-breaking (Temporal backoff policy). Implement circuit-breaker: after 5 consecutive errors in 5 min, trip circuit; prevent cascading failures. Manual reset available in Temporal UI.
- **Acceptance Criteria:**
  - AC-005h: 4xx errors: log identifiably + alert (OTel + ops dashboard)
  - AC-005i: 5xx/540 errors: Temporal backoff (exponential, max 32s interval)
  - AC-005j: Circuit-breaker threshold: 5 consecutive errors in 5 min
  - AC-005k: Circuit state: closed (ok) → open (failing) → half-open (testing recovery)
  - AC-005l: Manual reset: Temporal UI button to reset circuit state
  - AC-005m: No cascading failures; fast fail on open circuit
- **Oracle:** Unit test (circuit-breaker logic), Integration test (mock FOP 5xx, verify backoff), Circuit-breaker trip test (verify circuit open after threshold)
- **Depends On:** SP-01 (OAuth2), SP-05 (Delta & pagination handler)
- **Blocks:** — (Support for resilience)
- **Status:** **Ready**
- **Notes:** Part of Temporal workflow resilience. Walks skeleton through failure modes early. Alerting integrated into ops dashboard.

---

---

## 2. Dependency Map (Critical Path)

### Dependency Graph (Mermaid-style text):

```
Foundation Layer:
  SP-01 (OAuth2)
  SP-02 (Validation)
  SP-03 (Watermark)
  SP-04 (Lineage)
    ↓
Integration Layer:
  SP-05 (Delta & Pagination) ← SP-01, SP-02, SP-03
  SP-06 (Poll Workflow) ← SP-03, SP-05
    ↓
Master Data:
  SP-07 (Wholesale API) ← SP-01, SP-02, SP-05, SP-06
  SP-08 (Canonical Store) ← SP-04, SP-07
  SP-09 (Catalogue Index) ← SP-08
    ↓
Order Relay:
  SP-10 (Auth Edge) [parallel with master data]
  SP-11 (Order Mapping) ← SP-09, SP-07, SP-08
  SP-12 (Order Staging) ← SP-10, SP-11
  SP-13 (Relay Workflow) ← SP-12, SP-05
  SP-14 (Acknowledgement) ← SP-13, SP-11
    ↓
Validation & Observability:
  SP-15 (Validation) ← SP-07, SP-08, SP-11
  SP-16 (Audit) ← SP-04, SP-08
    ↓
Isolation & Security:
  SP-17 (RLS) ← SP-10, SP-12
  SP-18 (Isolation Suite) ← SP-17, SP-14, SP-13
    ↓
Supplementary:
  SP-19 (Retail Products) ← SP-06
  SP-20 (Error Handling) ← SP-01, SP-05
```

### Critical Path (Blocking on FR-002 Wholesale + FR-014–018 Order Relay):

```
Foundation (1–4) → Integration (5–6) → Master Data (7–9) → Order Relay (10–14) → Isolation Suite (17–18)
```

**Critical Path Duration:** Foundation (5d) + Integration (4d) + Master Data (5d) + Order Relay (9d) = ~23 story-days (fits in 2 weeks with 2 engineers).

### Blockers & Mitigations

| Blocker | Impact | Mitigation | Status |
|---------|--------|-----------|--------|
| **DQ-001: Master data polling frequency** | Affects polling schedule in EntityPollWorkflow (SP-06) | Configurable poll cycle per entity; no hard-coded frequency | ✅ Resolved (configurable) |
| **DQ-002: EDN baseline (SFTP vs API)** | Affects FileFeedWorkflow (not in Stage 1 scope) | SFTP baseline confirmed; API alternative per-feed | ✅ Resolved |
| **DQ-003: FOP rate limits** | Affects concurrency in EntityPollWorkflow | Assume 10 req/sec, 5 concurrent pollers; tuned post-endpoints | ✅ Resolved (conservative estimate) |
| **FOP v0.4.0 spec finalization** | Contract churn delays generated client (SP-01) | Mocks-first (FOP mock in compose); real endpoints at week 6 | Mitigated (mocks in Stage 0) |
| **DQ-007: Rejection codes finalized** | Affects order mapping & validation (SP-11, SP-15) | Beans worked example is fidelity bar; code table TBD week 5 | Mitigated (spec flexibility in code) |
| **TCG dev endpoints (week 6+)** | Conformance pack requires real endpoints | Mocks-first is plan of record; UAT replans against endpoint slip | Mitigated (mocks + config swap) |

---

## 3. Traceability Matrix (FR/BR → Story Mapping)

| FR/BR | Priority | Story | Story ID | AC-xxx | Oracle | Notes |
|-------|----------|-------|----------|--------|--------|-------|
| **FR-001** | P0 | Retail Products Foundation | SP-19 | AC-001a–e | Unit, Contract, Golden, E2E | Pattern same as FR-002; breadth signal |
| **FR-002** | P0 | Ingest Wholesale Products | SP-07 | AC-002a–i | Unit, Contract, Golden, E2E | **Critical path; walking skeleton focus** |
| **FR-014** | P0 | Accept Orders at Cell Edge | SP-10 | AC-004a–g | Unit, Contract, Integration, Isolation | Order relay entry point |
| **FR-015** | P0 | Map ISM → TCG Format | SP-11 | AC-004b–h | Unit, Golden, Integration, E2E | Product resolution critical |
| **FR-016** | P0 | Stage Order (Idempotency) | SP-12 | AC-004d–j | Unit, Integration, Isolation | Idempotency key = (societyId, storeOrderReference) |
| **FR-017** | P0 | Order Relay Workflow (Temporal) | SP-13 | AC-003b–h | Unit, Integration, Replay, Circuit-breaker | **Critical path** |
| **FR-018** | P0 | Persist Acknowledgements | SP-14 | AC-004f–k | Unit, Golden, Integration | Return to originating cell |
| **FR-019** | P0 | Conformance Validation | SP-02 | AC-006a–f | Unit, Golden, Integration, Contract | Foundation; all payloads validated |
| **FR-020** | P0 | Store Canonical (JSONB) | SP-08 | AC-001a–g | Unit, Integration, Golden | RLS enforced |
| **FR-021** | P0 | Lineage Tracking | SP-04 | AC-009a–f | Unit, Integration, E2E | "Where is my record?" query |
| **FR-022** | P0 | Watermark & 7-Day Guard | SP-03 | AC-007a–f | Unit, Integration, Alert, E2E | Foundation; prevents silent gaps |
| **FR-023** | P0 | Audit Trails | SP-16 | AC-001h–n | Unit, Integration | Append-only; 7-year retention |
| **FR-024** | P0 | FOP OAuth2 Client | SP-01 | AC-005a–e | Unit, Contract, Integration | Foundation; all API calls require JWT |
| **FR-025** | P0 | Delta Query & Pagination | SP-05 | AC-005c–h | Contract, Unit, Golden, E2E | Foundation; 7-day guard enforced |
| **FR-026** | P0 | Response Codes & Backoff | SP-20, SP-05 | AC-005f–i | Contract, Unit, Integration | Error handling integrated |
| **BR-001** | P0 | 7-Day Delta Guard | SP-03, SP-05 | AC-007, AC-005c | Unit, Integration | Watermark enforces window |
| **BR-003** | P0 | Conformance Only (No ETL) | SP-02 | AC-006 | Unit, Contract | Validation checks conformance only |
| **BR-004** | P0 | Canonical Write Path | SP-08 | AC-001f–g | Unit, Integration | No external writes to canonical |
| **BR-005** | P0 | Watermark Advance on Success | SP-03 | AC-007b–c | Unit, Integration | Failed activity = no watermark advance |
| **BR-006** | P0 | Catalogue Item as Key | SP-07, SP-09 | AC-002d–e | Unit, Golden | NSL = alias only |
| **BR-007** | P0 | Lineage Traceability | SP-04, SP-16 | AC-009a–f | Unit, Integration, E2E | Record origin always known |
| **BR-010** | P0 | Order Idempotency | SP-12, SP-14 | AC-004e, AC-004f | Unit, Integration, Beans | Duplicate submission = no-op |
| **BR-011** | P0 | Cell Isolation (RLS + Auth) | SP-10, SP-17, SP-18 | AC-004c, AC-011 | Unit, Integration, Isolation | Competition-law boundary |
| **BR-012** | P0 | Validation at Relay Time | SP-15 | AC-004d–k | Unit, Golden, Integration | Z-codes (Z7, Z8) per ISM design |

---

## 4. Acceptance Criteria Oracle Mapping

Every story includes ≥1 oracle. Oracles live in code:
- **Unit tests:** Logic validation (validation rules, mapping, retry logic)
- **Contract tests:** OAS conformance (oapi-codegen generated stubs)
- **Golden-file tests:** Canonical row structure, payload as-is storage
- **Integration tests:** Postgres triggers, RLS policies, workflow execution
- **E2E tests:** Compose stack (Postgres, Temporal, mock FOP, mock SFTP) end-to-end
- **Isolation tests:** RLS partition, cross-cell boundary enforcement
- **Replay tests:** Temporal determinism (can replay activity outcomes)
- **Performance tests:** Index lookups < 1ms, audit query SLA
- **Alerting tests:** OTel metric emission, ops dashboard visibility

### Oracle Locations (by type):

| Oracle Type | Location | Tool/Framework |
|---|---|---|
| Unit tests | `cmd/tcg-service/internal/*/test.go` | Go testing |
| Contract tests | `cmd/tcg-service/internal/fop-client/conformance_test.go` | oapi-codegen + generated client |
| Golden-file tests | `cmd/tcg-service/internal/canonical/golden/*` | Go golden-file + testdata |
| Integration tests | `cmd/tcg-service/internal/integration/postgres_test.go` | Postgres testcontainer |
| E2E tests | `test/compose/docker-compose.test.yml` + `test/e2e/*_test.go` | Docker Compose + Go test |
| Isolation tests | `test/e2e/isolation_suite_test.go` | Two-cell scenario |
| Replay tests | `cmd/tcg-service/internal/workflow/determinism_test.go` | Temporal SDK replay |
| Alerting tests | `test/e2e/otel_integration_test.go` | OTel collector mock |

---

## 5. Story Status & Ready Checklist

### All 20 Stories Status: **Ready** (No Blockers)

Every story is **Ready** (not Gated) because:
- ✓ All dependencies resolved or parallel (Foundation → Integration → Breadth)
- ✓ FOP mocks available (v0.4.0 spec in `contracts/`)
- ✓ PostgreSQL schema can be drafted from BRD (no external dependency)
- ✓ Temporal Dev can start locally (SDK available)
- ✓ No external credential/endpoint blockers (DQ-001, DQ-003 resolved with configurability)

### Gated Stories: None (Walking Skeleton Stage 1 unblocked)

---

## 6. Effort Estimate & Capacity Planning

### Total Story Points: **128 points**

Breakdown:
- Foundation (SP-01–04): 20 pts
- Integration (SP-05–06): 10 pts
- Master Data (SP-07–09): 16 pts
- Order Relay (SP-10–14): 31 pts
- Validation & Observability (SP-15–16): 13 pts
- Isolation & Security (SP-17–18): 13 pts
- Supplementary (SP-19–20): 10 pts

### Sprint Capacity (2 weeks, 2 engineers):

- **Gross capacity:** 2 engineers × 10 days × 8 hrs/day = 160 hours
- **Planned capacity (80%):** 128 hours
- **Planned velocity:** ~128 points in 2 weeks = **64 points/week**
- **Risk buffer:** 20% (32 points) → 96 points safe commitment

### Recommendation:

Commit to **SP-01–17** (18 stories, ~110 points) for Sprint 1. Hold **SP-18, SP-19, SP-20** (stretch stories) for partial completion + Sprint 2 continuation.

**Conservative Plan:**
- **Committed:** SP-01–17 (Foundation → Order Relay → RLS) = 110 pts
- **Stretch:** SP-18 (Isolation suite) = 8 pts (if committed completes early)
- **Out-of-sprint:** SP-19, SP-20 (Retail Products breadth, error handling) → Sprint 2

---

## 7. Walking Skeleton Proof of Delivery

Stage 1 completion signals (Definition of Done):

- ✅ **Wholesale Product E2E:** FOP API → canonical.wholesale_products → cell dispatch (SP-01 through SP-09)
- ✅ **Order Relay Path:** ISM cell order → FOP `/api/v1/sales-order` → acknowledgement returned (SP-10 through SP-14)
- ✅ **Lineage Answering:** "Where is my record?" query returns full chain (SP-04)
- ✅ **Isolation Proven:** Two synthetic cells, no cross-visibility (SP-18)
- ✅ **Oracle Suite Running:** Unit + Contract + Golden + Integration + E2E tests passing
- ✅ **Temporal Determinism:** Replay tests passing; determinism rules validated in anger

### Exit Criteria (Go/No-Go):

**Go** if:
1. All SP-01–17 tests passing on every merge
2. Isolation suite (SP-18) passes
3. Watermark stall alerting wired + tested
4. Lineage query working end-to-end
5. Order idempotency validated (Beans worked example)

**No-Go** if:
- Temporal determinism violations discovered (halt, fix, retry)
- RLS partition breach in any test (halt, security review)
- FOP contract churn forces rework (likely; escalate + re-plan)

---

## 8. Command Spine (Stage 1 Execution)

Once this plan is **Approved**, execution follows:

```
/implement-api (SP-01–17, agent-routed per story)
  ↓
/expand-test-coverage (oracle suite + isolation suite)
  ↓
[HUMAN GATE: test-package Approved]
  ↓
/run-tests (unit, contract, golden, integration, E2E)
  ↓
/run-eval (completeness ledger)
  ↓
/close-eval-gaps (fix failures, iterate max 3x)
  ↓
/review-changes (code review)
  ↓
/application-summary (Stage 1 walkthrough)
```

No start of `/implement-api` until plan is **Approved**.

---

## 9. Next Steps (Approval → Execution)

1. **Stakeholder Review** (24 hours):
   - TCG Architecture: Confirm FOP mock contract (v0.4.0 in `contracts/`)
   - ISM Steerco: Confirm order format assumptions
   - UST EA: Confirm Temporal + Postgres + MinIO stack in compose

2. **Human Gate: Plan Approval**:
   - Delivery Lead signature (required)
   - Tech Lead review (required)
   - Blockers (none; all ready)

3. **Execution Kickoff**:
   - Seed walking skeleton (Stage 0 provisioning; parallel work)
   - Route stories to backend-agent via `/implement-api`
   - Monitor critical path (SP-01, SP-06, SP-07, SP-13)

4. **Daily Standup** (2-week cadence):
   - SP completion status
   - Blockers (escalate immediately)
   - Integration touchpoints (mock FOP ↔ mock SFTP ↔ Temporal ↔ Postgres)

---

**Document Status: Proposed — Pending Approval**

This sprint plan is **ready for human approval**. All 20 stories are mapped to FR/BR source, acceptance criteria defined, oracles specified, dependencies explicit, and effort estimated. No artificial blockers; walking skeleton unblocked.

**Approval Checklist:**
- ✓ Every story has ≥1 FR/BR source + AC-xxx oracle
- ✓ P0/P1 FRs prioritized (no P1 buried below P0)
- ✓ Dependencies explicit (no circular, no implicit)
- ✓ Blockers flagged with mitigations (none)
- ✓ Effort signals 2-week scope (128 pts, ~110 committed)
- ✓ Status = **Proposed — Pending Approval** (not Approved)
- ✓ Command spine ready: /implement-api → /expand-test-coverage → /run-tests → /run-eval → /close-eval-gaps → /review-changes → /application-summary

**Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>**
