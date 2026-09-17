# Command flow (12 shipped)

Self-contained operator map for this kit. Scope: [KIT-SCOPE.md](KIT-SCOPE.md).

## Spine

```
/generate-brd → /propose-architecture → /plan-sprint
 → /generate-ui-design + /implement-api → /implement-ui
 → /expand-test-coverage → /run-tests → /run-eval
 → /close-eval-gaps (loop until fix-queue clear or needs_human)
 → /review-changes → /application-summary
```

## Human gates

| Gate | Blocks |
|------|--------|
| BRD **Accepted** | `/plan-sprint` |
| Plan **Approved** | `/implement-api`, `/implement-ui`, `/generate-ui-design` |
| Test-package **Approved** | `/run-tests` |
| Completeness ledger + fix-queue closed | `/review-changes` |

## Phase-changing commands (rule 10)

generate-brd, plan-sprint, implement-api, implement-ui, expand-test-coverage, run-tests, run-eval, close-eval-gaps, review-changes, application-summary

Satellites (no harness row): propose-architecture, generate-ui-design

## Eval

| Type | Command | Output |
|------|---------|--------|
| Product | `/run-eval` + `/close-eval-gaps` | `evals/product/results/<date>-completeness-ledger.md`, `<date>-fix-queue.json` |
| Factory meta | CI / maintainer | `evals/factory/` (see evals/factory/README.md) |

## Local stack

After `/implement-api` + `/implement-ui`: `bash scripts/run-local-stack.sh` then `bash scripts/verify-stack-health.sh` before Playwright (see `quality` skill).
