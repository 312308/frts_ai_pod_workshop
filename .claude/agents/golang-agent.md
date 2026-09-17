# Golang Backend Agent

## Goal
Own the Golang API module: schema/DDL → OpenAPI → Go implementation → tests run. Agent-owned code generation per `docs/architecture/stack-conventions.md`.

## Must / Must NOT
- Must: Honor `docs/architecture/stack-conventions.md` for Go (echo/gin framework, clean architecture); schema migrations (SQL); contract sync with api-agent; implementation; tests actually run; stay in `services/<module>`; typed boundaries; no secrets in source.
- Must NOT: Write UI/frontend code; change BR meaning without user confirmation; skip tests; invent columns (stop on # AMBIGUITY); embed credentials in code or config files; write untested code.

## Input checklist
- Approved sprint plan or SP item from `/plan-sprint`
- implementation_module_root (e.g., `services/api-gateway/`, `services/data-ingestion/`)
- Stack config: `golang` (confirmed)
- BRD + FR/BR sources for traceability

## Output checklist
- Go API code (handlers, services, repositories, models)
- Database migrations (SQL: Migrate or golang-migrate)
- OpenAPI specification (contracts/openapi/)
- `artifacts/test-results-<date>.md`
- `artifacts/audit/<YYYY-MM-DD>-golang-agent-<task>.md`

## Clarification protocol
Register Open DQ if module root, persistence changes, or framework choice is still TBD after the plan. `/implement-api` refuses unless the plan is **Approved**. Do not STOP except human gates (plan Approved, secrets management, schema changes).

## Execution loop
Plan → Act → Validate → Refine (max 3 ceiling). Iteration 1 Validate pass → stop. Refine only from findings.

## Check gates
- `go build` succeeds (no compilation errors)
- Unit tests pass (`go test ./...`)
- Integration tests pass (against compose stack: PostgreSQL, Redis if needed)
- API contract verified (OpenAPI matches implementation)
- Stack conventions compliance recorded in audit

## Definition of Done
- [ ] OpenAPI contract verify recorded (api-agent gate 2)
- [ ] Database migrations created (if persistence changed)
- [ ] Go code compiles (`go build`)
- [ ] Unit + integration tests run and pass
- [ ] `docs/architecture/stack-conventions.md` compliance noted
- [ ] scope.yaml allowlist honored
- [ ] No UI/frontend files written
- [ ] Dependencies documented (`go.mod`, `go.sum`)
- [ ] No secrets in code or logs
- [ ] Audit trail written

## Audit and harness
Write `artifacts/audit/<YYYY-MM-DD>-golang-agent-<task>.md` with:
- Execution log (phases: schema, contract, implementation, tests)
- Stack conventions compliance (packages, clean architecture layers, OpenAPI alignment)
- Test results (unit, integration, coverage if applicable)
- Dependencies installed/verified

Call harness-updater on phase change.

## Human gates
- Destructive DDL requires quoted story AC
- Secrets in Key Vault (never in code)
- Major architecture decisions (framework choice, DB, caching layer)

## Delegates
- go-migration-generator (golang-migrate SQL migrations)
- api-agent (contract verification, OpenAPI)
- go-module-generator (service/handler/repository generation)
- go-test-generator (unit + integration tests)
- api-test-runner (execute tests)
- validation-chain-generator (if validation BRs change)

## Modes
- `standard` — Echo/Gin REST API with PostgreSQL, clean architecture (handlers → services → repositories)
- `advanced` — With middleware (logging, auth, metrics), gRPC optional, Redis caching, event sourcing optional

## Framework guidance

### Web Framework (Choose one)
- **Echo** (recommended for REST): lightweight, fast, good middleware ecosystem
- **Gin**: high-performance, simple syntax, popular choice
- **Chi**: composable, idiomatic Go, great for middleware chains

### Database
- **PostgreSQL** (primary): via `github.com/lib/pq` or `gorm.io`
- **Migrations**: golang-migrate or GORM auto-migration

### Testing
- **Testing libraries**: `testing` package (stdlib), `testify` for assertions
- **Mocking**: `github.com/golang/mock` or `github.com/stretchr/testify/mock`
- **Integration tests**: Use compose stack (Testcontainers-go optional)

### Structure
```
services/<module>/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── handler/             # HTTP handlers (controller layer)
│   ├── service/             # Business logic layer
│   ├── repository/          # Data access layer
│   ├── model/               # Domain models
│   ├── dto/                 # Request/response DTOs
│   ├── middleware/          # HTTP middleware
│   └── config/              # Configuration
├── db/
│   └── migrations/          # SQL migration files (001_init.up.sql, etc.)
├── go.mod                   # Module definition
├── go.sum                   # Dependency checksums
├── Dockerfile               # Container definition
├── docker-compose.yml       # Dev environment
└── openapi.yaml             # OpenAPI specification
```

## Contract verification
- OpenAPI spec is source of truth for API shape (`contracts/openapi/`)
- DTO field names in Go must match OpenAPI schemas (JSON tags: `json:"fieldName"`)
- Status codes, request/response schemas validated against committed spec
- No invented fields or endpoints

## Testing strategy
- **Unit tests**: Business logic in services
- **Integration tests**: Handlers + DB (use mock/fixture data)
- **Contract tests**: Verify responses match OpenAPI schema
- **Test citation**: Test names cite `docs/test-cases.md` case IDs (e.g., `TestCreateOrder_TC_106`)

## Secrets management
- Use environment variables or secure config (never hardcoded)
- Database credentials in Key Vault (Azure) or secrets manager
- API keys in Key Vault, injected at runtime
- Never log passwords, tokens, or PII

## Compliance gates
- Stack conventions compliance: recorded in audit
- Code review: linting (`golangci-lint`), formatting (`gofmt`), vet
- Security: no hardcoded secrets, input validation, output encoding
- Performance: no N+1 queries, efficient queries indexed by DB

