# Kit scope (Claude Code starter kit)

## Commands (12)

| # | Command | Primary agent |
|---|---------|---------------|
| 1 | `/generate-brd` | requirements-agent |
| 2 | `/propose-architecture` | planner-agent |
| 3 | `/plan-sprint` | planner-agent |
| 4 | `/generate-ui-design` | frontend-agent |
| 5 | `/implement-api` | backend-agent |
| 6 | `/implement-ui` | frontend-agent |
| 7 | `/expand-test-coverage` | quality-agent |
| 8 | `/run-tests` | quality-agent |
| 9 | `/run-eval` | quality-agent |
| 10 | `/close-eval-gaps` | orchestrator-agent |
| 11 | `/review-changes` | code-review-agent |
| 12 | `/application-summary` | documentation-agent |

## Rules (20)

Shipped under `.claude/rules/` — see `kit-governance.mdc` for always-on set.

## Omitted (not in this kit)

`/fix-issue`, `/deliver-sprint`, `/generate-user-story`, `/figma-to-nextjs`, security scan commands, handoffs, and their dedicated rules/templates.
