---
name: todo-tracker
description: Collects, translates, and documents task comments from the codebase into a structured issues.md file. Use when the user wants to track pending tasks, update the project roadmap, or audit pending items with detailed context.
---

# Task Tracker

This skill automates the discovery and documentation of task comments (e.g., // task-marks) across the codebase.

## Workflow

### 1. Discovery
Search for all task comments using `grep_search` with the pattern `//\s*(?i)todo:?.*`.

### 2. Contextual Analysis
For each found task:
- Identify the file and line number.
- Read the surrounding code (approximately 10 lines above and below) to understand the technical intent.
- If the comment is not in English, translate it to clear, technical English.

### 3. Documentation Pattern
Maintain or create `issues.md` in the project root following this exact structure:

#### File Header
```markdown
# Project Issues & Task List

## Index

| Issue | File | Line |
|-------|------|------|
| [Issue Title](#anchor-link) | `path/to/file.go` | 123 |
...
```

#### Detailed Sections
For each issue, create a section below the index:
```markdown
---

### Issue Title
**File:** `path/to/file.go:123`

[Detailed description of the task, why it's needed, and technical context derived from the code analysis.]
```

## Rules
- **Incremental Updates:** If `issues.md` already exists, append new items and update the index. Do not remove existing issues unless they are no longer present in the code.
- **Title Generation:** Create concise, professional titles for each issue (e.g., "Decouple Server from Echo" instead of "Review this because is coupled").
- **Professional Tone:** Use a senior engineer's tone for descriptions. Focus on the "why" and "how" of the proposed fix.
- **Link Integrity:** Ensure anchor links in the index table correctly point to the corresponding headers.
