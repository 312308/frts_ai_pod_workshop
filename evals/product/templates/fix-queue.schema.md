# fix-queue.schema

JSON array of objects:

| Field | Type | Notes |
|-------|------|-------|
| id | string | Stable item id |
| case_path | string | Path to the case folder |
| owning_agent | string | backend-agent, frontend-agent, api-agent, quality-agent, documentation-agent, eval-agent |
| expected | string | What should be true |
| actual | string | What was observed |
| auto_fix | bool | Orchestrator walks only `true` |
| status | enum | open / in_progress / closed / needs_human / waived |
| attempts | number | Max 3 |
| human_gate_reason | string | Required when auto_fix is false |
| category | enum | verdict_failed \| oracle_too_weak \| catalog_gap \| product_gap \| waived \| imported_finding |
| severity | enum | P0 \| P1 \| P2 \| P3 — required. P0/P1 security, secrets, authz, and imported findings are `auto_fix: false`. |

Orchestrator walks only `auto_fix: true` with status open/in_progress. `auto_fix: false` stays `needs_human` or `waived`. Missing-case `catalog_gap` may be `auto_fix: true` for eval-agent (write the case). P0/P1 cannot be waived by eval-agent.
