---
trigger: model_decision
globs:
  - "src/**/*.go"
  - "scripts/*.sh"
description: Complete template demonstrating all possible Antigravity frontmatter keys.
---

# Antigravity Rules Frontmatter Reference

This file demonstrates all possible YAML frontmatter parameters used to configure local agent rules in the `.agents/rules/` directory.

## Frontmatter Parameters

*   **`trigger`** (`string`, Required): Determines how the rule is loaded into the agent's context. Supported options are:
    *   `always_on`: The rule is unconditionally loaded for every prompt in the workspace.
    *   `glob`: The rule is only loaded when the agent reads or writes files matching patterns in the `globs` list.
    *   `manual`: The rule is never loaded automatically, but can be loaded via `@rule` mentions in user prompts.
    *   `model_decision`: The agent dynamically decides whether to load the rule based on its relevance to the active task (matching the prompt against the `description` field).
*   **`description`** (`string`, Required for `model_decision`): Summarizes what the rule does or targets. The agent reads this summary to evaluate relevance during dynamic rule selection.
*   **`globs`** (`array` of strings, Required for `glob`): Specifies glob-matching patterns targeting files or file types (e.g. `["*.go", "scripts/*.sh"]`).
