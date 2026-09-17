# Commands (12 shipped)

| Command | Phase | Primary |
|---------|-------|---------|
| `/generate-brd` | Intake | requirements-agent |
| `/propose-architecture` | Architecture (satellite) | planner-agent |
| `/plan-sprint` | Plan | planner-agent |
| `/generate-ui-design` | UI mockup from website reference | frontend-agent |
| `/implement-api` | Backend | backend-agent (java\|python) or golang-agent (golang), per `config.yaml` `implementation.backend.language` |
| `/implement-ui` | Frontend + wiring | frontend-agent |
| `/expand-test-coverage` | Test design | quality-agent |
| `/run-tests` | Test execution | quality-agent |
| `/run-eval` | Product AI eval | quality-agent |
| `/close-eval-gaps` | Eval fix loop (inner spine) | orchestrator-agent |
| `/review-changes` | Review | code-review-agent |
| `/application-summary` | Summary | documentation-agent |

`/close-eval-gaps` is invoked from `/run-eval` when `fix-queue.json` has open `auto_fix: true` items. May also be run directly to resume the queue.

Flow: [../../CLAUDE-FLOW.md](../../CLAUDE-FLOW.md)
