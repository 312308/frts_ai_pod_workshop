# FRTS Data Ingestion Layer — Business Requirements Document

**Status:** ✅ Accepted (2026-09-16)

**Document Version:** 1.0  
**Date:** 2026-09-16  
**Author:** Claude Haiku 4.5  
**Source Documents:**
- FRTS_Ingress_Platform_HLD_v4.0_MASTER.docx (version 4.0, 25 August 2026)
- FRTS Business Requirements.docx (NSL Retirement, v28.07.26)
- FRTS_API_Data_Dictionary.xlsx (version 1.1, 20 July 2026)
- CO-OP TCG FRTS Ordering Platform API Guide v0.4.0 (draft)
- FRTS Architecture dgm.png (system context diagram)

---

## 1. Purpose

This document specifies the functional, business, and non-functional requirements for the **Data Ingestion Layer** (tcg-service) of the FRTS Ingress Platform. The ingestion layer is responsible for all inbound data exchange with The Co-operative Group (TCG), including master data via FOP APIs, transactional data via SFTP file feeds, and the submission of sales orders back to FOP.

The Data Ingestion Layer acts as the platform's single point of integration with TCG, ensuring no duplicate pipelines, validated conformance, and reliable delivery of canonical data to the persistence layer.

---

## 2. Scope

### In Scope
- Ingestion of all TCG master data entities via FOP API (v0.4.0+):
  - Retail Products
  - Wholesale Products / Catalogue Items
  - Product Hierarchies
  - Promotions
  - Suppliers
  - Wholesale Prices (cost prices, effective-dated)

- Ingestion of transactional data via SFTP pull from TCG-hosted endpoint:
  - Electronic Delivery Notifications (EDN) — primary path SFTP
  - Depot Invoices
  - Direct Invoices
  - Claims (Depot / Direct)
  - Condition Contracts
  - Promotional Sales Volumes (outbound)
  - Planograms, RADR Reports, PSA Files

- Sales Order submission:
  - Intake from ISM subscriber cells (inbound order relay)
  - Validation and mapping to TCG format
  - Idempotent submission to FOP `/api/v1/sales-order` POST endpoint
  - Receipt of order acknowledgements (accepted/rejected line-level detail)

- Durable workflows:
  - Entity poll workflows with delta watermarks
  - File feed workflows with fetch, parse, validate, store
  - Order relay workflows with idempotency guarantees
  - Retry policies and parked-workflow visibility

- Security & Compliance:
  - OAuth2 client credentials (JWT) for FOP API authentication
  - SSH key pair authentication for SFTP
  - Conformance-only validation (no ETL on ingress)
  - Lineage tracking (API page or file+line)
  - Audit trail of all upserts

### Out of Scope
- Re-implementing TCG functionality; FOP is the upstream source of truth.
- Transformation or enrichment of inbound data (done by curation layer).
- Subscriber-specific interface pack execution (ISM Subscriber Layer responsibility).
- Order lifecycle management beyond relay and acknowledgement (remains in subscriber systems).
- Business logic for pricing decisions (advisory only).
- Per-subscriber code or per-tenant stacks (variability lives in packs and configuration).

---

## 3. Actors & Roles

| Actor | Role | Responsibilities |
|-------|------|------------------|
| TCG FOP System | Upstream Data Master | Publishes master data via APIs and SFTP; processes sales orders; returns acknowledgements |
| ISM Subscriber System | Downstream Consumer | Submits sales orders to ingestion layer via cell endpoint; consumes approved & enriched data |
| Ingestion Layer (tcg-service) | Data Gateway | Polls FOP APIs, pulls SFTP files, validates, stores canonical data, relays orders back to FOP |
| Curation Layer (BFF + SPA) | Enrichment & Approval | Subscribes to canonical data, enriches, approves, triggers dispatch |
| Persistence Layer (PostgreSQL) | Canonical Store | Stores all ingested data as-is in JSONB format; maintains lineage and audit |
| Temporal Engine | Durable Orchestration | Executes workflows, manages retries, watermarks, idempotency, execution history |

---

## 4. Glossary with Citations

| Term | Definition | Source Locator |
|------|-----------|-----------------|
| **FOP** | FRTS Ordering Platform — TCG's REST API and SFTP-based data service replacing legacy NSL product feeds | HLD §2 Executive Summary; BR §1.1 Related Documents |
| **Catalogue Item** | Post-NSL, universally unique product identifier (alphanumeric code) uniquely identifying a wholesale product and sales order unit | HLD §8.1.2, §12.2; BR §12, Appendix 1 |
| **NSL** | National Stock Listing — legacy product identifier being retired; retained as transitional alias during cutover | HLD §8.1.2; BR §1, Glossary |
| **Canonical Data** | Inbound TCG payloads stored unchanged in JSONB with generated key columns; never modified after ingress | HLD §8.1, §7.1 Pattern |
| **Conformance Validation** | Validation confirming inbound data matches the FOP contract; does not reshape, enrich, or transform | HLD §7.1 Pattern, Principle KDD-04; BR §8.1 |
| **Watermark** | Per-entity delta query marker (last-successful API date); prevents reprocessing and gaps beyond 7-day window | HLD §7.2.2, §20 NFR |
| **Durable Workflow** | Temporal-based workflow (EntityPollWorkflow, FileFeedWorkflow, OrderRelayWorkflow) with typed retry policies, full history, and parked-workflow visibility | HLD §13, §15 Data Flow Catalogue |
| **Cell** | One deployment of ism-service serving exactly one subscriber; holds only that subscriber's credentials and DB role; auth edge for that society | HLD §9, Glossary; BR §1.3 Stakeholders |
| **ISM** | Independent Society Member — co-operative society subscriber to the platform | BR Glossary |
| **EDN** | Electronic Delivery Notification — transactional document indicating delivered items, quantities, substitutions, shelf-life | HLD §7.3.1; BR §8.2.3 |
| **SFTP** | Secure File Transfer Protocol — transport for TCG file feeds (JSONL format, immutable landing zone) | HLD §7.3, Principles; API Guide Rules |
| **JSONL** | JSON Lines format — one valid JSON object per line; used for transactional file feeds | BR §10 API Capabilities |
| **Idempotency** | Sales order submission by store order reference prevents duplicate submissions from returning error; original acknowledgement returned | HLD §12.1 Flow, §20 NFR; BR §8.2.1 |

---

## 5. Functional Requirements

### Master Data Ingestion (via FOP API v0.4.0)

| ID | Priority | Requirement | Source Locator | Notes |
|----|----------|-------------|-----------------|-------|
| **FR-001** | P0 | Ingest Retail Products via `/api/v1/retail-products` GET endpoint with delta query parameter (`since` ISO date, max 7 days back) and pagination. Store payload as-is in canonical table keyed by Product ID. | HLD §7.2.1; API Dict: Retail Products sheet; API Guide Rules |Retail Products contain consumer-facing details (EANs, packaging, hierarchy, restrictions, tax). Paginate until response length < total property. |
| **FR-002** | P0 | Ingest Wholesale Products via `/api/v1/wholesale-products` GET endpoint with delta & pagination. Store as-is, keyed by Catalogue Item ID. Extract and index Catalogue Item as canonical product key. | HLD §8.1.1, §8.1.2; API Dict: Wholesale Products sheet | Wholesale Product is the orderable unit; contains cost prices array, supplier discount, unit conversion, pack/pallet spec, related retail product. Catalogue Item ID format: `{18-digit-article}-{UOM}` (e.g., `000000000070005494-C00`). |
| **FR-003** | P0 | Ingest Promotions via `/api/v1/promotions` GET endpoint with delta & pagination. Store payload as-is, keyed by Promotion ID. | HLD §7.2.1; API Dict: Promotions sheet; BR §8.1.3 | Includes type, period, prices, qualifying products, funding attributes. Advisory only (Principle P5). |
| **FR-004** | P0 | Ingest Suppliers via `/api/v1/suppliers` GET endpoint with delta & pagination. Store as-is, keyed by Supplier Code. | HLD §7.2.1; API Dict: Suppliers sheet | Includes supplier name, ANA code (key for direct-to-store ordering), address. |
| **FR-005** | P0 | Ingest Product Hierarchies via `/api/v1/product-hierarchies` GET endpoint with delta & pagination. Store as-is, keyed by Hierarchy Node ID. | HLD §7.2.1; API Dict: Product Hierarchies sheet; BR §8.1.5 | 6 levels of hierarchy (Level 1 = broadest, Level 6 = most specific). Parent-child relationships. |
| **FR-006** | P1 | Ingest Product Cost Prices (Wholesale Cost) embedded in Wholesale Product payloads. Store as-is, keyed by product + effective-from date. Advisory for retail purposes; never write consumer-facing price (Principle P5). | HLD §8.1.1, §7.2.1; BR §8.1.2 Prices | Cost prices have effective-dated validity windows (start/end inclusive). Minimum Alcohol Floor Price included. |

### Transactional Data Ingestion (via SFTP Pull)

| ID | Priority | Requirement | Source Locator | Notes |
|----|----------|-------------|-----------------|-------|
| **FR-007** | P0 | Pull EDN (Electronic Delivery Notifications) from TCG SFTP endpoint on configured schedule (baseline SFTP; API alternative per-feed configurable per §7.2.1). Validate JSONL line-by-line (quarantine bad lines without discarding file). Store payload as-is, keyed by Delivery Reference. Archive original file and PDF artefacts; record manifest (hash, line count). | HLD §7.3, §7.3.1; BR §8.2.3; API Dict: Delivery Notifications sheet | SFTP baseline path; API endpoint available as per-feed config alternative (DQ-002). JSONL files land in immutable object-store landing zone. Lineage: API page or file+line. |
| **FR-008** | P0 | Pull Depot Invoice files from TCG SFTP on daily schedule. Validate JSONL line-by-line. Store as-is, keyed by Invoice Reference. Link to associated Sales Order and EDN. | HLD §7.3.1; BR §8.2.4 | Invoice includes co-op company details, depot, customer, store, invoice ref, product line items with cost & VAT, total charges. |
| **FR-009** | P0 | Pull Direct Invoice files from TCG SFTP on daily schedule. Validate JSONL line-by-line. Store as-is, keyed by Invoice Reference. Link to Sales Order and delivery. | HLD §7.3.1; BR §8.2.4 | Goods delivered directly from suppliers; includes supplier details, invoice ref, customer, store, product line items. |
| **FR-010** | P1 | Pull Claims files (Depot Claims, Direct Claims) from TCG SFTP on daily schedule. Validate JSONL. Store as-is, keyed by Claim Reference. Stage for ISM upload (outbound claim settlement). | HLD §7.3.1; BR §8.2.5, §8.2.5.1 | Claims are settlement documents (credit/debit notes) for disputes (returned goods, discrepancies). Include TCG/ISM co-op details, claim type, original invoice/delivery ref, product detail, claim amount. |
| **FR-011** | P1 | Pull Condition Contract files from TCG SFTP on daily schedule. Validate JSONL. Store as-is, keyed by Contract Reference. | HLD §7.3.1; BR §8.2.6.1 | Settlement information for promotions and supplier funds. Includes co-op details, customer, settlement details, condition contract ref, promotion ref, product details, settlement value. |
| **FR-012** | P1 | Pull Promotional Sales Volumes files from ISM cells on weekly schedule (outbound). Validate JSONL. Stage for submission to TCG. | HLD §7.3.1; BR §8.2.6 | Evidence of promotional sales performance; used to support promotional funding claims. Includes customer ref, promotion ref, product code/EAN, sales date & qty. |
| **FR-013** | P0 | Pull Range / Planogram / RADR / PSA files from TCG SFTP on configured schedule. Validate JSONL. Store files in immutable object store; index metadata in Postgres. Maintain file reference and publication date. | HLD §7.3.1; BR §8.3 JDA | TCG publishes planograms and PSA (space planning) files via SFTP; RADR (range & space reports) available via Open Access portal. Files are object-store resident; metadata indexed for curation workbench display. |

### Sales Order Intake & Relay

| ID | Priority | Requirement | Source Locator | Notes |
|----|----------|-------------|-----------------|-------|
| **FR-014** | P0 | Accept sales orders from ISM subscriber cells at the auth edge (per-cell endpoint authenticated with subscriber client credentials). Support both API intake (POST) and SFTP file drop on cell's inbound channel. | HLD §11.2, §12.1; BR §8.2.1; API Dict: Sales Orders sheet | Auth edge is per-cell; credential compromise maps to exactly one society. Orders arrive in ISM's format. |
| **FR-015** | P0 | Map ISM-format sales order to TCG format: resolve products (downstream key → Catalogue Item ID), validate against current canonical data (availability, orderable status, restrictions). Idempotency key derives from store order reference; duplicate submission returns original acknowledgement. | HLD §12.1, §12.2; BR §8.2.1 | TCG format includes: societyId, submissionDate, headOfficeId, orderProcessingType (FLW/PTW), storeId, storeOrderReference (22-char unique per society), items with catalogueItemId & quantity. |
| **FR-016** | P0 | Stage TCG-format order in Postgres staging table (idempotency-keyed by store order reference, RLS-partitioned by originating subscriber). Verify uniqueness; surface existing order if duplicate detected. | HLD §8.3, §12.1; BR §8.2.1 | Staging table: `staging.sales_order` (TCG format payload + originating subscriber ID + audit metadata). |
| **FR-017** | P0 | Submit staged order to FOP via OrderRelayWorkflow (Temporal): POST `/api/v1/sales-order` with platform JWT (OAuth2 client credentials) and partnerId. Implement retry policy and circuit-breaking. Handle response (200 with acknowledgement array, 400 malformed, 4xx contract/config, 5xx backoff). | HLD §12.2, §13, §7.2.2; API Guide: Response Codes | FOP returns 200 with acknowledgement: accepted[] + rejected[] per store order reference, with item/store rejection codes. Order cutoff: 03:30 (proposal: move to 02:00). |
| **FR-018** | P0 | Persist order acknowledgement against staged order (one-to-one with original submission). Extract accepted and rejected line detail with rejection codes (e.g., Z1–Z9 item codes, Z101+ store codes per ISM design). Return acknowledgement to originating ISM through same cell, in their format, with line-level detail. | HLD §12.1, §12.3; BR §8.2.2; API Dict: Sales Orders Response sheet | Acknowledgement includes: totalItems, acceptedItems, rejectedItems, timestamp, per-store accepted/rejected arrays with line codes. Rejection codes: Z6 = wrong UOM, Z7 = not orderable, Z8 = DC capacity, Z9 = zero lead time, etc. |

### Validation, Lineage & Storage

| ID | Priority | Requirement | Source Locator | Notes |
|----|----------|-------------|-----------------|-------|
| **FR-019** | P0 | Implement conformance validation only: validate inbound payloads against committed FOP OAS 3.0 spec and FRTS data dictionary. Check JSON Schema, required fields, data types, value constraints (e.g., partnerId format, HTTP method match, response code precedence). Quarantine malformed payloads without data loss. | HLD §7.1 Pattern, Principle KDD-04; API Guide: Request Validation | No ETL on ingress: validation checks conformance, never reshapes. TCG documentation describes stored rows byte-for-byte. Contract drift is a compile error (oapi-codegen + generated validation structs). |
| **FR-020** | P0 | Store all canonical inbound data as-is in JSONB columns with extracted & indexed key columns (product ID, catalogue item, supplier code, hierarchy node, promotion ID, etc.). Canonical tables written only by tcg-service's validated upsert path; overlay never modifies canonical (Principle P2). | HLD §8.1, Principle P2, P8; BR §8 | Canonical table structure: JSONB payload + generated key columns (indexed, unique). Example: `canonical.retail_products (product_id, payload JSON, updated_at)`. |
| **FR-021** | P0 | Maintain lineage for every persisted record: API page source (entity, page number, api call timestamp) or file source (file name, SFTP path, line number, manifest hash). Store in ops.lineage table for "where is my record?" query. | HLD §8.3, Principle P8, §20 NFR Durability | Lineage allows tracing record origin through file+line or API page. Append-only audit via triggers. |
| **FR-022** | P0 | Maintain per-entity watermark: last successful API poll timestamp, last processed file timestamp, delta query window (7-day guard). Advance watermark only on successful upsert completion. Fail loudly (alert, runbook) if watermark exceeds 7-day window; prevent silent gaps. | HLD §7.2.2, §20 NFR, Risk R-2; BR §10.1 Delta Queries | Watermark re-baseline runbook: weekly full SFTP feed or single-record endpoints if API window exceeded. Alerting on watermark stall. |
| **FR-023** | P0 | Append-only audit: every canonical upsert recorded with timestamp, source (API page or file+line), validated payload hash, actor (tcg-service). Audit retained per data retention policy (finance-adjacent: 7 years default). | HLD §8.3, Principle P4, P8; §20 NFR Auditability | Triggers on canonical tables generate audit rows. Decision audit trails for enrichment & approvals live in overlay tables. |

### API Integration & Authentication

| ID | Priority | Requirement | Source Locator | Notes |
|----|----------|-------------|-----------------|-------|
| **FR-024** | P0 | Implement FOP client wrapper (generated from oapi-codegen, committed OAS 3.0 spec). Acquire OAuth2 JWT token from TCG per environment (client credentials flow, per-environment registered client). Inject partnerId on every API call. Handle declared-egress-IP routing. | HLD §7.2, §7.2.1, Principle KDD-04; API Guide: Rules, Authentication | Token acquisition: OAuth2 Client Credentials. partnerId = environment config (never hard-coded). Declared-egress-IP: NAT on integration subnet. Secret in Key Vault per environment. |
| **FR-025** | P0 | Implement delta query semantics: `GET /entity?since=ISO-date` (max 7 days back). Implement pagination: `page` query parameter. Continue requesting incrementing pages while returned array length < total property. Stop on HTTP 204 (no more data). Both single-page case (length == total) and 204 terminator must be covered in conformance pack. | HLD §7.2.2; API Guide: API Pagination; BR §10.1, §10.3 | Page size owned & tuned by FOP; platform degrades gracefully under FOP rate limiting. Response code 204 = success, no more data available. Conformance: beans worked example includes pagination round-trips. |
| **FR-026** | P0 | Handle all FOP response codes per specification: 200 (success), 204 (success, no more data), 400 (malformed request), 401/403 (auth: IP/token/client/identity), 404 (bad URI), 406 (partnerId invalid or function not allowed for partner), 500 (platform error), 505 (API version retired), 540 (temporarily disabled). 4xx alert immediately; 5xx/540 backoff with circuit-breaking. | HLD §7.2.2; API Guide: Response Codes; BR §10 | 406: logged identifiably in FOP. 4xx alerts surface as operational defects (contract or config). 5xx/540: retry on activity policy (Temporal backoff). |
| **FR-027** | P1 | Support SFTP authentication via per-environment SSH key pair (private key in Key Vault, public key registered with TCG). Implement file landing zone: immutable object store, hash & manifest recorded per file. | HLD §7.3, Principle KDD-05; §16.1 Identity | SFTP pull only (no inbound port from TCG). Files drop unmodified into landing zone. EDN-by-API available as per-feed config alternative (documented fallback if TCG pushes). |

---

## 6. Business Rules

| ID | Priority | Rule | Source Locator | Consequence if Violated |
|----|----------|------|-----------------|------------------------|
| **BR-001** | P0 | Delta API queries may not exceed 7-day window (`since` parameter must be within last 7 days). | HLD §7.2.2; API Guide Rules; BR §10.1 Delta Queries | Watermark exceeds window → request invalid → loud failure → re-baseline runbook (weekly full feed or single-record endpoints). Never silent gap. |
| **BR-002** | P0 | Page size is owned and tuned by FOP; platform does not specify page size. | HLD §7.2.2; API Guide: API Pagination | ISM-side page size negotiation or tuning breaks contract; platform must be ready for any valid page size. |
| **BR-003** | P0 | Conformance validation only; no ETL on ingress. Validation checks conformance to FOP contract; never reshapes, filters, or enriches payload. | HLD §7.1 Pattern, Principle KDD-04; BR §8.1 Master Data | Transformation belongs to curation layer. If ingress transforms, contract drift hidden; canonical rows no longer match TCG documentation. |
| **BR-004** | P0 | Canonical tables written only by tcg-service's validated upsert path. The overlay never modifies canonical rows. Any write to canonical outside this path is a defect. | HLD §8.1, Principle P2; BR §8 Canonical | If overlay or external process writes canonical, data integrity lost; upstream truth corrupted. |
| **BR-005** | P0 | Watermarks advance only on successful upsert completion. Failed activities leave watermark unchanged; replay on next run. | HLD §7.4 Durable Workflows, §20 NFR Durability; BR §10.1 Delta Queries | If watermark advances on validation failure, records skipped on retry; data loss. Temporal idempotency ensures exact-once. |
| **BR-006** | P0 | Catalogue Item is the canonical product key everywhere inside the platform. NSL is carried as an indexed transitional alias on the wholesale product row only (for lookup and legacy documents during cutover). Nothing new is keyed on NSL. | HLD §8.1.2, Principle KDD-08; BR §12 New Product Identifier | If NSL becomes primary key, cutover and legacy debt incompletely addressed. Curation screens must resolve on Catalogue Item, not NSL. |
| **BR-007** | P0 | All canonical inbound records must be traceable to source: API page number and timestamp, or file name + line number + manifest hash. "Where is my record?" is a query (lineage + operations trace). | HLD §8.3 Staging, ops, roles; Principle P8; §20 NFR Durability | If lineage missing, operations cannot debug delivery failures or trace record origin through disputes. |
| **BR-008** | P1 | Wholesale price data is advisory; platform never writes a consumer-facing retail price. | HLD Principle P5; BR §8.1.2 Prices, §10.3 Price and promotion: advisory surveillance | If platform auto-updates retail price, Principle P5 violated; ISM commercial autonomy compromised; retail strategy becomes platform-driven. |
| **BR-009** | P1 | EDN baseline path is SFTP pull; API endpoint available as per-feed configuration alternative. Both paths converge on same validate → upsert chain. | HLD §7.3, Principle P3 (One way in, one way out); BR §8.2.3 | If API and SFTP paths diverge in validation or upsert logic, record inconsistency. Same chain ensures identical persistence. |
| **BR-010** | P0 | Sales order submission idempotent by store order reference. Duplicate submission is a no-op returning original acknowledgement. | HLD §12.1, §12.2, §20 NFR Idempotency; BR §8.2.1, §8.2.2 | If not idempotent, duplicate orders land in TCG; ISM trust damaged; financial reconciliation broken. Beans worked example includes rejection & error payloads as fidelity bar. |
| **BR-011** | P0 | Order acknowledgements and rejections return via same cell as intake. Per-cell credentials ensure response reaches only originating subscriber. | HLD §11.2, §12.1 Flow; Principle P7 (Isolation in data and cells) | If cross-cell or cross-subscriber messaging used, data boundary violated; competitor visibility into peer society's orders. |
| **BR-012** | P1 | Product availability & orderable status validated at order relay time against current canonical data. Orders for discontinued, long-term out-of-stock, or restricted-qty items rejected with code Z7 (not orderable) or Z8 (DC capacity). | HLD §12.1 Flow; BR §8.2.1; API Dict: Sales Orders Response | If validation skipped, orders for unavailable items reach TCG; TCG rejects downstream; ISM disruption, trust loss. |

---

## 7. Data Entities & Canonical Schema

### Master Data Entities

| Entity | Key | Source | Format | Validation | Lineage | Notes |
|--------|-----|--------|--------|-----------|---------|-------|
| **Retail Product** | Product ID (18 digits) | FOP API `/api/v1/retail-products` | JSONB as-is | OAS 3.0, Data Dict | API page + timestamp | Consumer-facing product: EANs, packaging, hierarchy, restrictions, tax. Effective-dated retail price. |
| **Wholesale Product / Catalogue Item** | Catalogue Item ID (`{article}-{UOM}`) | FOP API `/api/v1/wholesale-products` | JSONB as-is | OAS 3.0, Data Dict | API page + timestamp | Orderable unit: cost prices array, supplier discount, unit conversion, pack spec. Primary product key. NSL alias on same row (optional, transitional). |
| **Promotion** | Promotion ID (from SAP ECC) | FOP API `/api/v1/promotions` | JSONB as-is | OAS 3.0, Data Dict | API page + timestamp | Type, period, prices, qualifying products, funding attributes. Advisory (P5). |
| **Supplier** | Supplier Code | FOP API `/api/v1/suppliers` | JSONB as-is | OAS 3.0, Data Dict | API page + timestamp | Name, ANA code (direct-to-store key), address. |
| **Product Hierarchy** | Hierarchy Node ID | FOP API `/api/v1/product-hierarchies` | JSONB as-is | OAS 3.0, Data Dict | API page + timestamp | 6 levels; parent-child relationships; level 1 = broadest, 6 = most specific. |
| **Wholesale Cost Price** | product + effective-from date | Embedded in Wholesale Product payload | JSONB as-is | OAS 3.0, Data Dict | API page + timestamp | Effective-dated cost; duty rates per HMRC excise code. Advisory for retail (P5). |

### Transactional Data Entities

| Entity | Key | Source | Format | Validation | Lineage | Notes |
|--------|-----|--------|--------|-----------|---------|-------|
| **EDN** | Delivery Reference | SFTP (baseline); API alternative | JSONL (one per line) | Per-line JSON Schema | File + line + hash | Delivered items, substitutions, quantities, expiry, depot, store, delivery date. Links to Sales Order. |
| **Depot Invoice** | Invoice Reference | SFTP | JSONL | Per-line JSON Schema | File + line + hash | Co-op company, depot, customer, store, product lines with cost & VAT, total charges. Links EDN. |
| **Direct Invoice** | Invoice Reference | SFTP | JSONL | Per-line JSON Schema | File + line + hash | Supplier-delivered goods; supplier details, customer, store, product lines, charges. |
| **Claim** | Claim Reference | SFTP | JSONL | Per-line JSON Schema | File + line + hash | Settlement document (credit/debit note) for disputes. Includes original invoice/delivery ref, product, claim reason, amount. |
| **Condition Contract** | Contract Reference | SFTP | JSONL | Per-line JSON Schema | File + line + hash | Settlement for promotions & supplier funds. Condition contract ref, promotion ref, product, settlement value. |
| **Planogram / PSA** | File Reference | SFTP | Binary (PDF/Excel) + metadata JSON | Filename, size, hash | File + SFTP path | Range & space planning files. Metadata indexed; file in object store. |

### Staging & Ops Tables

| Table | Purpose | Schema | Retention |
|-------|---------|--------|-----------|
| **staging.sales_order** | TCG-format order intake, keyed by store order reference, RLS-partitioned by subscriber | (subscriber_id, store_order_reference, payload JSONB, received_at, staged_at, status) | Until acknowledgement persisted + 90 days |
| **ops.watermark** | Per-entity delta query marker, 7-day guard, last-successful timestamp | (entity, last_successful_poll, last_api_timestamp, 7day_guard_check) | Rolling (never purged) |
| **ops.lineage** | Record origin (API page or file+line); append-only | (entity, key, source_type, source_page/file_line, manifest_hash, recorded_at) | 7 years (finance-adjacent) |
| **ops.audit** | All canonical upserts: timestamp, source, actor, payload hash | (entity, key, action, source, payload_hash, actor, audit_at) | 7 years |

---

## 8. Integrations

### Inbound Integrations (TCG → Ingestion Layer)

| System | Protocol | Direction | Data | SLA / Schedule | Notes |
|--------|----------|-----------|------|-----------------|-------|
| **FOP API** | HTTPS REST + OAuth2 JWT | Inbound (pull) | Master data (Retail Products, Wholesale, Promotions, Suppliers, Hierarchies) | Delta: on poll cycle (freq TBD, DQ-001); 7-day window guard | Contract: FOP v0.4.0 OAS. Response codes per spec. Page size TBD by FOP. |
| **TCG SFTP** | SFTP SSH key pair | Inbound (pull) | Transactional files (EDN, invoices, claims, condition contracts, planograms, PSA, weekly full master data) | Per-feed schedule (daily for invoices/claims; TBD for others) | Immutable landing zone; JSONL validation per-line. Archive originals. |
| **FOP API (Orders)** | HTTPS REST + OAuth2 JWT | Outbound (push) | Sales order submission; acknowledgement retrieval | Cutoff: 03:30 (proposal 02:00); order processing at next F&R batch | Idempotency by store order reference. Rejection codes (Z1–Z9, Z101+) per ISM design. |

### Outbound Integrations (Ingestion Layer → Downstream)

| System | Interface | Data | Direction | Notes |
|---------|-----------|------|-----------|-------|
| **PostgreSQL Canonical** | Direct SQL upsert | All master & transactional entities, keyed by business key | Write | Canonical tables (JSONB) + ops tables (watermark, lineage, audit). RLS enforced by database role. |
| **PostgreSQL Staging** | Direct SQL upsert | Staged sales orders, acknowledgements | Write | `staging.sales_order` RLS-partitioned by subscriber. Idempotency key = store order reference. |
| **Object Store (MinIO/Blob)** | S3-compatible API | Landing zone (original files), archive (PDFs, planograms), lineage | Write | Immutable landing zone; hash & manifest recorded. File retention per subscriber policy (default 7 years). |
| **Temporal** | Go SDK | Workflow dispatch | Enqueue | EntityPollWorkflow, FileFeedWorkflow, OrderRelayWorkflow. Parked workflows visible in Temporal UI. |

---

## 9. Data Flows & Journeys

### Journey 1: Master Data Ingestion → Curation → Dispatch (Happy Path)

```
1. EntityPollWorkflow polls FOP API for Retail Products since last watermark
   ↓
2. Response paginated; lines validated per OAS & data dictionary
   ↓
3. Each validated record upserted to canonical.retail_products (JSONB + key columns)
   ↓
4. Lineage recorded: API page + timestamp
   ↓
5. Watermark advanced to poll timestamp
   ↓
6. Curation layer reads canonical + subscriber overlay → enrichment panel
   ↓
7. Subscriber approves selection → approval decision audit
   ↓
8. Cell dispatch reads egress-ready view (canonical ⨯ subscription ⨯ enrichment)
   ↓
9. Pack mapping transforms to destination format
   ↓
10. Send via dispatch transport (REST/SFTP/file); receipt recorded
```

### Journey 2: Transactional File Ingestion (EDN via SFTP)

```
1. FileFeedWorkflow polls TCG SFTP for EDN files per schedule
   ↓
2. File lands unmodified in immutable object-store landing zone
   ↓
3. Hash & manifest recorded (file name, size, line count)
   ↓
4. JSONL parsed line-by-line; bad lines quarantined (no file discard)
   ↓
5. Each valid EDN upserted to canonical.edn table
   ↓
6. Lineage recorded: file name + line number + manifest hash
   ↓
7. PDF artefacts (if supplied) archived in object store
   ↓
8. Metadata indexed for curation workbench display
   ↓
9. Original archived; referenced from canonical rows
```

### Journey 3: Sales Order Relay (ISM Cell → FOP → ISM Cell)

```
1. ISM cell receives order from its store system (API POST or SFTP drop)
   ↓
2. Order arrives in ISM's format at per-cell auth edge
   ↓
3. Map ISM format → TCG format:
   - Resolve products: downstream key (captured at enrichment) → Catalogue Item ID
   - Validate against canonical (availability, orderable status, restrictions)
   ↓
4. Stage order in staging.sales_order (idempotency key = store order reference)
   ↓
5. OrderRelayWorkflow submits to FOP POST /api/v1/sales-order
   ↓
6. FOP returns 200 with acknowledgement (accepted + rejected line arrays, codes)
   ↓
7. Acknowledgement persisted against staged order
   ↓
8. Acknowledgement mapped back to ISM format
   ↓
9. Return to originating cell (same auth edge); cell delivers to store system
   ↓
10. Store order reference idempotency: duplicate submission returns original ack
```

---

## 10. Non-Functional Requirements

| Category | Requirement | Baseline | Source Locator | Acceptance Criterion |
|----------|-------------|----------|-----------------|---------------------|
| **Availability** | Platform services uptime | 99.5% monthly | HLD §20; BR §17 Success Criteria | Order relay path critical during subscriber ordering windows (02:00–03:30 area, per confirmed business clocks). Monitored & alerted. |
| **Latency — Master Data** | Change visible in curation screens | 30 minutes max | HLD §20 | Default 15-minute poll + processing within 15-minute window. Turnaround from TCG API publication to SPA screen. |
| **Latency — Order Relay** | Order relayed to FOP within intake completion | 5 minutes | HL §20 | Order staged → mapped → submitted. Acknowledgement returned to ISM within 5 minutes of FOP response. |
| **Throughput** | Sized for peak load | 10 subscribers, full-catalogue weekly refresh, peak promotional window | HLD §20; BR §10.2 Request Validation | Degrades gracefully under FOP rate limiting (back-off honoured, no data loss). No timeouts in API requests. |
| **Durability / Lineage** | No record loss; full traceability | Every persisted record traceable to source | HLD §20 NFR; Principle P8 | Landing zone immutable. Temporal buffers failures. Parked workflows + replay. Lineage query: "where is my record?" |
| **Idempotency** | Order submission idempotent | All upserts idempotent by business key; order submission by store order reference | HLD §20; BR §8.2.1, §12.1 | Duplicate API call or SFTP resubmission returns original result; no duplicate in TCG. Beans worked example covers round-trips. |
| **Recovery** | Disaster recovery SLA | RPO 1 hour, RTO 8 hours | HLD §19.3, §20 | Postgres geo-replica + object-store geo-redundancy (paired region). Temporal Cloud multi-region. Graph rebuilds from git. Failover runbooks rehearsed at UAT. |
| **Auditability** | Commercial decision audit | Every upsert attributed to actor & source; retained 7 years | HLD Principle P4, P8; §20 NFR | Append-only audit tables. Timestamp, source, payload hash, actor. Finance-adjacent retention: 7 years default, configurable per subscriber policy. |

---

## 11. Security Architecture

### Authentication

| Component | Mechanism | Details | Source Locator |
|-----------|-----------|---------|-----------------|
| **FOP API** | OAuth2 Client Credentials (JWT) | Per-environment registered client. JWT issued against platform's registered client. Secret in Key Vault. Refresh per environment. | HLD §16.1; BR §10.2 Authentication & Authorisation |
| **SFTP** | SSH key pair | Per-environment pair (private in Key Vault, public registered with TCG). All TCG-bound traffic from declared egress IP (NAT on integration subnet). | HLD §16.1, Principle KDD-05 |
| **Platform Internal** | Workload Identities | Services & cells authenticated via workload identity federation. No secrets in pipelines. Key Vault references only. | HLD §16.1 |
| **Subscriber Users** | OIDC | Entra External ID (default) or Keycloak drop-in. MFA & conditional access. Mapped to personas (Reviewer, Approver, Admin, UST Ops). | HLD §10.4, §16.1 |
| **Subscriber Systems** | Client Credentials (per-cell) | Per-subscriber OAuth2 or mTLS credentials validated by that society's cell. Credential scope = cell only; compromise = one society. | HLD §11.2, §16.1, Principle P7 |

### Data Protection & Tenancy Isolation

| Measure | Implementation | Details | Source Locator |
|---------|-----------------|---------|-----------------|
| **In Transit** | TLS everywhere | All external and internal HTTPS. SFTP SSH. | HLD §16.2 |
| **At Rest** | Encryption | Postgres (encryption enabled), object store (Blob/S3), Key Vault. Customer-managed keys supported if required. | HLD §16.2 |
| **Subscriber Data Boundary** | Cell (process) + RLS (data) | Cell holds only subscriber's credentials & DB role. RLS filters every query. No cross-subscriber aggregation surface. Competition-law boundary proven mechanically by CI isolation suite on every merge. | HLD §9 Cellular Tenancy, §16.2, Principle P7 |
| **Attack Surface** | Public HTTPS only | One public surface (web + BFF + cell auth-edge endpoints) behind cloud front door/WAF. Postgres, Temporal, Neo4j private-network only. No inbound path from TCG. | HLD §16.2 |
| **Supply Chain** | Provenance & SBOM | Go modules with checksum verification. Distroless images. SBOM per image. Signed provenance in pipeline. | HLD §16.2 |

### Compliance & Audit

| Requirement | Mechanism | Details | Source Locator |
|-------------|-----------|---------|-----------------|
| **Human Approval Gates** | Merge → Deploy → Environment | No merge to main, no environment promotion, no pack or auto-approve rule live without named human approval. | HLD Delivery-time Guardrails §16.3 |
| **Secrets Management** | Key Vault References Only | No secrets in code, logs, prompts, repos, work items. Secret scanning on every push. Leak = stop-the-line event. | HLD §16.3, Principle D1 |
| **Audit Trail** | Append-Only Tables + OTel | Every canonical upsert logged (timestamp, source, actor, payload hash). Structured logs + OpenTelemetry traces export to cloud monitoring. | HLD §20 NFR Auditability; Principle P8 |
| **Lineage Tracing** | ops.lineage Query | "Where is my record?" answered by joining lineage table to OTel traces and Temporal execution history. | HLD §8.3, Principle P8 |
| **Decisions Recorded** | Architectural Decision Records (ADRs) | Every architecturally significant choice becomes ADR with frontmatter (id, status, governs). Merged only with named approval. Governance edges in context plane derive from ADR frontmatter. | HLD §18.2 Policy, §21 Key Design Decisions; CLAUDE.md §3 |

---

## 12. Assumptions

| ID | Assumption | Mitigation | Source Locator |
|----|-----------|-----------|-----------------|
| **AS-001** | TCG delivers FOP per v0.4.0 spec: OAuth2 client credentials, delta queries (7-day window), pagination with 204 terminator, sales order POST, JSONL SFTP feeds per published schedule. | Conformance pack against TCG dev endpoints (swap from mocks config-only). Beans worked example (rejection + error payloads) is fidelity bar. | HLD §22.1 Assumptions; BR §15 Assumptions |
| **AS-002** | TCG registers platform's clients and SSH keys; issues per-environment credentials and partnerIds; dev endpoints available ~week 6. | Mocks-first (FOP + SFTP mocks generated from spec, data dictionary, beans worked example) removes upstream availability dependency from critical path. | HLD §22.1, Risk R-8 |
| **AS-003** | Each ISM can receive data through ≥1 pack transport (REST, SFTP, file, webhook) and supply enrichment elements required by their systems. | Bulk subscribe by hierarchy/supplier; auto-approve rules with named sign-off. Curation burden mitigated by bulk operations and rule-based approvals. | HLD §22.1; BR §10.2 Documentation |
| **AS-004** | Subscriber commercial teams operate curation workflow; platform does not assume fully automated flow. | ISM Subscriber Layer accepts both API intake and SFTP drop for orders; humans approve in curation workbench. Not fully automated at cutover. | HLD §22.1 |
| **AS-005** | Build is greenfield: no mandated reuse of Dataverse, Power Apps, or MDDI framework. | If reuse becomes mandatory, affected layer decision revisited by ADR (EA approval). Requirement statement runtime-agnostic. | HLD §22.1 |

---

## 13. Dependencies

| ID | Dependency | Impact | Owner | Status |
|----|-----------|--------|-------|--------|
| **D-001** | FOP spec baseline v1.0 and stable JSON Schemas (TCG committed to baseline after ISM feedback); change control thereafter via TCG steerco. | If FOP spec unstable, contract churn delays platform delivery. Generated client + validation structs make churn a compile error; mocks version with spec. | TCG Steerco | Pending (v0.4.0 in flight) |
| **D-002** | TCG cutover/transition template (master-data split, NSL-last-order / catalogue-item-first-EDN scenarios). | Shapes cutover runbooks and ISM communication plan. Clarifies dual-identifier period and switchover dates. | TCG Change Control | TBD |
| **D-003** | Confirmed operating-model clocks: order cutoff (03:30 → 02:00 proposal), master-data change windows, API call cutoff points. | Schedules for entity pollers, file feeds, order relay windows, ISM communication. | TCG / ISM Steerco | TBD (DQ-001) |
| **D-004** | FOP rate limits (raised in ISM workshops, unanswered). | Finalizes polling concurrency and back-off configuration. Platform degrades gracefully but limits not yet known. | TCG Architecture | TBD (DQ-001) |
| **D-005** | Per-subscriber destination details and credentials at onboarding time. | Drives each pack (definition.yaml, mapping.go) and cell manifest. Onboarding runbook / playbook requires this data. | ISM Onboarding Owner | Captured at cell provisioning |

---

## 14. Open Questions (DQ-xxx)

### Critical Path Blockers (Blocking = Yes)

| ID | Question | Owner | Blocking | Target Resolution | Mitigation |
|----|----------|-------|----------|-------------------|-----------|
| **DQ-001** | **Exact master data polling frequency & window.** HLD §6.2 says "frequent, regimented times"; BR §6 Core Retail Systems says frequency "XXXXXXXXXXXXX" (redacted). Which entities (Retail Products, Wholesale, Promotions, Suppliers, Hierarchies) poll every hour, every 4 hours, every 24h? | TCG Arch + ISM Steerco | No (Resolved) | ✅ Configurable poll cycle | **Decision (2026-09-16):** Proceed with configurable poll cycle per entity. No hard-coded frequency; configuration in `config.yaml` per environment. Allows tuning post-endpoint availability. |
| **DQ-002** | **EDN baseline path: SFTP or API?** HLD §7.2.1 lists API endpoint `/api/v1/delivery-notifications` (delta by `since`); BR §8.2.3 baseline is SFTP. Are both live, or API TBD? | TCG Arch | No (Resolved) | ✅ SFTP default confirmed | **Decision (2026-09-16):** Default path is SFTP (baseline); API available as per-feed configurable alternative. Both converge on same validate → upsert chain (Principle P3). FR-007 accommodates both paths. |
| **DQ-003** | **FOP rate limits (requests/sec, pagination window, concurrent pollers).** Raised in ISM LLD workshop (HLD §22.2 D-4), unanswered. Platform concurrency depends on this. | TCG APIM | No (Resolved) | ✅ Rate limit baseline set | **Decision (2026-09-16):** Assume 10 req/sec per entity, 5 concurrent pollers. Degrades gracefully under FOP limit; no silent data loss. Tuned post TCG endpoints live. Backoff & circuit-breaking per HLD §13 Durable Workflows. |
| **DQ-004** | **Promotional Sales Volumes & Condition Contract frequency.** BR §8.2.6, §8.2.6.1 say weekly (PSV) and daily (CC). Confirm schedule and trigger (timer vs event). | TCG Arch | No (Resolved) | ✅ Daily batch schedule confirmed | **Decision (2026-09-16):** Both PSV and Condition Contracts on daily batch schedule (FileFeedWorkflow). No event-driven trigger; timer-based pull per shared SFTP schedule. Aligns with daily invoice/claim cycle (FR-008, FR-010). |
| **DQ-005** | **Which entities require immediate real-time ingestion vs scheduled polling?** E.g., promotions might need faster cadence near buy-in date (week -7 in BR business clocks). | TCG Arch + ISM Steerco | No | Week 2 business clock alignment | Design assumption: all master data delta-polled (not real-time); cutoff windows for order processing confirmed in D-003. ISM communication plan clarifies ISM responsibility to poll at schedule. |

### Design Clarifications (Blocking = No)

| ID | Question | Owner | Target | Note |
|----|----------|-------|--------|------|
| **DQ-006** | **Planogram/PSA delivery via SFTP vs Open Access portal.** HLD mentions Open Access; BR §8.3 says SFTP for planograms, "RADR reports available via Open Access only." Confirm both-channel for range files. | TCG Arch | Week 2 | Platform downloads files from both sources; metadata indexed. No blocking impact. |
| **DQ-007** | **Rejection code standard.** BR §8.2.2 Order Acknowledgements lists Z1–Z9 (item) and Z101+ (store) as examples; "to be determined" in Appendix 1 Delivery Notifications. Is table finalized? | TCG Arch | Week 5 (UAT) | Beans worked example (fidelity bar) includes rejection payloads. Code mapping table needed for ISM pack design; not blocking ingestion layer. |
| **DQ-008** | **Weekly full master data feed format.** BR §8.1 "weekly full feed of master data for all ISMs, which includes Product, Price, Hierarchy, Promotion and Supplier data." JSON or JSONL? How named? Schedule? | TCG Arch | Week 3 | Used as watermark re-baseline fallback if delta window exceeded. Currently assumed JSONL per standard. File naming/schedule TBD. |
| **DQ-009** | **Outbound claims & promotional sales volumes flow.** BR §8.2.5, §8.2.6 mark these as "platform → TCG" outbound. Are these SFTP drops to TCG inbound endpoint, or POST API? | TCG Arch | Week 4 | Currently assumed SFTP file drop (mirror of inbound); API alternative possible. Does not block data ingestion layer ingress. |

---

## 15. Testable Acceptance Criteria

### Acceptance Criteria (AC-xxx) Mapping to Functional Requirements

| AC ID | Requirement | Acceptance Criteria | Oracle | Notes |
|-------|-------------|-------------------|--------|-------|
| **AC-001** | FR-001 (Retail Products API) | Given FOP API returns paginated retail products with delta query, when tcg-service polls and validates, then canonical table persisted with Product ID key, payload in JSONB, lineage recorded. On subsequent poll with advanced watermark, only changed records returned; pagination continues while array length < total; 204 terminates. | Unit test (mock FOP), Contract test (oapi-codegen conformance), Golden-file test (expected canonical row). E2E test (compose + mock FOP). | Covers single-page and multi-page cases. Conformance pack includes pagination round-trip. |
| **AC-002** | FR-002 (Wholesale Products API) | Given FOP returns wholesale products with catalogue item ID key, when ingested and validated, then canonical table keyed by Catalogue Item ID; NSL optionally indexed on same row. Catalogue Item resolved from downstream_key at order relay time. | Unit + Contract + Golden-file tests. E2E with curation bulk-subscribe by Catalogue Item. | Curation screens accept Catalogue Item lookup. Order relay resolves downstream key → Catalogue Item. |
| **AC-003** | FR-007 (EDN SFTP) | Given SFTP file lands with JSONL EDN records and one corrupted line, when FileFeedWorkflow parses, then bad line quarantined (ops.failed_lines), valid lines persisted, file manifest (hash, count, line) recorded. Lineage = file+line. | Unit test (JSONL parse), Golden-file (fixture + expected rows). E2E with compose mock SFTP. | Demonstrates line-level isolation. Archive original file. |
| **AC-004** | FR-014, FR-015, FR-016 (Order Intake & Relay) | Given ISM cell submits order in ISM format (API POST or SFTP drop), when mapped to TCG format and staged, then store order reference idempotency key prevents duplicate. Duplicate submission returns original acknowledgement. Lines mapped: downstream_key → Catalogue Item ID (lookup in canonical). | Unit test (mapping, idempotency key generation). Contract test (ingestion cell endpoint). Beans worked example (beans-sample-200-rejection.json round-trip). E2E test (compose two cells + mock FOP). | Duplicate detection verified. Rejection codes (Z1–Z9, Z101+) tested. |
| **AC-005** | FR-024, FR-025, FR-026 (API Auth & Pagination) | Given platform must acquire JWT (OAuth2 client credentials), inject partnerId on all calls, implement delta (since ISO date, max 7 days), and pagination (page, 204), when all requests made, then contract tests pass. 4xx errors alert immediately; 5xx/540 retry per policy. | Contract tests (oapi-codegen vs committed OAS). Response code coverage (200, 204, 400, 401, 403, 404, 406, 500, 505, 540). Mock FOP fidelity includes all code paths. | Conformance pack runs on TCG dev endpoints at week 6. No production until conformance green. |
| **AC-006** | FR-019 (Conformance Validation) | Given inbound payload with one invalid field (wrong type, missing mandatory, out-of-range value), when validation runs, then record quarantined (ops.validation_failures). Payload schema matches OAS 3.0 spec committed to contracts/. | Unit test (JSON Schema validation against committed spec). Golden-file test (fixture → expected quarantine). Re-validation after spec upgrade is a query (not migration). | Supports contract churn as compile error. No data lost; quarantine visible for ops review. |
| **AC-007** | FR-022 (Watermark & 7-Day Guard) | Given entity poll succeeds at time T, when watermark advanced to T and next poll uses T as since parameter, then delta window respected. If outage causes watermark to exceed 7 days back, when next poll attempted, then loud failure (alert + runbook). Re-baseline runbook: fetch via single-record endpoint or weekly full feed. | Unit test (watermark logic, 7-day guard check). Integration test (mock FOP returns 7-day-invalid error). Alerting test (ops dashboard). | Watermark stall observable; never silent gap. Runbook tested in UAT. |
| **AC-008** | BR-010 (Order Idempotency) | Given order submitted twice with same store order reference, when second submission received, then no duplicate in TCG; original acknowledgement returned. Beans worked example fidelity bar: 200-rejection and 400-error payloads covered. | Integration test (OrderRelayWorkflow + mock FOP). Bean-fidelity conformance pack. E2E test (compose, submit duplicate, verify no duplicate in mock TCG logs). | Duplicate detection per store order reference. Idempotency key = (societyId, storeOrderReference). |
| **AC-009** | BR-007 (Lineage Tracing) | Given any persisted record (master data, transaction, staged order), when "where is my record?" query runs, then lineage returned (API page + timestamp OR file+line + manifest hash). Lineage query joins ops.lineage → OTel trace → Temporal execution history. | Unit test (lineage insert on canonical upsert). Integration test (lineage query). E2E test (trace record end-to-end from FOP API to Postgres to cell dispatch). | Ops debugging scenario: ISM reports missing shipment; trace order through acceptance & dispatch. |
| **AC-010** | BR-001 (7-Day Delta Guard) | Given delta query with since parameter > 7 days old, when FOP API called, then 4xx error returned (contract violation). When error received, when re-baseline runbook executed (weekly full feed or single-record endpoints), then data gap recovered without silent loss. | Contract test (FOP OAS guard). Integration test (watermark re-baseline). E2E test (simulate extended outage, verify loud failure + runbook recovery). | R-2 mitigation: no silent gaps. Runbook documented in iac/. |

---

## 16. Traceability to Source Documents

### Discussion Coverage Matrix (Document Section → BRD Section Mapping)

| Document | Section | Topic | BRD Sections | Coverage |
|----------|---------|-------|--------------|----------|
| **HLD v4.0** | §2 Executive Summary | Architecture overview, two-premise design (agentic, greenfield) | §1 Purpose, §3 Actors, §9 Integrations | Complete |
| | §5 Principles (P1–P8, D1–D9) | Platform & agent-first principles | §6 Business Rules BR-001–BR-012 | Complete (P2, P3, P5, P8 mapped to BR-004, BR-009, BR-008, BR-007) |
| | §7 TCG Integration Layer | Master data via API, SFTP, durable workflows | §5 FR-001–FR-027, §8 Data Entities, §9 Journeys | Complete |
| | §8 Persistence | Canonical JSONB, RLS, staging, audit | §7 Data Entities, §8 Integrations | Complete (FR-020, FR-021) |
| | §13 Durable Execution | Temporal workflows, retry, idempotency, parked workflows | §5 FR-017, §6 BR-005, §10 NFR Idempotency | Complete |
| | §16 Security Architecture | Auth, encryption, tenant boundary | §11 Security, §6 BR-011 | Complete |
| | §20 Non-Functional Requirements | Availability, latency, durability, auditability | §10 NFR | Complete |
| | §22.1–22.3 Assumptions, Dependencies, Risks | TCG spec, rate limits, ISM readiness | §12 Assumptions, §13 Dependencies, §14 Open Questions | Complete |
| **BR v28.07.26** | §8.1 Master Data (Products, Prices, Promotions, Suppliers, Hierarchy) | Detailed product attributes, HFSS, allergens, cost prices | §5 FR-001–FR-006, §8 Data Entities | Complete |
| | §8.2 Transactional Data (Sales Orders, EDNs, Invoices, Claims) | Order formats, acknowledgements, delivery notifications, invoicing | §5 FR-007–FR-018, §8 Data Entities | Complete |
| | §8.3 Range & Space (JDA Planograms, RADR, PSA) | Planogram files, space planning, range calendars | §5 FR-013, §8 Data Entities | Complete |
| | §10 API Capabilities & Behaviour | Delta queries, auth, pagination, rate limiting, validation, versioning, error handling | §5 FR-024–FR-026, §6 BR-001, §14 DQ-003 | Complete (DQ-001, DQ-003 unresolved) |
| | §11 Business Clocks | Range events, promotions, price changes, ordering window (03:30 cutoff proposal 02:00) | §12 AS-001, §13 D-003, §14 DQ-005 | Aligned; DQ-001 exact frequency TBD |
| | §12 New Product Identifier Structure | Catalogue Item as post-NSL key, 22-char format, Standard Product Code | §6 BR-006, §8 Data Entities | Complete |
| | §13 Integration Requirements (Appendix 1) | 15 data flows: master data APIs/files, sales order API, EDN/invoice/claim SFTP, promotional volumes, condition contracts | §5 FR-001–FR-018, §9 Integrations | Complete |
| | §15 Assumptions, §16 Dependencies | ISM readiness, API maturity, transition period | §12 Assumptions, §13 Dependencies | Complete |
| **API Guide v0.4.0** | Rules & Response Codes | partnerId validation, HTTP response codes (200, 204, 400, 401, 403, 404, 406, 500, 505, 540) | §5 FR-026, §6 BR-001 | Complete |
| | API Pagination | Page parameter, 204 terminator, single-page optimization | §5 FR-025, §6 BR-002 | Complete |
| **API Data Dictionary v1.1** | Retail Products, Wholesale Products, Promotions, Suppliers, Hierarchies, Sales Orders, EDN | Field definitions, data types, mandatory/optional, examples, ISM guidance | §5 FR-001–FR-018, §8 Data Entities | Complete (all sheets referenced) |
| **Architecture Diagram** | System Context | Three layers (Data Sources & Ingestion, Processing & Storage, Consumption) + Temporal, Event Hub, KV | §3 Actors, §9 Integrations, §1 Purpose | Complete |

---

## 17. Risks & Mitigations

| ID | Risk | Probability | Impact | Mitigation | Owner |
|----|------|-----------|--------|-----------|-------|
| **R-001** | FOP contract churn between v0.4.0 and baseline (v1.0) forces rework. | Medium | High | Generated client + validation structs make churn a compile error in one package; mocks version with spec; canonical stores as-is so churn cost stays at edge. Track TCG change control. | TCG Architecture |
| **R-002** | 7-day delta window exceeded after outage, causing silent gaps. | Medium | Critical | Watermark-stall alerting; loud workflow failure into re-baseline runbook (weekly full feed or single-record endpoints); never a silent invalid request. Alerts wired to operations dashboard. | UST Operations |
| **R-003** | Duplicate or lost orders during cutover damage trust. | Low | Critical | Idempotency by store order reference proven by oracle; staged orders with full acknowledgement audit; per-ISM reviewed-relay option; rejection payloads first-class in mocks and tests. Beans worked example fidelity bar. | UST Testing & QA |
| **R-004** | Temporal workflow-determinism rules trip up early agent-written workflows. | Medium | Medium | Determinism rules carried in tcg-service and ism-service DPDs; replay tests in CI; walking skeleton exercises them in anger in week one. | UST Engineering |
| **R-005** | Cell sprawl: per-subscriber manifests drift into snowflake configuration. | Low | Medium | Manifests schema-validated against one struct; cells are identical binaries; unknown fields fail the build. Divergence beyond schema structurally impossible. | UST Engineering |
| **R-006** | Knowledge graph drifts from code or quietly stops rebuilding. | Low | Medium | D9 makes staleness a build failure (SHA check); freshness oracle runs on every merge; no hand-edit path exists. | UST CI/CD |
| **R-007** | Curation burden underestimated by subscriber commercial teams. | Medium | Medium | Bulk subscribe by hierarchy/supplier; auto-approve rules with named sign-off; blind-approval metrics surfaced in console. ISM training during cutover. | ISM Onboarding |
| **R-008** | TCG dev endpoints slip past week 6. | Medium | High | Mocks-first is plan of record, not contingency (KDD-18); swap is configuration only; UAT rehearsal replans against endpoint availability. | TCG Project Mgmt |
| **R-009** | Competition-law challenge to shared-platform handling of subscriber data. | Low | Critical | Two-wall partition (cells + RLS) designed in from day one with CI proof; no cross-subscriber surface anywhere; legal review of partition design before canary onboarding. | UST Legal & Architecture |
| **R-010** | Watermark clock skew or timezone issues cause boundary violations. | Low | Medium | Watermark uses UTC ISO 8601 (strict); platform runs in UTC; tests cover DST transitions. Monitoring for skew. | UST Engineering |

---

## 18. Implementation Approach

The Data Ingestion Layer is built as part of the **Walking Skeleton** (Stage 1–2 of delivery) per HLD §23:

**Stage 0 (Week 1): Foundations**
- Repos provisioned; seed pack dropped.
- Compose stack (Postgres, Temporal dev, MinIO, mock-FOP, mock-SFTP).
- CI with oracle skeleton; graph extractors in place.

**Stage 1 (Weeks 1–2): Walking Skeleton**
- Wholesale Product end-to-end: poll → canonical → subscribe → cell dispatch.
- Order-relay slice on Temporal.
- Lineage answers "where is my record?"
- Isolation suite passes across two synthetic cells.

**Stage 2 (Weeks 2–5): Master Data Breadth**
- Remaining entities (Retail Products, Promotions, Suppliers, Hierarchies).
- Watermark hardening + re-baseline runbook.
- Cost & promotion surveillance advisory.

**Stage 3 (Weeks 4–7): Transactional + Files**
- FileFeedWorkflows for EDN, invoices, claims, condition contracts, planograms.
- Outbound claim/PSV files.
- Finance-bound packs.

**Stage 4 (Weeks 6–9): Order Return Path**
- Cell intake (file + API).
- OrderRelayWorkflow hardening.
- Acknowledgement and rejection hand-back.

**Stage 5 (Week 8+): Conformance & Cutover**
- Conformance pack against TCG dev endpoints (config swap).
- UAT rehearsal.
- Cutover + rollback runbooks.
- Canary cell onboarding playbook.

---

## 19. Success Criteria

| Criterion | Measurement | Target | Source Locator |
|-----------|-------------|--------|-----------------|
| **Master data freshness** | Time from TCG publication to SPA visibility | ≤ 30 minutes (default 15-min poll + 15-min processing) | HLD §20; BR §17 |
| **Order relay latency** | Time from intake to FOP submission | ≤ 5 minutes | HLD §20; BR §17 |
| **Zero data loss** | Acknowledged inbound records persisted | 100% (landing zone immutable, Temporal buffer, parked workflows + replay) | HLD §20 NFR Durability |
| **Idempotency** | Duplicate order submissions handled correctly | 100% (no duplicate in TCG; original ack returned) | Beans worked example + conformance pack |
| **Lineage traceability** | "Where is my record?" answerable for all records | 100% (ops.lineage query + OTel trace + Temporal history) | HLD Principle P8; §20 NFR Auditability |
| **Conformance** | All payloads conform to FOP contract | 100% (validated on ingress; quarantine failures) | Conformance pack; contract tests |
| **Availability** | Platform uptime during critical ordering windows | 99.5% monthly; alerting on order relay path during 02:00–03:30 area | HLD §20; BR §17 |
| **Watermark guard** | 7-day delta window never silently violated | 100% (loud failure + re-baseline runbook; never silent gap) | HLD §20; Risk R-2 mitigation |

---

## 20. Glossary (Consolidated)

[See §4 above for full glossary with citations.]

---

## Appendix A: Mapping to Delivery Driving Brief Principles (A1–A7)

The Data Ingestion Layer adheres to the AI-First Delivery Driving Brief:

- **A1: Specs are the product.** Every FR/BR/AC-xxx in this BRD becomes an agent-spec on the ADO board. Agents generate implementation to spec. Humans gate.
- **A2: Machine-checkable acceptance.** Every AC-xxx names the oracle (unit test, contract test, golden-file test, E2E test, conformance pack) that defines done.
- **A3: Code as instructions.** All validation, workflow, retry logic, lineage tracking, audit, and observability are in code (Go, SQL, Terraform, git). No portal-driven behavior.
- **A4: Testable.** Walking skeleton (Stage 1) proves the pattern; oracle suite (unit, golden, contract, isolation, E2E) runs on every merge. Agents write the tests alongside implementation.
- **A5: Observable.** Structured logs, OTel traces, Temporal execution histories, lineage queries, and per-flow dashboards answer ops questions from the terminal.
- **A6: Documented.** DPDs (Developer Platform Docs) for tcg-service and packs; ADRs for design decisions. KDD-04 (no ETL), KDD-05 (SFTP pull), KDD-11 (Temporal) are seeded as accepted ADRs.
- **A7: Human gates.** No merge to main, no environment promotion, no pack or auto-approve rule live without named human approval. Secrets never in context. Leaks = stop-the-line.

---

**Document Status: Draft — Pending Domain Review**

This BRD is ready for domain review by TCG Architecture, ISM representatives, UST Technical Leadership, and the Delivery Team. All critical-path blockers (DQ-001, DQ-003, DQ-004) are flagged and assigned to TCG Architecture for resolution by Week 3 detailed design.

**Next Steps:**
1. Circulate to stakeholders (TCG steerco, ISM lead, UST EA, delivery team lead).
2. Capture feedback and resolve DQ-xxx via comments or design sessions.
3. Approve BRD as input to Sprint Planning (Day 2).
4. Generate agent-specs on ADO board from FRs/BRs/ACs.
5. Seed walking skeleton (Stage 1) by end of Week 1.

---

**Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>**
