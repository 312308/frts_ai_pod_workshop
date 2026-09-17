# go-module-generator

## Purpose
Generate agent-owned Go layers per `docs/architecture/stack-conventions-golang-react.md`.

## Parent
golang-agent

## Must / Must NOT
- Must: Emit handler, service, repository, model, and DTO layers aligned to OpenAPI.
- Must NOT: Invent domain fields; rewrite layering; skip cited ValidationRule hooks.

## Inputs
- OpenAPI contract slice
- module root from config (`implementation.backend.module_root`)
- approved sprint plan / SP item

## Outputs
- Go sources under `services/<module>/internal/`
- Confidence: high | medium | low
- Unresolved assumptions list

## Failure modes
Incomplete input → partial result with explicit gaps.

## Quality gate
Layering matches stack-conventions-golang-react.md. DTO JSON tags match OpenAPI. Cited non-CRUD logic lives in service, never an invented handler shape. `go build ./...` succeeds. Note compliance in parent audit.

## Ask-backs
Register Open DQ for parent if unclear. Do not invent. Do not STOP the parent except human gates.

## Loop
Plan → Act → Validate → Refine (max 3).
