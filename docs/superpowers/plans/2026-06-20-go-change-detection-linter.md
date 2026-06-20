# Go Change Detection Linter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Modify `.agents/lint.sh` to only execute linting steps if any `.go` files are modified, created, or deleted.

**Architecture:** Use `git status --porcelain` to query for staged, unstaged, and untracked changes, and filter for files ending in `.go` via `grep`. Exit early with code `0` and print a JSON response containing `{"decision": "allow"}` if no matching files are found.

**Tech Stack:** Bash

## Global Constraints
- Target Go Version: Go 1.26+
- Script location: `.agents/lint.sh`
- Hook format output: Must print `{"decision": "allow"}` to stdout if successful or skipped.

---

### Task 1: Add change detection logic to lint.sh

**Files:**
- Modify: `.agents/lint.sh`

**Interfaces:**
- Consumes: Workspace Git status
- Produces: JSON string to stdout: `{"decision": "allow"}` if skipping or if lint succeeds.

- [ ] **Step 1: Modify `.agents/lint.sh` to add the change-detection check**

Add the Git status check right after the color definitions.

```bash
# Check if any .go files have changes (added, modified, deleted, staged, or unstaged)
if ! git status --porcelain | grep -q '\.go$'; then
    echo -e "ℹ️  No .go file changes detected. Skipping lint checks." >&2
    echo '{"decision": "allow"}'
    exit 0
fi
```

Here is the complete modified file layout for reference around the edits:

```bash
#!/bin/bash

# Get the current folder pwd
CURRENT_DIR=$(pwd)

echo "🚀 Initializing linter script in: $CURRENT_DIR"

# Set colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if any .go files have changes (added, modified, deleted, staged, or unstaged)
if ! git status --porcelain | grep -q '\.go$'; then
    echo -e "ℹ️  No .go file changes detected. Skipping lint checks." >&2
    echo '{"decision": "allow"}'
    exit 0
fi

# 1. Run go format
echo -e "🧹 ${GREEN}Running go fmt...${NC}"
go fmt ./...
...
```

- [ ] **Step 2: Run verification with no changes**

Ensure git is clean (or at least has no Go changes). Run:
```bash
./.agents/lint.sh
```
Expected output:
```
🚀 Initializing linter script in: <path>
ℹ️  No .go file changes detected. Skipping lint checks.
{"decision": "allow"}
```
Verify the exit status:
```bash
echo $?
```
Expected: `0`

- [ ] **Step 3: Run verification with a simulated Go file change**

Create a temporary dummy Go file:
```bash
touch temp_test_dummy.go
```
Run:
```bash
./.agents/lint.sh
```
Expected output:
The script should bypass the skip condition and attempt to run formatting and linting (e.g. `🧹 Running go fmt...`).
Clean up the temporary file:
```bash
rm temp_test_dummy.go
```

- [ ] **Step 4: Commit the change**

Use semantic commit format to commit the implementation plan and the modified script.
Run:
```bash
git add docs/superpowers/plans/2026-06-20-go-change-detection-linter.md .agents/lint.sh
git commit -m "feat(linter): skip checks if no .go files changed"
```
