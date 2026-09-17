# Product evals

Port the **mechanism**, not another program's cases.

```
evals/product/
  STATUS.md
  templates/          # three-file case + fix-queue schema
  results/
  wave-1/<type>/SPEC.md
  wave-1/<type>/cases/<id>/prompt.md + GRADING.md
```

Modes: incremental (`/run-eval`), full-sweep (`/run-eval --mode=full-sweep`). `/expand-test-coverage` and `/review-changes` do **not** run eval-agent.
Incremental still includes always-on safety families (L1-04, L1-05, L1-08, L1-09, L4-03).
`config.yaml` `evals.rule_source` points at the project BRD or story pack.
`evals.test_commands` lists the real runners — never invent new ones.
Every run writes `evals/product/results/<date>-completeness-ledger.md`. A family with zero cases is FAIL, not green.
