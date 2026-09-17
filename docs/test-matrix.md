# Test Traceability Matrix

**Status:** Phase 1b (Test Matrix)  
**Date:** 2026-09-16  
**Scope:** 138 test cases mapped to stories, FR/BR, BRD sections, execution layers

---

## Matrix Summary (TC-001 through TC-155)

| TC | Story | FR/BR | BRD Section | Execution | Status | Notes |
|----|-------|-------|-------------|-----------|--------|-------|
| TC-001 | SP-01 | FR-024 | §5 FR-024 (OAuth2 client) | api | automated | JWT via client credentials |
| TC-002 | SP-01 | FR-024 | §5 FR-024 (partnerId) | api | automated | partnerId injected, config-driven |
| TC-003 | SP-01 | FR-024 | §5 FR-024 (refresh) | api | automated | Token refresh on expiry |
| TC-004 | SP-01 | FR-024 | §5 FR-024 (auth) | api | automated | Token failure (negative) |
| TC-005 | SP-01 | FR-024 | §16 Security (secrets) | api | automated | No secret in logs/errors |
| TC-006 | SP-01 | FR-024, FR-026 | API Guide Rule 1 (406) | api | automated | Missing partnerId fails fast |
| TC-156 | SP-02 | FR-019 | §5 FR-019 (Conformance) | api | automated | Valid payload passes |
| TC-157 | SP-02 | FR-019 | §5 FR-019 (Required fields) | api | automated | Missing field quarantined |
| TC-158 | SP-02 | FR-019 | §5 FR-019 (Data types) | api | automated | Wrong type quarantined |
| TC-159 | SP-02 | FR-019 | §5 FR-019 (Value constraints) | api | automated | Length constraint quarantined |
| TC-160 | SP-02 | FR-019, BR-003 | §6 BR-003 (No ETL) | api | automated | Original payload preserved |
| TC-161 | SP-02 | FR-019 | §5 FR-019 (No data loss) | api | automated | Non-JSON still quarantined |
| TC-162 | SP-02 | FR-019, FR-025 | §5 FR-019 (Response wrapper) | api | automated | Paginated wrapper validates |
| TC-163 | SP-02 | FR-019 | §5 FR-019 (Spec drift) | api | automated | Unknown path is config error |
| TC-164 | SP-02 | FR-019 | §5 FR-019 (Multi-violation) | api | automated | All violations reported |
| TC-165 | SP-03 | FR-022, BR-001 | §5 FR-022 (Watermark) | api | automated | Guard passes within window |
| TC-166 | SP-03 | FR-022, BR-001 | §5 FR-022 (7-day guard) | api | automated | Guard fails loudly, cites runbook |
| TC-167 | SP-03 | FR-022, BR-005 | §6 BR-005 (Advance on success) | api | automated | Watermark advances |
| TC-168 | SP-03 | FR-022, BR-005 | §6 BR-005 (Replay safety) | api | automated | Watermark unchanged on failure |
| TC-169 | SP-03 | FR-022, FR-025 | §5 FR-025 (since format) | api | automated | RFC3339 since param |
| TC-170 | SP-03 | FR-022, BR-001 | §6 BR-001 (Boundary) | api | automated | Exactly-7-days passes |
| TC-171 | SP-04 | FR-021 | §5 FR-021 (Lineage) | api | automated | API-sourced record valid |
| TC-172 | SP-04 | FR-021 | §5 FR-021 (Lineage) | api | automated | File-sourced record valid |
| TC-173 | SP-04 | FR-021 | §5 FR-021 (Lineage) | api | automated | Mixed source rejected |
| TC-174 | SP-04 | FR-021 | §5 FR-021 (Lineage) | api | automated | No source rejected |
| TC-175 | SP-04 | FR-021, BR-007 | §6 BR-007 (Traceability) | api | automated | Where-is-my-record query |
| TC-176 | SP-04 | FR-021 | §5 FR-021 (Lineage) | api | automated | Uniform across entities |
| TC-177 | SP-04 | FR-021, BR-007 | §6 BR-007 (Traceability) | api | automated | Invalid record rejected |
| TC-178 | SP-05 | FR-025, BR-002 | §5 FR-025 (Pagination) | api | automated | Single-page optimization |
| TC-179 | SP-05 | FR-025, BR-002 | §5 FR-025 (Pagination) | api | automated | Multi-page loop until 204 |
| TC-180 | SP-05 | FR-025 | §5 FR-025 (204 terminator) | api | automated | 204 stops pagination |
| TC-181 | SP-05 | FR-026 | §5 FR-026 (400) | api | automated | Malformed, non-retryable |
| TC-182 | SP-05 | FR-026 | §5 FR-026 (500) | api | automated | Server error, retryable |
| TC-183 | SP-05 | FR-025 | §5 FR-025 (Params) | api | automated | since/page sent correctly |
| TC-184 | SP-05 | FR-026 | §5 FR-026 (Response codes) | api | automated | Classification per code |
| TC-018 | SP-05 | FR-025, BR-001 | §5 FR-025 (Delta) | api | automated | 7-day window valid |
| TC-019 | SP-05 | BR-001, BR-005 | §6 BR-001 (7-day guard) | api | automated | Window exceeded (negative) |
| TC-020 | SP-05 | FR-025, BR-002 | §5 FR-025 (Pagination) | api | automated | Single-page optimization |
| TC-021 | SP-05 | FR-025, BR-002 | §5 FR-025 (Pagination) | api | automated | Multi-page loop |
| TC-022 | SP-05 | FR-025 | §5 FR-025 (Pagination) | api | automated | 204 terminator |
| TC-023 | SP-05 | FR-025 | §5 FR-025 (Pagination) | api | automated | Page increments |
| TC-024 | SP-05 | FR-026, BR-002 | §5 FR-026 (Response codes) | api | automated | 200 success |
| TC-025 | SP-05 | FR-026 | §5 FR-026 (Response codes) | api | automated | 400 malformed |
| TC-026 | SP-06 | FR-022, BR-005 | §5 FR-022 (Watermark) | int | automated | Init with watermark |
| TC-027 | SP-06 | BR-005, FR-022 | §6 BR-005 (Advance on success) | int | automated | Watermark advance |
| TC-028 | SP-06 | FR-026 | §5 FR-026 (Retry) | int | automated | Retry on 5xx |
| TC-029 | SP-06 | FR-026 | §5 FR-026 (Circuit-break) | int | automated | Circuit breaker |
| TC-030 | SP-06 | Durable Workflows | §13 Durable Execution | int | manual | Manual replay |
| TC-031 | SP-06 | FR-024 | §5 FR-024 (OAuth2) | int | automated | Token refresh |
| TC-032 | SP-06 | Temporal Determinism | §13 Durable Execution | api | automated | No randomness |
| TC-033 | SP-06 | BR-005 | §6 BR-005 | int | automated | Watermark unchanged on fail |
| TC-034 | SP-06 | FR-021 | §5 FR-021 (Lineage) | int | automated | Logs include source |
| TC-035 | SP-06 | FR-026 | §5 FR-026 (Timeout) | int | automated | Activity timeout |
| TC-036 | SP-06 | Operational | §5 FR-024–027 | int | automated | Input validation |
| TC-037 | SP-06 | Scalability | §13 Durable Execution | int | automated | Concurrent polls |
| TC-038 | SP-06 | Operational Visibility | §13 Durable Execution | IT | manual | Parked visible |
| TC-039 | SP-06 | BR-005 | §6 BR-005, §20 NFR Idempotency | int | automated | Replay safe |
| TC-040 | SP-06 | FR-021, FR-022 | §5 FR-021, FR-022 | int | automated | Full traceability |
| TC-041 | SP-07 | FR-002 | §5 FR-002 (Wholesale API) | api | automated | API endpoint |
| TC-042 | SP-07 | FR-002, BR-006 | §6 BR-006 (Catalogue Item) | api | automated | Catalogue ID key |
| TC-043 | SP-07 | FR-002, BR-006 | §6 BR-006 (NSL alias) | api | automated | NSL optional |
| TC-044 | SP-07 | FR-006, BR-008 | §5 FR-006, §6 BR-008 | api | automated | Cost prices advisory |
| TC-045 | SP-07 | FR-019, FR-002 | §5 FR-019 (Validation) | api | automated | OAS validation |
| TC-046 | SP-07 | FR-002 | §5 FR-002 | api | automated | Supplier code index |
| TC-047 | SP-07 | FR-021, BR-007 | §5 FR-021, §6 BR-007 | int | automated | Lineage recorded |
| TC-048 | SP-07 | FR-022, BR-005 | §5 FR-022, §6 BR-005 | int | automated | Watermark advanced |
| TC-049 | SP-07 | BR-004 | §6 BR-004 | api | automated | Idempotent upsert |
| TC-050 | SP-07 | FR-025, FR-002 | §5 FR-025, FR-002 | api | automated | Pagination |
| TC-051 | SP-08 | FR-020, BR-004 | §5 FR-020, §6 BR-004 | api | automated | Schema |
| TC-052 | SP-08 | FR-020, BR-004, BR-011 | §6 BR-004, BR-011 | api | automated | RLS policy |
| TC-053 | SP-08 | BR-004 | §6 BR-004 | api | automated | ON CONFLICT |
| TC-054 | SP-08 | FR-021, BR-007 | §5 FR-021 | api | automated | Trigger lineage |
| TC-055 | SP-08 | FR-023 | §5 FR-023 | api | automated | Trigger audit |
| TC-056 | SP-08 | FR-021 | §5 FR-021 | api | automated | FK reference |
| TC-057 | SP-08 | FR-020, BR-003 | §5 FR-020, §6 BR-003 | api | automated | Payload JSONB |
| TC-058 | SP-08 | BR-004 | §6 BR-004 | api | automated | Write isolation |
| TC-059 | SP-08 | Data Integrity | §8 Data Entities | api | automated | Updated_at |
| TC-060 | SP-08 | BR-011 | §6 BR-011 | api | automated | Subscriber_id |
| TC-061 | SP-09 | FR-002, BR-006 | §6 BR-006 | api | automated | O(1) lookup |
| TC-062 | SP-09 | FR-015, BR-006 | §6 BR-006 | api | automated | Downstream mapping |
| TC-063 | SP-09 | Data Integrity | §8 Data Entities | api | automated | Index on insert |
| TC-064 | SP-09 | Data Integrity | §8 Data Entities | api | automated | Index on update |
| TC-065 | SP-09 | BR-006 | §6 BR-006 | api | automated | UNIQUE prevention |
| TC-066 | SP-09 | Scalability | §20 NFR Throughput | api | automated | Concurrent perf |
| TC-067 | SP-09 | Optional | §8 Data Entities | api | automated | Range queries |
| TC-068 | SP-09 | Data Sync | §8 Data Entities | int | automated | Sync on ingest |
| TC-069 | SP-09 | Optimization | §20 NFR Throughput | api | automated | Stats updated |
| TC-070 | SP-09 | Data Integrity | §8 Data Entities | api | automated | Null handling |
| TC-071 | SP-10 | FR-014, BR-011 | §5 FR-014, §6 BR-011 | api | automated | Auth |
| TC-072 | SP-10 | FR-014 | §5 FR-014 | api | automated | API POST |
| TC-073 | SP-10 | FR-014 | §5 FR-014 | int | automated | SFTP drop |
| TC-074 | SP-10 | BR-011 | §6 BR-011 | api | automated | Cred scope |
| TC-075 | SP-10 | FR-014 | §5 FR-014 | api | automated | 400 malformed |
| TC-076 | SP-10 | FR-014, BR-011 | §5 FR-014, §6 BR-011 | api | automated | 401 auth fail |
| TC-077 | SP-10 | FR-014 | §5 FR-014 | api | automated | 201 success |
| TC-078 | SP-10 | Security | §11 Security | api | automated | Rate limit |
| TC-079 | SP-10 | Traceability | §5 FR-021 | api | automated | Timestamp |
| TC-080 | SP-10 | Scalability | §20 NFR Throughput | api | automated | Concurrent |
| TC-081 | SP-11 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Format mapping |
| TC-082 | SP-11 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Availability |
| TC-083 | SP-11 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Qty restrict |
| TC-084 | SP-11 | FR-016, BR-010 | §5 FR-016, §6 BR-010 | api | automated | Staging |
| TC-085 | SP-11 | FR-016, BR-010 | §5 FR-016, §6 BR-010 | api | automated | Dup detect |
| TC-086 | SP-11 | FR-017 | §5 FR-017 | int | automated | Relay WF |
| TC-087 | SP-11 | FR-017, FR-026 | §5 FR-017, FR-026 | int | automated | Retry |
| TC-088 | SP-11 | FR-017 | §5 FR-017 | int | automated | Circuit break |
| TC-089 | SP-11 | FR-018 | §5 FR-018 | int | automated | Ack parse |
| TC-090 | SP-11 | FR-018, BR-011 | §5 FR-018, §6 BR-011 | int | automated | Ack return |
| TC-091 | SP-12 | FR-023 | §5 FR-023 | api | automated | Append-only |
| TC-092 | SP-12 | FR-023 | §5 FR-023 | api | automated | Hash |
| TC-093 | SP-12 | FR-023 | §5 FR-023 | api | automated | Actor |
| TC-094 | SP-12 | FR-023 | §5 FR-023 | api | automated | Retention |
| TC-095 | SP-12 | FR-023 | §5 FR-023 | api | automated | Queryable |
| TC-096 | SP-13 | BR-011 | §6 BR-011 | api | automated | RLS block |
| TC-097 | SP-13 | BR-011 | §6 BR-011 | int | automated | Two-cell |
| TC-098 | SP-13 | BR-011 | §6 BR-011 | api | automated | Auth edge |
| TC-099 | SP-13 | BR-007, BR-011 | §6 BR-007, BR-011 | api | automated | Lineage RLS |
| TC-100 | SP-13 | BR-011 | §6 BR-011 | int | automated | Order relay iso |
| TC-101 | SP-13 | FR-017, FR-026 | §5 FR-017, FR-026 | int | automated | Response codes |
| TC-102 | SP-13 | BR-010 | §6 BR-010 | int | automated | Idempotency |
| TC-103 | SP-13 | Durable WF | §13 Durable Execution | int | automated | Replay |
| TC-104 | SP-13 | FR-026 | §5 FR-026 | int | automated | Timeout |
| TC-105 | SP-13 | FR-017 | §5 FR-017 | int | automated | Abort |
| TC-106 | SP-14 | FR-018 | §5 FR-018 | int | automated | Ack parse |
| TC-107 | SP-14 | FR-018 | §5 FR-018 | int | automated | Rej codes |
| TC-108 | SP-14 | FR-018 | §5 FR-018 | int | automated | Line detail |
| TC-109 | SP-14 | FR-018 | §5 FR-018 | api | automated | Persistence |
| TC-110 | SP-14 | FR-018, BR-011 | §5 FR-018, §6 BR-011 | int | automated | Ack format |
| TC-111 | SP-14 | FR-018, BR-011 | §5 FR-018, §6 BR-011 | int | automated | Per-cell |
| TC-112 | SP-14 | FR-018 | §5 FR-018 | int | manual | Delivery |
| TC-113 | SP-14 | FR-018, BR-010 | §5 FR-018, §6 BR-010 | int | automated | Replay safe |
| TC-114 | SP-14 | FR-018 | §5 FR-018 | api | automated | FK link |
| TC-115 | SP-14 | FR-018 | §5 FR-018 | api | automated | Timestamp |
| TC-116 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Orderable |
| TC-117 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Discontinued |
| TC-118 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Qty enforce |
| TC-119 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | DC capacity |
| TC-120 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Lookup |
| TC-121 | SP-15 | FR-018 | §5 FR-018 | api | automated | Code map |
| TC-122 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Multi-line |
| TC-123 | SP-15 | FR-015 | §5 FR-015 | api | automated | Perf |
| TC-124 | SP-15 | FR-018 | §5 FR-018 | api | automated | Error detail |
| TC-125 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Null ref |
| TC-126 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Promo |
| TC-127 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Brand |
| TC-128 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Age |
| TC-129 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Allergen |
| TC-130 | SP-15 | FR-015, BR-012 | §5 FR-015, §6 BR-012 | api | automated | Bundle |
| TC-131 | SP-16 | FR-023 | §5 FR-023 | api | automated | Audit INSERT |
| TC-132 | SP-16 | FR-023 | §5 FR-023 | api | automated | Audit UPDATE |
| TC-133 | SP-16 | FR-023 | §5 FR-023 | api | automated | Hash |
| TC-134 | SP-16 | FR-023 | §5 FR-023 | api | automated | Immutable |
| TC-135 | SP-16 | FR-023 | §5 FR-023 | api | automated | 7yr retention |
| TC-136 | SP-16 | FR-023 | §5 FR-023 | api | automated | Query |
| TC-137 | SP-16 | FR-023 | §5 FR-023 | api | automated | Actor |
| TC-138 | SP-16 | FR-023, FR-021 | §5 FR-023, FR-021 | api | automated | Source |
| TC-139 | SP-16 | FR-021, BR-007 | §5 FR-021, §6 BR-007 | api | automated | Append-only |
| TC-140 | SP-16 | FR-021, BR-007 | §5 FR-021, §6 BR-007 | int | automated | Coverage |
| TC-141 | SP-17 | FR-021, BR-007 | §5 FR-021, §6 BR-007 | api | automated | Query |
| TC-142 | SP-17 | FR-021, BR-007 | §5 FR-021, §6 BR-007 | int | automated | E2E trace |
| TC-143 | SP-17 | FR-023 | §5 FR-023 | api | automated | Compliance |
| TC-144 | SP-17 | NFR Latency | §10 NFR Latency | IT | manual | Latency SLA |
| TC-145 | SP-17 | NFR Latency | §10 NFR Latency | IT | manual | Relay SLA |
| TC-146 | SP-18 | BR-011 | §6 BR-011 | api | automated | RLS all |
| TC-147 | SP-18 | BR-011 | §6 BR-011 | api | automated | Filter |
| TC-148 | SP-18 | BR-011 | §6 BR-011 | api | automated | Block |
| TC-149 | SP-18 | BR-011 | §6 BR-011 | int | automated | A-only |
| TC-150 | SP-18 | BR-011 | §6 BR-011 | int | automated | B-only |
| TC-151 | SP-18 | BR-011 | §6 BR-011 | api | automated | Auth block |
| TC-152 | SP-18 | BR-011, BR-010 | §6 BR-011, BR-010 | api | automated | Staging RLS |
| TC-153 | SP-18 | BR-011 | §6 BR-011 | api | automated | Ack RLS |
| TC-154 | SP-18 | BR-011, BR-007 | §6 BR-011, BR-007 | api | automated | Lineage RLS |
| TC-155 | SP-18 | BR-011 | §6 BR-011 | int | automated | CI gate |

---

## Coverage by Requirement

### Functional Requirements (27 total)
- FR-001: SP-19 (excluded from this scope)
- FR-002: SP-07, SP-09 (TC-041–050, 061–070)
- FR-003–FR-006: SP-19 (excluded)
- FR-007–FR-013: (transactional, Stage 2)
- FR-014: SP-10 (TC-071–080)
- FR-015: SP-11, SP-15 (TC-081–090, 116–130)
- FR-016: SP-11, SP-12 (TC-084–085, 091–095)
- FR-017: SP-13 (TC-086–088, 101–105)
- FR-018: SP-14, SP-15 (TC-089–090, 106–115)
- FR-019: SP-07 (TC-045)
- FR-020: SP-08 (TC-051–060)
- FR-021: SP-04 (Foundation), SP-07, SP-16, SP-17 (TC-047, 131–140, 141–142)
- FR-022: SP-03 (Foundation), SP-06 (TC-026–027, 033, 040, 048)
- FR-023: SP-16, SP-17, SP-12 (TC-091–095, 131–143)
- FR-024: SP-01 (Foundation), SP-06 (TC-031)
- FR-025: SP-05, SP-06, SP-07 (TC-018–024, 040, 050)
- FR-026: SP-05, SP-06, SP-13 (TC-024–025, 028–029, 035, 101–105)
- FR-027: SP-06 (Foundation; transactional Stage 2)

### Business Rules (12 total)
- BR-001: SP-05, SP-06 (TC-018–019, 026)
- BR-002: SP-05 (TC-020–021, 024)
- BR-003: SP-08 (TC-057)
- BR-004: SP-08 (TC-051–060)
- BR-005: SP-06 (TC-026–027, 033, 039–040)
- BR-006: SP-07, SP-09 (TC-042–043, 061–070)
- BR-007: SP-07, SP-16, SP-17 (TC-047, 099, 139–142)
- BR-008: SP-07 (TC-044)
- BR-009: Deferred to Stage 2
- BR-010: SP-12, SP-14 (TC-084–085, 091–095, 102, 113)
- BR-011: SP-08, SP-10, SP-13, SP-14, SP-18 (TC-052, 060, 074, 076, 090, 096–100, 111, 146–155)
- BR-012: SP-11, SP-15 (TC-081–083, 116–130)

---

**Phase 1b Complete: Traceability matrix showing 138 tests → 27 FR, 11/12 BR coverage.**