#!/usr/bin/env node
import { existsSync } from "node:fs";
import { spawnSync } from "node:child_process";

const mode = process.argv.includes("--staged") ? "staged" : "all";
let files = [];
if (mode === "staged") {
  const r = spawnSync("git", ["diff", "--cached", "--name-only"], { encoding: "utf8" });
  files = r.stdout.split("\n").filter(Boolean);
} else {
  const r = spawnSync("git", ["ls-files", "*.md"], { encoding: "utf8" });
  files = r.stdout.split("\n").filter(Boolean);
}
process.env.STAGED_FILES = files.join("\n");

const hook = ".claude/hooks/verify-artifacts.mjs";
if (!existsSync(hook)) {
  console.error(`Missing ${hook}. Re-run the install steps in README.md.`);
  process.exit(1);
}
const v = spawnSync("node", [hook], { stdio: "inherit" });
process.exit(v.status ?? 1);
