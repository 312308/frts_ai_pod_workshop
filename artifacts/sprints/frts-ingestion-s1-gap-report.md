# FRTS Data Ingestion Layer — Sprint 1 Gap Analysis Report

**Sprint ID:** `frts-ingestion-s1`  
**Status:** Proposed (Pending Approval)  
**Date Generated:** 2026-09-16  
**Sprint Duration:** 2 weeks (Weeks 1–2)  
**BRD Source:** `artifacts/brd/data-ingestion-layer-BRD-draft.md` (✅ Accepted)

---

## 1. Open Design Questions (DQ-xxx) — Status & Blocking Impact

### DQ-001: Master Data Polling Frequency & Window

**Question:** Exact polling frequency for FOP entities (Retail Products, Wholesale, Promotions, Suppliers, Hierarchies). Hourly, 4-hourly, daily?

**Source:** BRD §14, HLD §6.2 ("frequent, regimented times"), BR §6 Core Retail Systems (redacted)

**Blocking?** **No** — Resolved by decision (2026-09-16)

**Decision:** **Configurable poll cycle per entity**. No hard-coded frequency; configuration in `config.yaml` per environment. Allows tuning post-endpoint availability (week 6+).

**Impact on Sprint 1:**
- ✅ SP-06 (Entity Poll Workflow) unblocked; workflow reads poll frequency from config
- ✅ Schedule trigger in Temporal uses configurable interval
- ✅ Default config: {retail: 4h, wholesale: 4h, promotions: 24h, suppliers: 24h, hierarchies: 24h}; tuned post-TCG feedback

**Testing Strategy:** Test with mock FOP; validate frequency can be changed without code change (config-only swap)

**Mitigation:** Document default config in DPD; add ops runbook for tuning poll frequency post-endpoint availability

---

### DQ-002: EDN Baseline Path (SFTP vs API)

**Question:** Electronic Delivery Notification baseline: SFTP pull or API endpoint?

**Source:** BRD §14, HLD §7.2.1 (API endpoint `/api/v1/delivery-notifications`), BR §8.2.3 (SFTP baseline)

**Blocking?** **No** — Resolved by decision (2026-09-16)

**Decision:** **SFTP baseline confirmed; API available as per-feed configurable alternative**. Both paths converge on same validate → upsert chain (Principle P3).

**Impact on Sprint 1:**
- ✅ SP-07 (Wholesale Products) unblocked; uses API (confirmed as available)
- ⏸ EDN file-feed story (FR-007) **out-of-scope for Sprint 1** (Stage 2, week 3)
- ✅ Order relay path (FR-014–018) unblocked; uses API for product validation

**Testing Strategy:** Mocks support both SFTP and API paths in Stage 0; swap configuration per feed

**Mitigation:** Document both paths in DPD; ensure validate → upsert chain identical for both (code review gate)

---

### DQ-003: FOP Rate Limits

**Question:** FOP rate limits (requests/sec, pagination window, concurrent pollers). Raised in ISM LLD workshop; unanswered.

**Source:** BRD §14, HLD §22.2 (D-4)

**Blocking?** **No** — Resolved by assumption (2026-09-16)

**Decision:** **Assume 10 req/sec per entity, 5 concurrent pollers**. Degrades gracefully under FOP limit; no silent data loss. Tuned post TCG endpoints live.

**Impact on Sprint 1:**
- ✅ SP-05 (Delta & Pagination handler) unblocked; implements backoff on rate limit (429 Retry-After header)
- ✅ SP-06 (Entity Poll Workflow) unblocked; workflow respects rate limit via circuit-breaker
- ✅ SP-13 (Order Relay Workflow) unblocked; submit rate capped at 5 concurrent
- ✅ Configuration tunable without code change (concurrency limit in config)

**Testing Strategy:** Mock FOP returns 429 (Retry-After) on concurrent threshold; verify backoff and retry

**Mitigation:** 
- Implement rate-limit header parsing (Retry-After, RateLimit-Remaining)
- Circuit-breaker trip on persistent 429 (prevents cascading)
- Ops dashboard shows current rate-limit state per entity
- Post-TCG endpoints: tune config based on actual FOP limits

---

### DQ-004: Promotional Sales Volumes & Condition Contract Frequency

**Question:** PSV & Condition Contract schedule. Weekly (PSV) and daily (CC) per BR §8.2.6, §8.2.6.1. Confirm schedule and trigger (timer vs event).

**Source:** BRD §14, BR §8.2.6, §8.2.6.1

**Blocking?** **No** — Resolved by decision (2026-09-16)

**Decision:** **Both PSV and Condition Contracts on daily batch schedule** (FileFeedWorkflow). No event-driven trigger; timer-based pull per shared SFTP schedule. Aligns with daily invoice/claim cycle (FR-008, FR-010).

**Impact on Sprint 1:**
- ⏸ PSV & Condition Contract file-feed stories (FR-011, FR-012) **out-of-scope for Sprint 1** (Stage 3, weeks 4–5)
- ✅ Order relay path (FR-014–018) unblocked; does not depend on PSV/CC

**Testing Strategy:** Mock SFTP provides PSV and Condition Contract files on timer; verify parse and quarantine bad lines

**Mitigation:** Document schedule in DPD; add calendar reminder for ISM comms re: daily batch window

---

### DQ-005: Real-Time vs Scheduled Ingestion

**Question:** Which entities require immediate real-time ingestion vs scheduled polling? Promotions near buy-in date (week -7 business clocks)?

**Source:** BRD §14

**Blocking?** **No** — Design assumption carries forward

**Decision:** **All master data delta-polled (not real-time)**. Order processing cutoff windows confirmed in dependency D-003 (TBD). ISM communication plan clarifies ISM responsibility to poll at schedule.

**Impact on Sprint 1:**
- ✅ All poll workflows use schedule-based approach (EntityPollWorkflow, no event-driven ingestion)
- ⏸ Real-time ingestion (webhook, event-stream) **out-of-scope** (Post-Stage 5, if business requirements change)
- ✅ Order relay path (time-sensitive) uses Temporal, not polling

**Testing Strategy:** Test confirms no event-driven workflows in Stage 1; poll schedule respected

**Mitigation:** ISM onboarding runbook clarifies expected visibility latency (≤30 min for master data, ≤5 min for orders)

---

### DQ-006: Planogram/PSA Delivery (SFTP vs Open Access Portal)

**Question:** Planogram/PSA delivery via SFTP vs Open Access portal. HLD mentions Open Access; BR §8.3 says SFTP for planograms, "RADR reports available via Open Access only."

**Source:** BRD §14, HLD, BR §8.3

**Blocking?** **No** — Out-of-scope for Sprint 1

**Target Resolution:** Week 2 (design clarification, no code impact)

**Impact on Sprint 1:** None (FR-013 out-of-scope for Stage 1)

**Mitigation:** Spike: identify Open Access portal API, evaluate both-channel download in Stage 3 (week 3+)

---

### DQ-007: Rejection Code Standard (Z1–Z9, Z101+)

**Question:** Rejection codes for order acknowledgements. BR §8.2.2 lists Z1–Z9 (item) and Z101+ (store) as examples; "to be determined." Is table finalized?

**Source:** BRD §14, BR §8.2.2, Appendix 1

**Blocking?** **No** — Design flexibility in Sprint 1; code table TBD week 5 UAT

**Target Resolution:** Week 5 (Beans worked example fidelity bar)

**Impact on Sprint 1:**
- ✅ SP-11 (Order mapping) unblocked; rejection code mapping logic implemented generically
- ✅ SP-15 (Order validation) unblocked; Z7 (not orderable), Z8 (DC capacity) coded; others mapped via config table
- ✅ Code lookup table externalized (codes.yaml or database); no hardcoding

**Testing Strategy:** Beans worked example includes known rejection payloads; test with mock FOP responses

**Mitigation:**
- Rejection code enum generated from `codes.yaml` (committed to contracts/)
- Code table updatable without code change (configuration-driven)
- UAT phase (week 5) finalizes table; merge config change before production cutover

---

### DQ-008: Weekly Full Master Data Feed Format

**Question:** Weekly full master data feed format. BR §8.1: "weekly full feed of master data for all ISMs, which includes Product, Price, Hierarchy, Promotion and Supplier data." JSON or JSONL? How named? Schedule?

**Source:** BRD §14, BR §8.1

**Blocking?** **No** — Used as watermark re-baseline fallback; Sprint 1 assumes 7-day window honored

**Target Resolution:** Week 3 (design clarification)

**Impact on Sprint 1:**
- ✅ SP-03 (Watermark & 7-day guard) unblocked; re-baseline runbook references weekly full feed as fallback (assumed JSONL per standard)
- ⏸ Weekly full-feed download (FR-022 re-baseline) **out-of-scope for Sprint 1** (alert → runbook trigger; operator-initiated)

**Testing Strategy:** Simulate watermark exceeding 7-day window; verify alert + runbook reference appears

**Mitigation:**
- Document expected format in iac/watermark-rebase-runbook.md (file naming, schedule, SFTP path)
- Post-TCG endpoint clarification: implement weekly-full-feed download as fallback activity (Stage 3)

---

### DQ-009: Outbound Claims & Promotional Sales Volumes Flow

**Question:** BR §8.2.5, §8.2.6 mark Claims and PSV as "platform → TCG" outbound. Are these SFTP drops to TCG inbound endpoint, or POST API?

**Source:** BRD §14, BR §8.2.5, §8.2.6

**Blocking?** **No** — Out-of-scope for Sprint 1 (outbound; order relay is inbound)

**Target Resolution:** Week 4 (design clarification)

**Impact on Sprint 1:** None (outbound claims/PSV in Stage 3–4)

**Mitigation:** Spike: align with TCG on expected outbound format (SFTP drop vs API); document in Stage 3 planning

---

## 2. Open Critical Questions Not Blocking Sprint 1

| DQ ID | Question | Owner | Target | Sprint Impact | Mitigation |
|-------|----------|-------|--------|---------------|-----------|
| **DQ-001** | Polling frequency | TCG + ISM | Resolved (configurable) | ✅ Unblocked | Config-driven, tuned post-week 6 |
| **DQ-002** | EDN baseline | TCG | Resolved (SFTP default) | ✅ Unblocked | Both paths converge; Stage 2 |
| **DQ-003** | FOP rate limits | TCG APIM | Resolved (10 req/sec assumption) | ✅ Unblocked | Tuned post-endpoints; backoff implemented |
| **DQ-004** | PSV/CC schedule | TCG | Resolved (daily batch) | ✅ Unblocked | Stage 3 file-feed |
| **DQ-005** | Real-time vs scheduled | Business | Design assumption | ✅ Unblocked | All scheduled in Stage 1 |
| **DQ-006** | Planogram delivery | TCG | Week 2 design | ⏸ Out-of-scope | Stage 3 spike |
| **DQ-007** | Rejection codes | TCG | Week 5 UAT | ✅ Unblocked | Code table externalized; UAT finalizes |
| **DQ-008** | Weekly full-feed format | TCG | Week 3 design | ⏸ Out-of-scope (fallback) | Runbook reference; Stage 3 impl |
| **DQ-009** | Outbound flow | TCG | Week 4 design | ⏸ Out-of-scope (outbound) | Stage 3–4 planning |

---

## 3. Risk Mitigations in Sprint 1

| Risk ID | Risk | Probability | Impact | Sprint 1 Mitigation | Owner |
|---------|------|----------|--------|-------------------|-------|
| **R-001** | FOP contract churn v0.4.0 → v1.0 | Medium | High | Generated client + mocks; contract drift = compile error | TCG Architecture |
| **R-002** | 7-day delta window exceeded (silent gap) | Medium | Critical | Watermark-stall alerting (SP-03); loud failure; re-baseline runbook | UST Operations |
| **R-003** | Duplicate/lost orders damage trust | Low | Critical | Idempotency by store order reference (SP-12, SP-16); Beans worked example | UST Testing |
| **R-004** | Temporal workflow determinism rules trip up | Medium | Medium | Determinism rules in tcg-service DPD; replay tests in SP-13 | UST Engineering |
| **R-005** | Cell sprawl / snowflake config | Low | Medium | Cell manifests schema-validated; unknown fields fail build | UST Engineering |
| **R-008** | TCG dev endpoints slip past week 6 | Medium | High | Mocks-first is plan of record (Stage 0); real endpoints optional | TCG Project Mgmt |
| **R-009** | Competition-law challenge to shared platform | Low | Critical | Two-wall partition (cells + RLS) designed from day one; CI proof; legal review pre-canary | UST Legal & Architecture |

**Sprint 1 Mitigations:**
- ✅ **Mocks-first** (Stage 0 work): FOP API mock, SFTP mock, Temporal dev in compose
- ✅ **Contract-first** (oapi-codegen): drift = compile error; no runtime surprise
- ✅ **Alerting wired** (SP-03, SP-20): watermark stall + 4xx errors → OTel + ops dashboard
- ✅ **Isolation suite** (SP-18): CI gate; two-cell partition proven on every merge
- ✅ **Determinism tests** (SP-13): Temporal replay validates workflow restartability

---

## 4. Out-of-Scope for Sprint 1 (Stage 2–5 Scope)

### Stage 2 (Weeks 2–5): Master Data Breadth & Watermark Hardening

| FR/BR | Story | Title | Impact | Reason |
|-------|-------|-------|--------|--------|
| **FR-001** | SP-19 (hold for stretch) | Retail Products Foundation | Low (not for order relay) | Walking skeleton focuses on Wholesale (orderable unit); Retail is consumer-facing, not order-critical |
| **FR-003** | — | Ingest Promotions API | Low (advisory) | Included for breadth (Stage 2); supports surveillance, not order logic |
| **FR-004** | — | Ingest Suppliers API | Low | Included for breadth (Stage 2); supports sourcing, not critical path |
| **FR-005** | — | Ingest Product Hierarchies | Low | Included for breadth (Stage 2); supports curation, not critical path |
| **FR-006** | — | Ingest Cost Prices | Low (advisory) | Cost prices embedded in Wholesale payload; advisory only (Principle P5) |
| **DQ-008** | — | Weekly Full-Feed Watermark Re-baseline | Medium | Used as fallback when 7-day window exceeded; operator-initiated runbook (Stage 2, week 3) |

### Stage 3 (Weeks 3–7): Transactional Files & Lineage

| FR/BR | Story | Title | Impact | Reason |
|-------|-------|-------|--------|--------|
| **FR-007** | — | Pull EDN via SFTP | Medium | Transactional file ingestion; critical for delivery tracking but not for order submission |
| **FR-008** | — | Pull Depot Invoices | Medium | Transactional file ingestion |
| **FR-009** | — | Pull Direct Invoices | Medium | Transactional file ingestion |
| **FR-010** | — | Pull Claims (Depot/Direct) | Medium | Transactional file ingestion |
| **FR-011** | — | Pull Condition Contracts | Medium | Transactional file ingestion |
| **FR-012** | — | Pull Promotional Sales Volumes (outbound) | Medium | Outbound staging (ISM cells → platform) |
| **FR-013** | — | Pull Planogram/PSA/RADR Files | Low | Range & space planning; curation support, not order-critical |
| **DQ-006** | — | Planogram via SFTP vs Open Access Portal | Low | Design clarification (week 2); implementation (Stage 3) |
| **DQ-009** | — | Outbound Claims/PSV Flow | Medium | Outbound submission path (post-order); Stage 3–4 |

### Stage 4 (Weeks 6–9): Order Return Path & Conformance

| FR/BR | Story | Title | Impact | Reason |
|-------|-------|-------|--------|--------|
| Conformance pack | — | Test against real TCG endpoints | Critical | Real endpoints available ~week 6 (per AS-002, D-001) |
| UAT runbooks | — | Cutover + rollback runbooks | Critical | Rehearsed with ISM teams; week 7+ |
| Canary cell | — | Onboarding playbook for first ISM | Critical | After conformance green; week 8+ |

### Stage 5+ (Future, Optional)

| Scope | Title | Reason |
|-------|-------|--------|
| Event-driven ingestion | Real-time webhooks for high-priority entities | Out-of-scope; scheduled polling adequate for current SLA (30 min master data, 5 min orders) |
| Multi-region replication | Geo-disaster recovery (RPO 1h, RTO 8h) | Infrastructure / Terraform scope; not application scope (Stage 5+) |
| Performance tuning | Pagination size negotiation, connection pooling | Post-Stage 1 validation; baseline performance meets SLA |
| Advanced features | GraphQL for curation, real-time notifications, webhooks | Post-walking skeleton; not in 2-week scope |

---

## 5. Dependency & Blocker Resolution Status

### Critical-Path Dependencies (All Resolved)

| Dependency | Blocker? | Resolution | Impact on Sprint 1 |
|------------|----------|-----------|-------------------|
| **FOP v0.4.0 OAS spec baseline** | No | Spec in `contracts/fop-oas-v0.4.0.yaml` (committed); mocks generated from spec | Unblocked (contract-first approach) |
| **Temporal framework & SDK** | No | Temporal SDK available; Temporal dev mode in compose | Unblocked (start week 1) |
| **PostgreSQL + RLS support** | No | Postgres 13+ supports RLS; schema can be drafted from BRD | Unblocked (start week 1) |
| **MinIO for object store** | No | S3-compatible MinIO available; no external endpoint required | Unblocked (compose-based, no TCG dependency) |
| **FOP JWT/OAuth2 mocks** | No | Mocks generated from OAS; token endpoint mocked | Unblocked (start week 1) |
| **FOP rate limits** | No | Assumption: 10 req/sec, 5 concurrent; tuned post-endpoints | Unblocked (conservative assumption, no code impact) |
| **Master data polling frequency** | No | Configurable per entity; no hard-coded frequency | Unblocked (configuration-driven) |
| **EDN baseline path (SFTP vs API)** | No | SFTP baseline confirmed; API alternative for Stage 3 | Unblocked (FR-007 moved to Stage 2, not critical path) |

---

## 6. Testing & Validation Gap Analysis

### Test Coverage Plan (Per Story)

All stories include ≥1 oracle. Oracles executed on every merge (CI gate).

| Test Type | Stories Covered | SLA | Gap Mitigation |
|-----------|-----------------|-----|-----------------|
| **Unit tests** | All 20 stories | 90%+ line coverage | Code review gate; coverage metrics in CI |
| **Contract tests** | SP-01, SP-02, SP-05 (API) | 100% response code paths | oapi-codegen validates; mocks cover all codes |
| **Golden-file tests** | SP-07, SP-08, SP-11, SP-14 (data shape) | 100% canonical row structure | Testdata fixtures in repo; reviewed on change |
| **Integration tests** | SP-08, SP-12, SP-16, SP-17 (DB) | All RLS policies + triggers | Postgres testcontainer; CI gate |
| **E2E tests** | SP-13, SP-14, SP-18 (full path) | Happy path + rejection + error | Docker Compose; mocks for all external deps |
| **Isolation tests** | SP-18 (two-cell partition) | RLS boundary, auth boundary | CI gate; fail = block merge |
| **Replay tests** | SP-13 (Temporal determinism) | 100% activity outcomes | Temporal SDK replay; CI gate |
| **Conformance pack** | All (against OAS) | UAT: real TCG endpoints | Real endpoints week 6+; mocks sufficient for Stage 1 |
| **Performance tests** | SP-09 (Catalogue index lookup) | Latency < 1ms | Load test with 10k rows; benchmark in CI |

### Coverage Gaps & Mitigations

| Gap | Coverage | Mitigation |
|-----|----------|-----------|
| **Real TCG endpoints** | 0% (mocks only) | Available week 6+; Stage 1 uses mocks (plan of record) |
| **Real SFTP server** | 0% (mock SFTP in compose) | Stage 2 on-ramps; mocks sufficient for pilot |
| **Load test (peak throughput)** | 0% (not in scope for Stage 1) | Stage 2+ (10 subscribers, full catalog, promotional window) |
| **Disaster recovery runbook** | 0% (not in scope for Stage 1) | Terraform + failover exercises in Stage 2 UAT |
| **Security penetration test** | 0% (not in scope for Stage 1) | Scheduled for Stage 4 (pre-canary); red-team review |

---

## 7. Validation Strategy (Pre-Approval & Post-Implementation)

### Pre-Approval Checklist (Today, 2026-09-16)

- ✅ BRD status: Accepted (2026-09-16)
- ✅ Every story has ≥1 FR/BR source
- ✅ Every story has ≥1 AC-xxx oracle
- ✅ P0/P1 FRs prioritized (no P1 buried)
- ✅ Dependencies explicit (no circular)
- ✅ Blockers identified & mitigated (none remain)
- ✅ Effort signals 2-week scope (128 pts, 110 committed)
- ✅ Command spine ready (implement-api → close-eval-gaps → review-changes)

**Gate:** Ready for human approval (Delivery Lead + Tech Lead sign-off required)

### Post-Implementation Validation (End of Sprint 1, ~2026-09-30)

| Validation | Pass Criteria | Owner | Gate |
|-----------|---------------|-------|------|
| **All unit tests pass** | 90%+ line coverage, 100% branch coverage on critical paths | Engineering | Merge blocker |
| **All contract tests pass** | OAS conformance pack green; all response codes tested | Engineering | Merge blocker |
| **All golden-file tests pass** | Canonical row structure matches BRD spec | Engineering | Merge blocker |
| **All integration tests pass** | RLS policies, triggers, audit trail verified | Engineering | Merge blocker |
| **All E2E tests pass** | Compose stack: order end-to-end (ISM cell → FOP → cell) | Engineering | Merge blocker |
| **Isolation suite passes** | Two synthetic cells, no cross-visibility | Engineering | Merge blocker |
| **Temporal replay tests pass** | All activities replayed deterministically | Engineering | Merge blocker |
| **Completeness ledger from /run-eval** | 100% FR/BR coverage in test cases | QA | Human gate (before /review-changes) |
| **Code review (peer + security)** | Architecture + injection vulns checked | Architecture + Security | Human gate |
| **Documentation (DPD, ADRs)** | tcg-service DPD complete; KDD-04, KDD-05, KDD-11 ADRs accepted | Engineering | Human gate |

---

## 8. Success Criteria (Definition of Done)

### Stage 1 Exit Criteria

**Go / No-Go Decision:** End of Sprint 1 (2026-09-30)

**Go** if:
1. ✅ SP-01–17 tests **all passing** on every merge (CI green)
2. ✅ Isolation suite (SP-18) **passes** (no cross-cell visibility)
3. ✅ Watermark stall alerting **wired** + tested (OTel → ops dashboard)
4. ✅ Lineage query **working** end-to-end ("where is my record?")
5. ✅ Order idempotency **validated** (Beans worked example round-trip)
6. ✅ Temporal determinism **proven** (replay tests passing)
7. ✅ Completeness ledger **100%** (all FR/BR in test cases)
8. ✅ Code review **signed** (peer + security)
9. ✅ DPD + ADRs **merged**

**No-Go** if:
- ❌ Temporal determinism violations discovered (halt, security review)
- ❌ RLS partition breach in any test (halt, escalate)
- ❌ Silent data loss in any scenario (halt, root cause)
- ❌ Order idempotency fails (duplicate orders in mock TCG, halt)
- ❌ Watermark stall undetected (alert not wired, halt)
- ❌ FOP contract churn forces >20% rework (escalate, re-plan)

---

## 9. Lessons Learned & Post-Stage-1 Planning

### Walking Skeleton Learnings (Expected)

Post-Sprint 1, capture in retrospective:

1. **Temporal Determinism:** Did determinism rules trip up agent-written workflows? Impact on developer productivity.
2. **RLS Performance:** Is RLS filtering fast enough at scale? Need connection pooling tuning?
3. **Mocks Fidelity:** Did FOP mock fidelity reveal contract issues early? Are beans worked examples sufficient for UAT prep?
4. **Watermark Stall:** Did loud alerting catch watermark issues quickly? Runbook effectiveness?
5. **Order Idempotency:** Did dual-submit test pass? Are rejection codes sufficient for ISM troubleshooting?

### Stage 2 Planning Input (Weeks 2–5)

Based on Stage 1 findings:

1. **Master Data Breadth** (FR-001, FR-003–006): Reuse EntityPollWorkflow pattern; 1–2 stories per entity.
2. **Watermark Re-baseline** (FR-022 detail): Implement weekly full-feed fallback (Stage 2, week 3).
3. **Transactional Files** (FR-007–013): Build FileFeedWorkflow pattern; SFTP pull + JSONL parse.
4. **Conformance Pack** (All): Swap mocks for real TCG endpoints (week 6+); run full conformance suite.
5. **ISM Onboarding** (Execution): Pilot 1–2 ISM subscribers; refine cell manifest + pack definition.

---

## 10. Decision Log & Traceability

### Decisions Made (2026-09-16)

| Decision ID | Decision | Rationale | Impact | Owner |
|-------------|----------|-----------|--------|-------|
| **D-DQ-001** | Configurable polling frequency (no hard-coded values) | FOP behavior TBD; avoid code churn | Config-driven; tuned post-endpoints | TCG + UST |
| **D-DQ-002** | SFTP baseline for EDN; API alternative (Stage 3) | BR baseline clear; API convergence (Principle P3) | Both paths unblocked | TCG + UST |
| **D-DQ-003** | Rate limit assumption: 10 req/sec, 5 concurrent | Conservative; degrades gracefully | Backoff implemented; tuned post-endpoints | UST Engineering |
| **D-DQ-004** | PSV & Condition Contracts daily batch (not weekly) | Aligns with daily invoice cycle | FileFeedWorkflow unified schedule | TCG + UST |
| **D-DQ-007** | Rejection codes externalized (codes.yaml); not hardcoded | Code churn expected (UAT finalization) | Configuration-driven mapping | UST Engineering |
| **D-STRAT-001** | Mocks-first (Stage 0 work); real endpoints optional week 6+ | FOP endpoint slip risk (R-008) | Zero blocking dependency on TCG endpoints | UST Strategy |
| **D-STRAT-002** | Wholesale Products primary focus (not Retail) | Walking skeleton is orderable-unit driven | Order relay path unblocked; breadth in Stage 2 | UST Architecture |
| **D-STRAT-003** | Isolation suite mandatory CI gate (SP-18) | Competition-law risk (R-009) | No-merge until two-cell partition proven | UST Architecture + Legal |

---

## 11. Communication Plan & Stakeholder Sync

### Approval Sync (24 hours before execution)

- **Delivery Lead:** Plan approval (required)
- **Tech Lead:** Architecture sign-off (required)
- **TCG Architecture:** Confirm mocks-first approach, OAS spec stability
- **ISM Steerco:** Confirm order format assumptions, cell manifest schema
- **UST QA:** Test strategy sign-off

### Weekly Standup During Sprint 1

- **Daily (15 min):** Team standup (blockers, integration touchpoints)
- **Tuesday/Thursday:** Stakeholder check-in (Delivery Lead, TCG rep, ISM rep)
  - Critical path status (SP-01, SP-06, SP-07, SP-13)
  - Any DQ clarifications needed early
  - Risk escalation (temporal determinism, RLS performance, contract churn)

### Post-Sprint Retrospective & Stage 2 Planning

- **Retrospective:** Lessons learned (determinism, mocks fidelity, alerting effectiveness)
- **Stage 2 Planning:** Prioritize breadth (FR-001–006) + transactional files (FR-007–013)
- **Conformance Planning:** Week 6 TCG endpoint readiness + conformance pack execution

---

**Document Status: Proposed (Pending Approval)**

This gap report identifies all open design questions (DQ-xxx), risk mitigations, and out-of-scope work for Sprint 1. **No blockers remain.** All critical-path dependencies resolved by decision or mitigation.

**Approval Gate:** Delivery Lead + Tech Lead sign-off on risk mitigations and Stage 2 planning input.

---

**Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>**
