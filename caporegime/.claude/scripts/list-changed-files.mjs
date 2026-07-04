#!/usr/bin/env node
// Lists files that differ between the current checkout and the repo's
// main/master branch, so quality tooling can target only what changed.
// Usage: npx ./scripts/list-changed-files.mjs [--ext .go]
import { execSync } from "node:child_process";
import path from "node:path";

function run(cmd) {
  try {
    return execSync(cmd, { stdio: ["ignore", "pipe", "ignore"] }).toString().trim();
  } catch {
    return "";
  }
}

function resolveBaseRef() {
  run("git fetch origin main master --quiet");
  const candidates = ["origin/main", "origin/master", "main", "master"];
  for (const ref of candidates) {
    if (run(`git rev-parse --verify ${ref}`)) return ref;
  }
  throw new Error("Could not resolve a main/master ref to diff against.");
}

function main() {
  const extFlagIndex = process.argv.indexOf("--ext");
  const ext = extFlagIndex !== -1 ? process.argv[extFlagIndex + 1] : null;

  const base = resolveBaseRef();
  const mergeBase = run(`git merge-base HEAD ${base}`) || base;

  const committed = run(`git diff --name-only --diff-filter=ACMR ${mergeBase}...HEAD`);
  const staged = run("git diff --name-only --diff-filter=ACMR --cached");
  const unstaged = run("git diff --name-only --diff-filter=ACMR");
  const untracked = run("git ls-files --others --exclude-standard --full-name");

  const repoRoot = run("git rev-parse --show-toplevel");

  const files = new Set(
    [committed, staged, unstaged, untracked]
      .flatMap((block) => block.split("\n"))
      .map((f) => f.trim())
      .filter(Boolean)
      // git diff/ls-files report paths relative to the repo root; normalize
      // them to be relative to the caller's cwd (e.g. a package subdir).
      .map((f) => path.relative(process.cwd(), path.join(repoRoot, f)))
  );

  const result = [...files]
    .filter((f) => (ext ? f.endsWith(ext) : true))
    .sort();

  for (const file of result) {
    console.log(file);
  }
}

main();
