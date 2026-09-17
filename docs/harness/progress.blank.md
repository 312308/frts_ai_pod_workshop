# Harness Progress

| Field | Value |
|-------|-------|
| **Updated** | {{YYYY-MM-DD}} |
| **Last command** | Factory installed |

## Pipeline

| Stage | Status |
|-------|--------|
| Factory installed | **Done** |
| `/generate-brd` | Pending |
| Plan (`/plan-sprint`) | Pending |
| `/implement-api` | Pending |
| `/implement-ui` | Pending |
| `/expand-test-coverage` | Pending |
| `/run-tests` | Pending (after test-package Approved) |
| `/run-eval` | Pending |
| `/review-changes` | Pending |
| `/application-summary` | Pending (optional satellite) |
| `/application-summary` | Pending |

## Notes

- Only `harness-updater` writes this file.
- Status values: Pending / In progress / Done / Blocked / Skipped (with audit pointer).
