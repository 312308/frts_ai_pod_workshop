# Hooks

- `beforeSubmitPrompt` scaffolds sprint folders for known `/command`s (`failClosed: false`).
- `beforeShellExecution` on commit/push runs artifact policy (`failClosed: true`).
- Duplicate checker: `tools/check-artifacts.mjs` + `.githooks/pre-commit`.
