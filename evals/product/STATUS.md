# Product eval status

Fill per consuming app. Status legend: `EXISTS` (pointer to product test), `NEW` (no case yet), `IMPROVED`, `MISSING` (family empty or rule has no row — FAIL, not green), `WAIVED` (human, with reason; never P0/P1).

Always-on safety families must not stay `NEW`/`MISSING` after `/run-eval` unless queued as `catalog_gap`.

| Type | Status | Notes |
|------|--------|-------|
| L1-01-pure-function-correctness | NEW | add project cases |
| L1-02-edge-case-coverage | NEW | |
| L1-03-rest-endpoint-correctness | NEW | |
| L1-04-input-validation | NEW | |
| L1-05-error-handling | NEW | |
| L1-06-test-generation-inclusion | NEW | |
| L1-07-style-conventions | NEW | |
| L1-08-security-anti-pattern | NEW | |
| L1-09-secret-handling | NEW | |
| L1-10-requirement-completeness | NEW | pointer cases to product tests |
| L1-13-prompt-ab-harness | NEW | optional |
| L4-01-pipeline-happy-path | NEW | |
| L4-02-gate-blocks-low-coverage | NEW | JaCoCo check enabled |
| L4-03-gate-blocks-security | NEW | |
| L4-04-gate-blocks-forbidden-paths | NEW | CODEOWNERS |
| LX-01-judge-human-agreement | NEW | |
| LX-02-anchor-ambiguity | NEW | |
| LX-03-judge-prompt-versioning | NEW | |
