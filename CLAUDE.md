# Claude AI Factory — Starter Kit Governance

Portable **12-command spine** for Claude Code. **11 public commands** + shipped **`/close-eval-gaps`** (eval fix loop).

Do **not** assume a vendor, warehouse, optical, AS400, or HIPAA domain. Fill `.claude/config.yaml` before generating product code.

Hard guarantees are enforced by `.claude/hooks/` and git hooks (`.githooks/`). Claude Code loads `.claude/rules/*.mdc`.

---

## 1. Prime directives

1. **Provenance over invention.** Never invent business behavior, domain rules, UI copy, API fields, DDL columns, or error semantics. Cite a source or register `ASSUMPTION-<id>`. Uncited behavior is blocked from Ready.
2. **Plan → Act → Validate → Refine** (build/quality) or **Plan → Act → Observe → Reflect** (intake). Max 3 ceiling; stop when iteration 1 Validate/Observe passes. Refine/Reflect only from findings. Escalate after iteration 3 with a documented gap.
3. **Resolve before ask.** When a governed plan, sprint plan, or story artifact is supplied, load plan + manifest + logs first. Ask only about Open critical gaps or `TBD` bindings. Skill: `.claude/skills/plan-as-input/`. Protocol: `.claude/templates/plan-as-input-protocol.md`.
4. **Least privilege.** Destructive or irreversible actions require a human gate. Secrets never enter code, logs, prompts, or artifacts.
5. **Everything is an artifact.** Persist output to a versioned file with timestamp, iteration count, and validation summary. Write `artifacts/audit/<YYYY-MM-DD>-<agent>-<task>.md`. Only `harness-updater` writes `docs/harness/**`.
6. **Traceability.** Rule → story → contract/DDL/UI → code → test → evidence.

## 2. Command spine

```
/generate-brd → /propose-architecture → /plan-sprint
  → /generate-ui-design + /implement-api → /implement-ui
  → /expand-test-coverage → /run-tests → /run-eval
  → /close-eval-gaps (loop until fix-queue clear)
  → /review-changes → /application-summary
```

Execute spine:

```
/implement-api → /implement-ui → /expand-test-coverage → [human: test-package Approved] → /run-tests → /run-eval → /close-eval-gaps → /review-changes → /application-summary
```

Operator map: [CLAUDE-FLOW.md](CLAUDE-FLOW.md).

## 3. Orchestration

- Route slash commands through **orchestrator-agent** (`.claude/agents/orchestrator-agent.md`) by **explicit name**. Do not rely on description-match auto-delegation across specialists.
- Spawn other primaries from `.claude/agents/*.md` via the Claude Code **Agent** tool by explicit name (or the command's listed primary).
- Spawn specialists from `.claude/subagents/<name>.md` as the task spec. Specialists must not be selected as slash-command primaries.
- Prefer a **Skill** (`.claude/skills/`) for reusable procedure; a hook (`.claude/hooks/`) or git hook for a hard invariant.
- Subagent returns structured IO + confidence; parent rejects below-gate output (max 3).
- **Agent-owned implement:** Java/Next.js per `docs/architecture/stack-conventions.md` — no J2 template render pipeline in kit DoD.

## 4. Model routing

- **opus** — orchestrator, intake judgment, review, security, release, quality coordination.
- **sonnet** — implementation and artifact generation.
- **haiku** — narrow mechanical collectors/formatters.

Do not hardcode a vendor model into product code or factory contracts.

## 5. Human gates (do not auto-proceed)

- `/generate-brd` → BRD **Accepted** (or waiver in `docs/harness/decisions.md`)
- `/plan-sprint` → plan **Approved**
- `/expand-test-coverage` → test-package **Approved** before `/run-tests`
- Completeness ledger from `/run-eval` before `/review-changes`
- Any `deny`/`ask` permission, secret access, or destructive op

## 6. Imported always-on rules

These `.mdc` files use `alwaysApply: true` (loaded with this file):

- `.claude/rules/00-global-governance.mdc`
- `.claude/rules/01-project-context.mdc`
- `.claude/rules/02-security-and-compliance.mdc`
- `.claude/rules/07-testing-and-quality.mdc`
- `.claude/rules/09-agent-orchestration.mdc`
- `.claude/rules/10-audit-trail.mdc`
- `.claude/rules/11-intake-planning-governance.mdc`
- `.claude/rules/22-behavior-provenance-and-anti-invention.mdc`
- `.claude/rules/31-untrusted-intake.mdc`
- `.claude/rules/kit-governance.mdc`

@.claude/rules/00-global-governance.mdc
@.claude/rules/01-project-context.mdc
@.claude/rules/02-security-and-compliance.mdc
@.claude/rules/07-testing-and-quality.mdc
@.claude/rules/09-agent-orchestration.mdc
@.claude/rules/10-audit-trail.mdc
@.claude/rules/11-intake-planning-governance.mdc
@.claude/rules/22-behavior-provenance-and-anti-invention.mdc
@.claude/rules/31-untrusted-intake.mdc
@.claude/rules/kit-governance.mdc

## Kit scope

This kit ships **12 commands** per [KIT-SCOPE.md](KIT-SCOPE.md). Commands such as `/fix-issue`, `/deliver-sprint`, `/generate-user-story`, security scans, and handoffs are **not** included.
