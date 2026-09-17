# Loop max-3 then escalate

Fail if a primary agent or public command in `.claude/` omits max 3 iterations or ships the first Act without Validate/Observe.

Pass if max 3 is a ceiling: iteration 1 Validate/Observe pass → stop remaining iterations; Refine/Reflect only from findings.
