Read `.claude/subagents/eval-agent.md`. Also `evals/product/templates/completeness-ledger.md` and `.claude/skills/eval-suite/SKILL.md`.

Pass only if the kit spec requires: eval-agent runs via `/run-eval` only (not expand or review); completeness ledger; always-on safety families L1-04, L1-05, L1-08, L1-09, L4-03 in incremental mode; agent-result families; silence / zero-case family is FAIL not green; no invented vulns; P0/P1 cannot be waived; parent rejects an omitted ledger.
