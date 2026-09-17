# Three-file eval case

Use when a BRD/story rule has **no** existing product test.

| File | Purpose |
|------|---------|
| `prompt.md` | Numbered SHALL/SHOULD requirements the implementation must satisfy |
| `reference_solution.*` | Style/shape only — not copied into product code |
| `reference.test.*` | The verdict: must pass against a correct implementation |

If an existing test already asserts the rule, write a **pointer case** (`prompt.md` + `GRADING.md` only) with status `EXISTS` instead.
