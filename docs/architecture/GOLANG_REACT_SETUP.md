# Golang + React Setup Guide

This document explains how to use the Golang Backend Agent and React Frontend components in the AI Factory starter kit.

---

## Quick Start

### 1. Create a Golang API Module

Run `/implement-api` with the following inputs:

```
Sprint Plan: artifacts/sprints/<sprint-id>-plan.md (Status: Approved)
Module Root: services/api-gateway/  (or your module name)
Stack: golang
Framework: echo  (or gin)
```

The `golang-agent` will:
1. Create the Go project structure (handlers, services, repositories)
2. Generate database migrations
3. Verify OpenAPI contract
4. Write unit + integration tests
5. Build and test the module

### 2. Create a React Frontend

Run `/implement-ui` with the following inputs:

```
Sprint Plan: artifacts/sprints/<sprint-id>-plan.md (Status: Approved)
Mockup/UX Spec: artifacts/design/*-mockup.png or UX spec
OpenAPI: contracts/openapi.yaml
```

The `frontend-agent` will:
1. Generate React components from mockup/spec
2. Create typed API client from OpenAPI
3. Wire components to backend endpoints
4. Write unit + E2E tests
5. Verify FE↔BE contract

---

## File Structure Overview

After both agents complete, your project will look like:

```
.
├── services/
│   ├── data-ingestion/           # Java Spring service (existing)
│   │   ├── tcg-service-api/
│   │   ├── tcg-service-core/
│   │   ├── tcg-service-persistence/
│   │   └── tcg-service-test/
│   └── api-gateway/              # NEW: Golang service
│       ├── cmd/
│       │   └── server/main.go
│       ├── internal/
│       │   ├── handler/
│       │   ├── service/
│       │   ├── repository/
│       │   ├── model/
│       │   ├── dto/
│       │   └── middleware/
│       ├── db/migrations/
│       ├── test/
│       ├── go.mod
│       ├── openapi.yaml
│       └── docker-compose.yml
├── frontend/                     # NEW: React app
│   ├── src/
│   │   ├── components/
│   │   │   ├── common/
│   │   │   └── features/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── lib/
│   │   │   └── api/
│   │   ├── context/
│   │   └── styles/
│   ├── tests/
│   ├── public/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── Dockerfile
├── contracts/
│   └── openapi.yaml             # Shared API contract
├── docs/
│   ├── architecture/
│   │   ├── stack-conventions.md  (Java + Next.js)
│   │   └── stack-conventions-golang-react.md  (Golang + React)
│   ├── test-cases.md
│   ├── test-matrix.md
│   └── test-cases-br-coverage.md
└── artifacts/
    ├── brd/
    ├── sprints/
    ├── audit/
    └── test-results-*.md
```

---

## Agent Usage

### Golang Agent (`golang-agent`)

**Responsibilities:**
- Golang API implementation (Echo/Gin framework)
- Database schema and migrations
- OpenAPI contract verification
- Unit + integration tests
- Security (no hardcoded secrets)
- Stack conventions compliance

**How to invoke (via `/implement-api`):**

```bash
/implement-api SP-05  # Implement a specific Golang story
```

**Input requirements:**
- Approved sprint plan (`artifacts/sprints/<id>-plan.md`)
- Story SP item (e.g., SP-05: Delta Query Handler)
- implementation_module_root (e.g., `services/api-gateway/`)

**Output artifacts:**
- `services/<module>/` — Complete Go project
- `services/<module>/openapi.yaml` — API contract
- `db/migrations/*.sql` — Database migrations
- `artifacts/test-results-<date>.md` — Test execution log
- `artifacts/audit/<date>-golang-agent-<task>.md` — Audit trail

**Example story mapping:**
- **SP-05 (Delta Query Handler)** → `internal/handler/delta_query.go` + tests
- **SP-06 (Entity Poll Workflow)** → `internal/service/poll_service.go` + tests
- **SP-11 (Cell API Intake)** → `internal/handler/order_intake.go` + tests

### Frontend Agent (React via `frontend-agent`)

**Responsibilities:**
- React component generation
- Typed API client from OpenAPI
- FE↔BE wiring verification
- UI unit + E2E tests
- Component styling (Tailwind CSS)

**How to invoke (via `/implement-ui`):**

```bash
/implement-ui SP-11  # Implement UI for order intake story
```

**Input requirements:**
- Approved sprint plan
- Story SP item with UI spec
- OpenAPI contract (`contracts/openapi.yaml`)
- Mockup or UX spec

**Output artifacts:**
- `frontend/src/` — React project with components
- `frontend/src/lib/api/client.ts` — Typed API client
- `docs/sprint<N>/wiring/` — FE↔BE wiring evidence
- `tests/unit/*.test.tsx` — Component tests
- `tests/e2e/*.spec.ts` — E2E tests with Playwright

**Example story mapping:**
- **SP-11 (Cell API Intake UI)** → `components/features/orders/OrderForm.tsx` + tests
- **SP-13 (Order Format Mapping UI)** → `components/features/orders/OrderMapping.tsx` + tests

---

## Configuration

### golang-agent Config

Set in `.claude/config.yaml` or when invoking:

```yaml
implementation:
  modules:
    backend:
      language: golang
      framework: echo  # or 'gin'
      root: services/api-gateway/
      database: postgresql
    frontend:
      framework: react
      root: frontend/
```

### Stack Conventions

- **Golang**: `docs/architecture/stack-conventions-golang-react.md`
- **React**: Same file (shared OpenAPI contract)

All agents reference these conventions and verify compliance in audits.

---

## OpenAPI Contract (Shared)

The OpenAPI spec is **single source of truth** for both frontend and backend.

**Location:** `contracts/openapi.yaml`

**Usage:**
- Golang: `golang-agent` verifies handlers match OpenAPI schemas
- React: `frontend-agent` generates typed client from OpenAPI

**Workflow:**
1. API contract defined in OpenAPI (paths, schemas, response codes)
2. Golang handler implements per spec (DTOs with JSON tags)
3. React client generated from spec (TypeScript types)
4. Tests verify both sides conform to spec

Example:
```yaml
# contracts/openapi.yaml
paths:
  /api/v1/orders:
    post:
      operationId: createOrder
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                customer_id:
                  type: string
                amount:
                  type: number

# ↓ Golang handler matches this spec
type CreateOrderRequest struct {
  CustomerID string  `json:"customer_id"`  // ← JSON tag must match
  Amount float64     `json:"amount"`
}

// ↓ React client generated with matching types
export interface CreateOrderRequest {
  customer_id: string;
  amount: number;
}
```

---

## Development Workflow

### Local Setup

1. **Start dependencies** (PostgreSQL, Redis if needed):
   ```bash
   cd services/api-gateway/
   docker-compose up -d
   ```

2. **Run Golang service**:
   ```bash
   go run ./cmd/server/main.go
   # Listens on http://localhost:8080
   ```

3. **Run React frontend** (in separate terminal):
   ```bash
   cd frontend/
   npm run dev
   # Listens on http://localhost:5173 (Vite) or http://localhost:3000 (CRA)
   ```

4. **Verify FE↔BE communication**:
   - Open http://localhost:5173 in browser
   - Test API calls (check Network tab)
   - Verify no CORS errors, correct status codes

### Testing

**Golang:**
```bash
cd services/api-gateway/
go test ./...              # Unit tests
go test -race ./...        # Race detection
go test -cover ./...       # Coverage
```

**React:**
```bash
cd frontend/
npm test                   # Unit tests (Vitest)
npm run test:e2e          # E2E tests (Playwright)
npm run test:coverage     # Coverage report
```

### Code Quality

**Golang:**
```bash
golangci-lint run         # Linting
gofmt -w ./...            # Format
go vet ./...              # Vet
```

**React:**
```bash
npm run lint              # ESLint
npm run format            # Prettier
npm run type-check        # TypeScript check
```

---

## Testing Strategy

### Test Organization

**Golang:**
- Unit tests: Business logic in services
- Integration tests: Handlers + DB with fixtures
- Test names cite `docs/test-cases.md` IDs (TC-106, etc.)

**React:**
- Unit tests: Component rendering, user interactions
- E2E tests: Full user journeys (Playwright)
- Test names cite `docs/test-cases.md` IDs

### BR Coverage Validation

After `/implement-api` and `/implement-ui`:

```bash
# Check BR coverage
docs/test-cases-br-coverage.md  # Status should be PASS (90%+ coverage)
```

If FAIL, `/expand-test-coverage` adds more cases before `/run-tests`.

---

## Secrets Management

### Never in Code/Logs
- Database passwords
- API keys, OAuth tokens
- Encryption keys
- Private credentials

### Correct Approach

**Golang:**
```go
// Load from environment variables
dbPassword := os.Getenv("DB_PASSWORD")  // Never hardcode!

// Or use Key Vault (Azure)
import "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
// Inject at runtime
```

**React:**
```typescript
// Environment variables (Vite prefix: VITE_)
const apiUrl = import.meta.env.VITE_API_URL  // ✓ OK
const apiKey = import.meta.env.VITE_API_KEY  // ✓ OK (but never log)

// Don't hardcode:
// const apiKey = "secret-key-12345"  // ✗ NEVER
```

**Configuration files:**
```
.env.example     # Template (no secrets, in git)
.env.local       # Local only (in .gitignore, NOT in git)
.env.production  # Production secrets via CI/CD, Key Vault, not in git
```

---

## Deployment

### Golang Service

**Docker:**
```dockerfile
# services/api-gateway/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

**Kubernetes/Cloud (outline):**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api-gateway
        image: registry.example.com/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: host
```

### React Frontend

**Docker:**
```dockerfile
# frontend/Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
RUN npm install -g serve
COPY --from=builder /app/dist ./dist
EXPOSE 3000
CMD ["serve", "-s", "dist", "-l", "3000"]
```

---

## Common Tasks

### Add a New Golang Endpoint

1. Add path to `contracts/openapi.yaml`
2. Create handler in `internal/handler/`
3. Implement service in `internal/service/`
4. Add repository method in `internal/repository/`
5. Write tests (cite TC IDs)
6. Run `go test ./...`
7. Commit + push

### Add a New React Component

1. Create component in `src/components/features/`
2. Use typed API client from `lib/api/`
3. Add test in `tests/unit/`
4. Run `npm test`
5. Run E2E test if journey changes
6. Commit + push

### Update OpenAPI Contract

1. Edit `contracts/openapi.yaml`
2. Update Golang DTOs (JSON tags)
3. Regenerate React client types
4. Update tests if response schema changed
5. Verify `/implement-api` and `/implement-ui` against new spec

---

## Troubleshooting

### Golang Build Fails
```bash
go mod tidy               # Update dependencies
go mod verify             # Check checksums
go mod download           # Re-download
go build ./...            # Retry build
```

### React Build Fails
```bash
npm ci                    # Clean install
npm audit fix             # Fix vulnerabilities
npm run build             # Retry build
```

### CORS Errors (FE ↔ BE)
- Check `frontend/.env` has correct `VITE_API_URL`
- Check Golang service has CORS middleware enabled
- Verify API endpoint is accessible: `curl http://localhost:8080/api/v1/health`

### Tests Fail
- Check test data fixtures in `test/fixtures/`
- Verify database is running (`docker-compose ps`)
- Check test database is clean (migrations applied)
- Run with verbose: `go test -v ./...`

---

## References

- **Golang Agent Definition**: [`.claude/agents/golang-agent.md`](./.claude/agents/golang-agent.md)
- **Stack Conventions (Golang + React)**: [`docs/architecture/stack-conventions-golang-react.md`](./stack-conventions-golang-react.md)
- **OpenAPI Specification**: [`contracts/openapi.yaml`](../contracts/openapi.yaml)
- **Test Cases**: [`docs/test-cases.md`](../test-cases.md)

