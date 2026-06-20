# Spec: Go Change Detection Linter

## 1. Overview
Currently, the lint script in `.agents/lint.sh` runs for every single agent invocation through the `AfterAgent` hook, regardless of what files were changed. This spec outlines the addition of a Git-based check to skip linting/formatting unless changes to `.go` files are detected.

## 2. Goals
- Skip execution of formatting and lint checks when no `.go` files have been added, modified, or deleted in the workspace.
- Ensure that if the checks are skipped, the hook still succeeds (`exit 0`) and returns `{"decision": "allow"}` to avoid breaking the tool agent workflow.

## 3. Detailed Design

### Check Logic
We will query Git status in porcelain format:
```bash
git status --porcelain
```
This output lists staged changes, unstaged changes, and untracked files. We search this list for any files ending in `.go`:
```bash
git status --porcelain | grep -q '\.go$'
```
If the exit status of the pipeline is non-zero, it means no `.go` files were modified, created, or deleted. In this case, we immediately output:
```json
{"decision": "allow"}
```
to standard output (stdout), log a message to standard error (stderr), and exit with code `0`.

### Changes to `.agents/lint.sh`
We will insert the following block right after setting the output colors:

```bash
# Check if any .go files have changes (added, modified, deleted, staged, or unstaged)
if ! git status --porcelain | grep -q '\.go$'; then
    echo -e "ℹ️  No .go file changes detected. Skipping lint checks." >&2
    echo '{"decision": "allow"}'
    exit 0
fi
```

## 4. Verification Plan
1. Run `./.agents/lint.sh` with no changes to verify it skips and prints `{"decision": "allow"}`.
2. Modify a `.go` file and run `./.agents/lint.sh` to verify it runs the formatters and linters.
