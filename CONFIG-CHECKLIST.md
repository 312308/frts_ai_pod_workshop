# config.yaml checklist

Copy `.claude/config.yaml.example` → `.claude/config.yaml` and fill before running commands.

| Key | Your value | Recommended default |
|-----|------------|------------------|
| `project.name` | | *(required — no placeholders)* |
| `project.domain` | | bounded context name |
| `modules.api` | | `modules/api` |
| `modules.ui` | | `modules/ui` |
| `modules.openapi` | | `contracts/openapi/<service>-v1.yaml` |
| `evals.rule_source` | | path to Accepted BRD after `/generate-brd` |
| `evals.test_commands.api` | | `cd <api-module-root> && mvn test` |
| `evals.test_commands.ui` | | `cd <ui-module-root> && npm test` |
| `evals.test_commands.e2e` | | `cd <ui-module-root> && npm run test:e2e` |
| `pipeline.eval_fix_command` | | `close-eval-gaps` |
| `local_stack.start_script` | | `scripts/run-local-stack.sh` |
| `quality_gates.api_unit_line_coverage_min` | | `0.70` |

**Hard stop:** orchestrator refuses product code while placeholders remain (rule 01).
