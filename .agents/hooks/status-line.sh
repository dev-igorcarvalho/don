#!/usr/bin/env bash

read -r JSON_INPUT
echo "$JSON_INPUT" > /home/igor/Documents/projetos/don/.agents/hooks/statusline.log

CONTEXT=$(echo "$JSON_INPUT" | jq -r '(.workspace.context_pct // .context_window.used_percentage // 0) | round')
IN_T=$(echo "$JSON_INPUT" | jq -r '.context_window.total_input_tokens // "0"')
OUT_T=$(echo "$JSON_INPUT" | jq -r '.context_window.total_output_tokens // "0"')
QUOTA_5H=$(echo "$JSON_INPUT" | jq -r '.quota["gemini-5h"].remaining_fraction | if . == null then "N/A" else (. * 100 | round) end')
QUOTA_WEEKLY=$(echo "$JSON_INPUT" | jq -r '.quota["gemini-weekly"].remaining_fraction | if . == null then "N/A" else (. * 100 | round) end')
BRANCH=$(git branch --show-current 2>/dev/null || git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "no-git")

COLOR_RESET="\e[0m"
COLOR_BRANCH="\e[36m"
COLOR_QUOTA="\e[32m"
COLOR_TOKENS="\e[33m"

if (( CONTEXT <= 25 )); then
    COLOR_CONTEXT="\e[94m" # Bright Blue
elif (( CONTEXT <= 50 )); then
    COLOR_CONTEXT="\e[33m" # Yellow
elif (( CONTEXT <= 75 )); then
    COLOR_CONTEXT="\e[38;5;208m" # Orange
else
    COLOR_CONTEXT="\e[31m" # Red
fi

printf "${COLOR_BRANCH}[Branch: %s]${COLOR_RESET} ── ${COLOR_QUOTA}[5h Quota: %s%% | Weekly: %s%%]${COLOR_RESET} ── ${COLOR_CONTEXT}[Context: %s%%]${COLOR_RESET} ── ${COLOR_TOKENS}Tokens: [In: %s| Out: %s]${COLOR_RESET}\n" \
    "$BRANCH" "$QUOTA_5H" "$QUOTA_WEEKLY" "$CONTEXT" "$IN_T" "$OUT_T"
