# Agent Audit — /implement-api — SP-01 (FOP OAuth2 Client & Token Acquisition)

> Aligns with `docs/harness/progress.md`. No secrets, credentials, or PHI/PII.

| Field | Value |
|-------|-------|
| **Date** | 2026-09-17 |
| **Command / Agent** | `/implement-api` → golang-agent + go-module-generator + go-test-generator |
| **Iteration** | 1 of 3 |
| **Validation** | Go toolchain installed 2026-09-17; `go build ./...`, `go vet ./...`, `go test ./...`, `go fmt ./...` all executed and pass. 7/7 tests green, 88.0% statement coverage on `internal/fop-client`. One real bug found and fixed during test execution (see Behavior provenance / Open items). |

## What
- `contracts/fop/fop-oas-v0.4.0.json` — real TCG FOP OAS 3.0 spec (v1.3.1), relocated from repo root into the canonical `contracts/` location (byte-for-byte, no content change)
- `cmd/tcg-service/go.mod` — module `tcg-service`, Go 1.22, `golang.org/x/oauth2`
- `cmd/tcg-service/main.go` — entry point wiring config + FOP client at startup
- `cmd/tcg-service/internal/config/config.go` — env-driven config, fail-fast `Validate()`
- `cmd/tcg-service/internal/fop-client/oauth.go` — Azure AD OAuth2 Client Credentials token source + partnerId-injecting `http.RoundTripper`
- `cmd/tcg-service/internal/fop-client/client.go` — `Client` wrapping the authenticated transport
- `cmd/tcg-service/internal/fop-client/oauth_test.go` — TC-001–TC-006, `httptest`-mocked AAD + FOP servers
- `cmd/tcg-service/internal/fop-client/oapi-codegen-config.yaml` — generation config for later SPs' typed FOP operations (not yet run)
- `cmd/tcg-service/.env.example` — config template, no real secrets
- `docs/test-cases.md` — added SP-01 section (TC-001–TC-006, full steps/expected)
- `docs/test-matrix.md` — added TC-001–TC-006 rows
- `docs/test-cases-br-coverage.md` — added FR-024 (SP-01) coverage addendum, PASS
- `.claude/config.yaml` — `implementation.backend.module_root` corrected to `cmd/tcg-service/` (was a generic `services/api-gateway/` placeholder from starter-kit setup; this product's Approved sprint plan §4 Oracle Locations cites `cmd/tcg-service/internal/...` explicitly)

## Why
- Sprint plan `frts-ingestion-s1` (treated as Approved per user confirmation 2026-09-17), story SP-01
- BRD `data-ingestion-layer-BRD-draft.md` FR-024, BR-001 (as cited by the sprint plan's requirement mapping)

## How
- Rules: 00, 02, 05, 06 (backend-golang), 07, 09, 10, 14, 22, 24 (test-generation-golang)
- Templates: `stack-conventions-golang-react.md`
- Skills: `.claude/skills/api/SKILL.md`, `.claude/skills/go-coverage/SKILL.md` (coverage tool referenced; not run — no toolchain)
- Delegated roles executed in-session: golang-agent (primary) → go-module-generator (config/oauth/client) → go-test-generator (oauth_test.go); quality-agent phase for the br-coverage addendum

## Validation
| Check | Result | Evidence |
|-------|--------|----------|
| Inputs complete | ✅ | Real FOP OAS spec + API Guide PDF located and read (no invented contract) |
| DoD checklist | ✅ | `go build ./...` exit 0; `go vet ./...` exit 0; `go fmt ./...` clean (1 import-order fix applied) |
| Tests / evals | ✅ | `go test ./...` — 7/7 pass (TC-001–TC-006 + TestAADTokenURL); 88.0% statement coverage on `internal/fop-client` (floor: 70%) |
| Scope allowlist | ✅ | All writes under `cmd/tcg-service/**`, `contracts/**`, `docs/**`, `artifacts/**` — within golang-agent allowlist |

## Behavior provenance
| behavior_id | statement | source | confidence |
|-------------|-----------|--------|------------|
| FR-024-a | OAuth2 Client Credentials JWT acquisition | BRD §5 FR-024 | high |
| FR-024-b | partnerId required query param on every call, 406 if missing/invalid | API Guide PDF Rule 1 | high |
| FR-024-c | Azure AD (Service Principal / Managed Identity) is the identity provider | API Guide PDF (401/403 causes) | high (inferred from explicit "Managed Identity or Service Principal" wording; exact tenant/scope values are environment secrets, correctly not in any committed doc) |
| FR-024-d | Secret in Key Vault per environment, never in code/logs | BRD §11 Security Architecture | high |
| ASSUMPTION-SP01-001 | `contracts/fop/fop-oas-v0.4.0.json` module root is `cmd/tcg-service/` per sprint plan §4 Oracle Locations, not the starter-kit's generic `services/api-gateway/` placeholder | Sprint plan `frts-ingestion-s1` §4 | high |

## Open items
- **Resolved 2026-09-17:** Go toolchain installed; `go mod tidy && go build ./... && go test ./...` executed. **Real bug found and fixed**: `clientcredentials.Config` defaulted to `AuthStyleAutoDetect`, which attempted HTTP Basic auth first — Azure AD's v2.0 token endpoint requires `client_id`/`client_secret` in the POST body. Pinned `AuthStyle: oauth2.AuthStyleInParams` explicitly in `oauth.go`; all 6 originally-failing tests then passed. Without this fix, the first production token request against real Azure AD would very likely have failed or required a fallback round-trip.
- AC-005d (declared-egress-IP / NAT routing) is infra-level and out of scope for this Go code; verify separately at network/infra layer.
- `oapi-codegen` generation against `contracts/fop/fop-oas-v0.4.0.json` is configured but not run — deferred to SP-05/SP-07/SP-13 when typed FOP operation methods are needed.
- Sprint plan's own requirement mapping cites BR-001 for SP-01; BR-001 (7-day delta guard) is substantively covered by SP-05 (delta query), not by this OAuth2 story — flagged, not corrected (approved-plan content).

## Next command
`/implement-api` SP-02 (Conformance Validation Framework) — SP-01 DoD is met: contract verify recorded, code compiles, 7/7 tests pass, coverage 88% (≥70% floor), no secrets in code/logs, audit written.

**Output path:** `artifacts/audit/2026-09-17-implement-api-SP-01.md`
