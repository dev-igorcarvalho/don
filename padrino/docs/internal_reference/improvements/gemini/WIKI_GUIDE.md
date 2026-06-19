# Documentation-as-Code Wiki Guide

This guide explains how to create, implement, and use a linked Markdown wiki for long-term project persistence and knowledge management.

## 1. Core Philosophy
Instead of using external tools or vectorized databases, this project uses a **Markdown Wiki** stored directly in the repository.
- **Git-Backed:** Every change is versioned, peer-reviewed, and traceable.
- **Low Friction:** Written in standard Markdown, accessible by any text editor or CLI tool.
- **High Signal:** Explicit linking provides deterministic context for both humans and AI agents.

## 2. Structure
The wiki is organized hierarchically starting from an index file.

### Recommended Directory Layout
```text
docs/
├── index.md             # The entry point (Table of Contents)
├── architecture/        # System design, ADRs, and diagrams
├── domain/              # Business logic and glossary
├── operations/          # Deployment, CI/CD, and runbooks
└── guides/              # "How-to" for developers
```

## 3. Implementation Patterns

### Relative Linking
Always use relative paths for links to ensure they work in any environment (local, GitHub, or IDE).
- Good: `[Architecture Overview](./architecture/overview.md)`
- Bad: `[Architecture Overview](/home/user/docs/architecture/overview.md)`

### Frontmatter (Metadata)
Include a YAML block at the top of each file for categorization and searchability.
```markdown
---
title: System Overview
status: stable
last_updated: 2024-04-28
tags: [architecture, core]
---
```

### Atomic Files
Keep files focused on a single topic. If a file exceeds 500 lines, it's a sign to split it into sub-pages and link them from a parent directory.

### Internal Agent Knowledge (.gemini/wiki)
While `docs/` is standard for human-facing documentation, the wiki can also live in `.gemini/wiki/`. This is particularly useful for information that is primarily intended to guide AI agents like Gemini CLI.

#### Referencing from GEMINI.md
`GEMINI.md` serves as the root "brain" for the project. It should explicitly point to the wiki index to ensure I discover it immediately upon activation.

Example entry in `GEMINI.md`:
```markdown
## 🧠 Knowledge Base
All architectural decisions, domain rules, and developer guides are maintained in the [Internal Wiki](./.gemini/wiki/index.md).
```

## 4. How Gemini CLI Uses This
As an AI agent, I prioritize this wiki in my **Research** phase:
1. **Context Discovery:** I look for `GEMINI.md` and `docs/index.md` to map the system.
2. **Surgical Reads:** I use `grep_search` to find specific keywords across the wiki.
3. **Deterministic Navigation:** I follow links in `.md` files to understand dependencies without "hallucinating" relationships.

## 5. Usage Workflow

### For Humans
- **Adding Knowledge:** Create a new `.md` file in the relevant folder and add a link to it in `docs/index.md`.
- **Updating:** Treat documentation updates as part of the PR (Pull Request) process. If code changes, the wiki must change.

### For Gemini CLI
- You can ask me to: *"Update the architecture wiki to reflect the new messaging adapter"* or *"Extract a summary of the domain rules from the wiki."*

## 6. Maintenance
- **Pruning:** Periodically remove or archive outdated guides into an `archives/` folder.
- **Validation:** Use a markdown linter or link checker (like `markdown-link-check`) to ensure no broken paths exist.
