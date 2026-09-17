# Factory meta-evals (claude starter kit)

Grade **this kit's** agents, commands, and rules — not the participant product.

Participant labs use **`/run-eval`** (product evals under `evals/product/`).

## Scope

**`.claude/`** starter kit (12 primaries, 12 commands). Copied from master `evals/factory/` with kit-scoped paths.

## When to run

- Before releasing an updated kit drop-in
- After editing agents, commands/prompts, or rules/instructions

## Eval suite

Catalog lock: 12 primaries; no public `/generate-backend`, `/generate-frontend`, `/wire-*`, `/write-tech-spec`, `/gap-analysis`, or `/phase-*`. Copilot docker/perf agents are not primaries.

| Eval | What it grades |
|------|----------------|
| `agent-contract-compliance` | Canonical headings including Audit and harness |
| `sod-boundaries` | 12 primaries; orchestrator/quality/docs Must NOTs |
| `catalog-lock` | Forbidden public commands; packs gone |
| `audit-trail-required` | Every public command mentions `artifacts/audit` |
| `harness-updated` | Rule 10 phase-changing commands require harness |
| `clarification-protocol` | Commands ask back when inputs missing |
| `provenance-anti-invention` | Intake agents require source locators |
| `loop-max-3-escalate` | Max 3 ceiling; early exit on iteration-1 pass |
| `eval-agent-completeness` | Completeness ledger; silence is not a pass |
| `test-case-before-script` | Cases md + BR coverage PASS before scripts |
| `template-hybrid-enforcement` | Vars YAML + render_check PASS; named agent-owned slots only |
| `self-approve-forbidden` | Humans own BRD/plan gates |
| `orchestrator-no-code` | Orchestrator does not implement features |
