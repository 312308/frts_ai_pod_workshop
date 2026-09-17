# Stack conventions (Java + Next.js)

This starter kit uses **agent-owned** code generation. There is no J2 template render pipeline.

## Spring Boot (`modules/api`)

```
controller/ → REST endpoints, validation, ResponseEntity
service/ → business logic, transactions
repository/ → JPA repositories
dto/ → request/response DTOs
entity/ → JPA entities
mapper/ → MapStruct or manual mappers
db/migration/ → Flyway versioned SQL
```

- OpenAPI is source of truth for API shape (`contracts/openapi/`).
- DTO field names must match OpenAPI schemas.
- No secrets in code or logs.

## Next.js (`modules/ui`)

```
app/ → App Router pages
components/ → presentational + container components
lib/api/ → typed client from OpenAPI
```

- Wiring gate: every UI action traces to OpenAPI operation + story AC.
- Use typed fetch client; no invented API fields.

## Tests

- JUnit 5 for API; cite `docs/test-cases.md` case IDs in test names.
- Vitest for UI units; Playwright for E2E journeys.
- br-coverage-validator PASS before test scripts.

## Audit note

Record stack-conventions compliance (layering, OpenAPI alignment, cited case IDs) in the implement audit.
