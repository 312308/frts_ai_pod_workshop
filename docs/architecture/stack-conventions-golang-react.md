# Stack conventions (Golang + React)

This starter kit uses **agent-owned** code generation for the Golang stack. There is no J2/template render pipeline — `golang-agent` and its subagents (`go-module-generator`, `go-test-generator`, `go-migration-generator`) write Go source directly.

Selected via `.claude/config.yaml` → `implementation.backend.language: golang`.

## Golang backend (`services/<module>`)

```
services/<module>/
├── cmd/server/main.go        # Application entry point
├── internal/
│   ├── handler/               # HTTP handlers — REST endpoints, request parsing, status codes
│   ├── service/                # Business logic, orchestration
│   ├── repository/             # Data access (SQL/GORM)
│   ├── model/                  # Domain models
│   ├── dto/                    # Request/response DTOs (JSON tags match OpenAPI)
│   └── middleware/              # Auth, logging, recovery
├── db/migrations/               # golang-migrate SQL files (NNN_name.up.sql / .down.sql)
├── go.mod / go.sum
└── openapi.yaml
```

- Layering: handler → service → repository. No business logic in handlers; no HTTP concerns in services.
- OpenAPI is source of truth for API shape (`contracts/openapi/`).
- DTO field names and JSON tags must match OpenAPI schemas exactly — no invented fields.
- Migrations are additive-only via `golang-migrate` (see `go-migration-generator`); never edit an applied version.
- No secrets in code or logs; load from environment variables or Key Vault at runtime.

## React (`frontend`)

- Unchanged from the kit's existing frontend conventions (typed API client from OpenAPI, no invented microcopy, FE↔BE wiring gate on `/implement-ui`).

## Tests

- Go: stdlib `testing` + `testify` assertions. Test names cite `docs/test-cases.md` case IDs (e.g. `TestCreateOrder_TC_106`). Unit tests for service/repository layers; integration tests for handlers against a real Postgres via the compose stack.
- Coverage: `go test ./... -coverprofile=coverage.out`; parsed by `go-coverage-analyzer` (skill: `.claude/skills/go-coverage/SKILL.md`). Same floor as Java: `config.yaml` → `quality_gates.api_unit_line_coverage_min`.
- Vitest for UI units; Playwright for E2E journeys (unchanged).
- br-coverage-validator PASS required before test scripts, same as the Java path.

## Audit note

Record stack-conventions compliance (layering, OpenAPI alignment, cited case IDs, no invented fields) in the implement audit, same as the Java path — see `artifacts/audit/<date>-golang-agent-<task>.md`.
