Read `java_master_templates/module-variable-schema.yaml`, `java_master_templates/template-registry.v1.yaml`, `java_test_templates/test-variable-schema.yaml`, `tools/templates/render_check.py`, and kit commands `/implement-api`, `/apply-requirements`, `/fix-issue`, `/run-tests`, `/review-changes` under `.claude/`. Also backend-agent, quality-agent, code-review-agent, spring-module-generator, spring-test-generator, and kit rules 03/23.

Pass only if all twins require:

1. Persist `artifacts/templates/<resource>-vars.yaml` (no `# AMBIGUITY`).
2. `tools/templates/render_check.py` PASS before DoD on Java.
3. Controller / CRUD / error envelope stay template-owned.
4. Cited domain logic only in service `AGENT-OWNED DOMAIN` markers, ValidationRule, or a new cited type.
5. Registry tests keep harness; agent-owned extras are sibling classes (or marked blocks) citing `docs/test-cases.md`.
6. `/review-changes` Blocks on missing or FAIL render_check when Java changed or vars YAML exists.
