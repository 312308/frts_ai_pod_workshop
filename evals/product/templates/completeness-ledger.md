# Completeness ledger

Required output of eval-agent: `evals/product/results/<date>-completeness-ledger.md`.

Silence is a finding. A family with zero cases is FAIL, not green.

## Header

| Field | Value |
|-------|-------|
| Date | YYYY-MM-DD |
| Mode | incremental \| full-sweep |
| Rule source | path from config `evals.rule_source` |
| Iteration | 1–3 (1 if first Validate passed) |

## Always-on safety families

Must appear in **both** modes: `L1-04-input-validation`, `L1-05-error-handling`, `L1-08-security-anti-pattern`, `L1-09-secret-handling`, `L4-03-gate-blocks-security`.

## Rows

| rule_id | family | case_path | status | verdict | severity | owner | notes |
|---------|--------|-----------|--------|---------|----------|-------|-------|

- `status`: EXISTS \| NEW \| IMPROVED \| WAIVED \| MISSING \| imported_finding
- `verdict`: pass \| fail \| not_run \| sanity_error \| catalog_gap
- `severity`: P0 \| P1 \| P2 \| P3
- `WAIVED` requires a human reason and cannot be used for P0/P1
- `MISSING` / `NEW` / `catalog_gap` must appear on the fix-queue
- `imported_finding` is copied from an existing security-agent artifact (cite path). Do not invent.

## Pass rule

Ledger is complete only when every in-scope `rule_id` from `rule_source` and every always-on family has a row. Parent rejects the eval run if this file is absent or any always-on family is omitted.
