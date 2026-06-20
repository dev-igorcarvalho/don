#!/bin/bash
# Stop hook script for Don monorepo to update AGENTS.md before exit.

cd "$(git rev-parse --show-toplevel)" || exit 1
# 1. Prevent recursion loop
if [ "$AGY_STOP_HOOK_ACTIVE" = "true" ]; then
  exit 0
fi

# 2. Check if there are changes in the three project directories
CHANGES=$(git status --porcelain artefatti consiglieri padrino)
if [ -z "$CHANGES" ]; then
  echo "No changes detected in artefatti, consiglieri, or padrino. Skipping AGENTS.md update."
  exit 0
fi

# 3. Export environment variable to child processes
export AGY_STOP_HOOK_ACTIVE=true

TEMP_DIFF_FILE=".agents/hooks/temp_session_diffs.md"

# Ensure cleanup on script exit
trap 'rm -f "$TEMP_DIFF_FILE"' EXIT

# 4. Generate deterministic diff and change summary
echo "# Session Diffs & Changed Files" > "$TEMP_DIFF_FILE"
echo "Generated on: $(date)" >> "$TEMP_DIFF_FILE"
echo "" >> "$TEMP_DIFF_FILE"

for PROJECT in artefatti consiglieri padrino; do
  echo "## Project: $PROJECT" >> "$TEMP_DIFF_FILE"
  echo "### Changed Files" >> "$TEMP_DIFF_FILE"
  
  # List changed and untracked files for the project
  PROJECT_CHANGES=$(git status --porcelain "$PROJECT")
  if [ -z "$PROJECT_CHANGES" ]; then
    echo "No changes." >> "$TEMP_DIFF_FILE"
  else
    echo "\`\`\`" >> "$TEMP_DIFF_FILE"
    echo "$PROJECT_CHANGES" >> "$TEMP_DIFF_FILE"
    echo "\`\`\`" >> "$TEMP_DIFF_FILE"
  fi
  echo "" >> "$TEMP_DIFF_FILE"
  
  # Append git diff for tracked files
  echo "### Git Diff" >> "$TEMP_DIFF_FILE"
  PROJECT_DIFF=$(git diff HEAD -- "$PROJECT")
  if [ -z "$PROJECT_DIFF" ]; then
    echo "No diff (untracked or unmodified files only)." >> "$TEMP_DIFF_FILE"
  else
    echo "\`\`\`diff" >> "$TEMP_DIFF_FILE"
    echo "$PROJECT_DIFF" >> "$TEMP_DIFF_FILE"
    echo "\`\`\`" >> "$TEMP_DIFF_FILE"
  fi
  echo "" >> "$TEMP_DIFF_FILE"
done

# 5. Call agy non-interactively with a coordinator prompt
echo "Running agy to update AGENTS.md with project subagents..."
agy --dangerously-skip-permissions -p "Review the changes described in $TEMP_DIFF_FILE. For each project in [artefatti, consiglieri, padrino] that has changes, launch a focused subagent to analyze the changes, read key modified files to identify coding conventions/guidelines, and report back. Consolidate their findings to update the main AGENTS.md file. Keep all other sections intact."
