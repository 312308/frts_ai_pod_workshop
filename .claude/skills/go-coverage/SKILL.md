---
name: go-coverage
description: Run Go coverage for the API module, interpret line/branch gaps, and recommend unit tests. Use when improving Golang API coverage or reporting metrics.
---

# Go Coverage

## Run
Use the project's Go toolchain (see `config.yaml` `implementation.backend.module_root`). Typical:

```bash
cd <api-module-root>
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Parse `coverage.out` for per-function statement coverage; use `go tool cover -html=coverage.out` for a browsable gap view.

## Priority for gaps
1. Validation / domain rules (one test per branch)
2. Application services
3. Parsers / calculators / adapters

## Gates
`.claude/config.yaml` → `quality_gates.api_unit_line_coverage_min` and `api_unit_branch_coverage_target`. Do not lower them.

Run after unit/IT tests; use with live stack when IT probes run (skill: `quality`).

## Subagent
`.claude/subagents/go-coverage-analyzer.md`
