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
if ! git status --porcelain --untracked-files=all | grep -q -E '\.go"?$'; then
    echo -e "ℹ️  No .go file changes detected. Skipping lint checks." >&2
    echo '{"decision": "allow"}'
    exit 0
fi

# 1. Run go format
echo -e "🧹 ${GREEN}Running go fmt...${NC}"
go fmt ./...

# 2. Run go vet
echo -e "🔍 ${GREEN}Running go vet...${NC}"
go vet ./...

# 3. Run gosec (sec)
echo -e "🛡️  ${GREEN}Checking for gosec...${NC}"
if ! command -v gosec >/dev/null; then
    echo -e "⚠️  ${YELLOW}gosec not installed. Installing gosec...${NC}"
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi
gosec ./...

# 4. Check if golangci-lint is installed otherwise install it
echo -e "⚙️  ${GREEN}Checking golangci-lint...${NC}"
if ! command -v golangci-lint >/dev/null; then
    echo -e "📥 ${YELLOW}golangci-lint not found. Installing...${NC}"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
else
    echo -e "✅ ${GREEN}golangci-lint is already installed.${NC}"
fi

# 5. Check if golangci.yml exist in the current folder, otherwise warn the client
if [ ! -f ".golangci.yml" ]; then
    echo -e "⚠️  ${RED}Warning: .golangci.yml does not exist in the current folder!${NC}"
fi

# 6. Run golangci-lint run --new-from-rev=HEAD ./...
echo -e "🚦 ${GREEN}Running golangci-lint run --new-from-rev=HEAD ./...${NC}"
golangci-lint run --new-from-rev=HEAD ./...
LINT_STATUS=$?

# 7. Run go mod tidy
echo -e "📦 ${GREEN}Running go mod tidy...${NC}"
go mod tidy

# 8. Print if everything went well or not
if [ $LINT_STATUS -eq 0 ]; then
    echo -e "\n✨ ${GREEN}Everything went well! All checks passed.${NC}" >&2
    echo '{"decision": "allow"}'
else
    echo -e "\n❌ ${RED}Linter failed. Please fix the issues above.${NC}" >&2
    # Create a JSON response for the AfterAgent hook to trigger a retry with the errors
    # Note: We use >&2 for human-readable output to keep stdout clean for the JSON response
    REASON="${RED}The linter failed with the following status code: $LINT_STATUS. Please review the output above and fix the linting/formatting issues.${NC}"
    echo "{\"decision\": \"deny\", \"reason\": \"$REASON\"}"
    exit 0 # Exit 0 so the CLI processes the JSON decision
fi
