# Claude Code starter kit

Self-contained **12-command** factory (11 public + `/close-eval-gaps`). Default stack: Java 21 / Spring Boot 3 + Next.js.

## Install (manual copy)

```bash
ASSETS="$(cd "$(dirname "$0")" && pwd)"

cp -R "$ASSETS/.claude" ./
cp "$ASSETS/CLAUDE.md" ./
cp "$ASSETS/KIT-SCOPE.md" ./
cp "$ASSETS/CONFIG-CHECKLIST.md" ./
cp -R "$ASSETS/.githooks" ./
cp -R "$ASSETS/tools" ./
cp -R "$ASSETS/scripts" ./
cp "$ASSETS/requirements.txt" ./
chmod +x scripts/*.sh
cp -R "$ASSETS/evals" ./
cp -R "$ASSETS/docs/harness" docs/harness/
cp -R "$ASSETS/docs/architecture" docs/architecture/
cp "$ASSETS/docs/harness/progress.blank.md" docs/harness/progress.md

git config core.hooksPath .githooks
```

Merge `.claude/config.yaml.example` → `.claude/config.yaml`. See [CONFIG-CHECKLIST.md](CONFIG-CHECKLIST.md).

## Python tooling

**Python 3.9+** is required for kit scripts under `scripts/` and `tools/design/`.

| Script | Dependencies | Used by |
|--------|--------------|---------|
| `scripts/intake-safety-scan.py` | stdlib only | `/generate-brd`, untrusted `/expand-test-coverage` intake |
| `tools/design/capture-reference.py` | Playwright (optional) | `/generate-ui-design` URL capture — fallback: supply `screenshot_path` |

After copying the kit:

```bash
python3 -m venv .venv          # optional but recommended
source .venv/bin/activate      # Windows: .venv\Scripts\activate
pip install -r requirements.txt
playwright install chromium    # skip if you will not use URL capture
```

Verify:

```bash
python3 scripts/intake-safety-scan.py docs/harness   # exit 0 = clean
python3 tools/design/capture-reference.py --help
```

## Commands (12)

| # | Command | Primary |
|---|---------|---------|
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

Flow: [CLAUDE-FLOW.md](CLAUDE-FLOW.md). Scope: [KIT-SCOPE.md](KIT-SCOPE.md). Command index: [.claude/commands/README.md](.claude/commands/README.md).

## Skills (14)

Under `.claude/skills/<name>/SKILL.md`. Pipeline-critical: `quality`, `playwright-e2e`, `eval-suite`, `api`, `br-coverage`, `jacoco-coverage`.

## Lab flow

1. Fill config (see CONFIG-CHECKLIST.md)
2. `/generate-brd` → **BRD Accepted**
3. `/propose-architecture` → copy config bindings
4. `/plan-sprint` → **Approved**
5. `/generate-ui-design` ∥ `/implement-api`
6. `/implement-ui` with `mockup_path`
7. `/expand-test-coverage` → test-package **Approved**
8. `bash scripts/run-local-stack.sh` && `bash scripts/verify-stack-health.sh`
9. `/run-tests` → `/run-eval` → `/close-eval-gaps` → `/review-changes` → `/application-summary`

**Outcome:** Running, validated application — stack healthy, tests green, fix-queue closed.
