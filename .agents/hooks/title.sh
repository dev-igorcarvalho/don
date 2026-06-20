#!/usr/bin/env bash

# Read JSON payload from stdin
read -r JSON_INPUT
echo "INPUT: $JSON_INPUT" > /home/igor/Documents/projetos/don/.agents/hooks/title.log


# 1. Project/Context Name
PROJECT_DIR=$(echo "$JSON_INPUT" | jq -r '.workspace.project_dir // ""')
if [ -n "$PROJECT_DIR" ]; then
    PROJECT_NAME=$(basename "$PROJECT_DIR")
else
    PROJECT_NAME=$(basename "$PWD")
fi

# 2. Git Branch
BRANCH=$(git branch --show-current 2>/dev/null || git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "no-git")

# 3. Active AI Model
MODEL=$(echo "$JSON_INPUT" | jq -r '.model.display_name // .model.id // "Unknown"')

# 4. Current Task Status
TASK=$(echo "$JSON_INPUT" | jq -r '.task.description // .active_tool.name // .status // .state // ""')
if [ -z "$TASK" ] || [ "$TASK" = "idle" ] || [ "$TASK" = "Idle" ]; then
    TASK_STATUS="Idle"
elif [ "$TASK" = "working" ] || [ "$TASK" = "running" ] || [ "$TASK" = "Running" ]; then
    TASK_STATUS="Running..."
else
    TASK_STATUS="Running $TASK"
fi

# 5. Elapsed Time (using Session ID)
SESSION_ID=$(echo "$JSON_INPUT" | jq -r '.session_id // "default"')
START_FILE="/tmp/agy_session_start_${SESSION_ID}"

if [ ! -f "$START_FILE" ]; then
    date +%s > "$START_FILE"
fi

START_TIME=$(cat "$START_FILE")
NOW=$(date +%s)
ELAPSED=$((NOW - START_TIME))

if (( ELAPSED < 60 )); then
    DURATION="${ELAPSED}s"
elif (( ELAPSED < 3600 )); then
    MINUTES=$(( ELAPSED / 60 ))
    SECONDS=$(( ELAPSED % 60 ))
    DURATION="${MINUTES}m ${SECONDS}s"
else
    HOURS=$(( ELAPSED / 3600 ))
    MINUTES=$(( (ELAPSED % 3600) / 60 ))
    SECONDS=$(( ELAPSED % 60 ))
    DURATION="${HOURS}h ${MINUTES}m ${SECONDS}s"
fi

# Output the formatted window title using terminal escape codes
#OUTPUT="${PROJECT_NAME} (${BRANCH}) | ${TASK_STATUS} | ${MODEL} | ${DURATION}"
#echo "OUTPUT: $OUTPUT" >> /home/igor/Documents/projetos/don/.agents/hooks/title.log
#printf "\033]0;%s\007" "$OUTPUT"
echo "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

