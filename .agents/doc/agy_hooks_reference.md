# Antigravity CLI (`agy`) Lifecycle Hooks Reference

This document provides a comprehensive configuration guide and example for using lifecycle hooks in the Antigravity CLI (`agy`).

## Configuration Locations
Hooks can be defined globally or project-specifically:
* **Global:** [~/.gemini/antigravity-cli/settings.json](file:///home/igor/.gemini/antigravity-cli/settings.json)
* **Project / Workspace:** [.agents/settings.json](file:///home/igor/Documents/projetos/don/.agents/settings.json)

---

## Supported Hook Types

The `type` field in the hook definition determines how the hook logic is executed:

* **`command`**: Executes a local shell script (e.g. Bash/Python). Used for deterministic tasks like linting, testing, and system verification.
* **`http`**: Sends an HTTP POST request to an external endpoint containing the event JSON payload. Useful for integrating webhooks, Slack alerts, CI/CD systems, and remote audit loggers.
* **`mcp_tool`**: Invokes an Model Context Protocol (MCP) tool directly.
* **`prompt`**: Sends a prompt to an LLM to evaluate logic using judgment-based metrics (e.g., matching security patterns, checking code style constraints).
* **`agent`**: Spawns a dedicated AI subagent to run multi-turn processes (like checking files, resolving errors, or running multiple CLI tools recursively).

---

## Comprehensive Hooks Configuration Example

```json
{
  "hooks": {
    "enabled": true,
    
    "SessionStart": [
      {
        "name": "Init Workspace",
        "type": "command",
        "command": "echo 'Session started. Initializing...'"
      }
    ],
    
    "BeforeAgent": [
      {
        "matcher": "*",
        "hooks": [
          {
            "name": "Audit Prompt Security",
            "type": "command",
            "command": "python3 .gemini/scripts/audit_prompt.py"
          },
          {
            "name": "Log Remote Start Event",
            "type": "http",
            "url": "https://api.internal.dev/hooks/agent-start",
            "timeout": 15
          }
        ]
      }
    ],

    "BeforeToolSelection": [
      {
        "name": "Enforce Restricted Mode",
        "type": "command",
        "command": "bash .gemini/scripts/restrict_tools.sh"
      }
    ],

    "BeforeModel": [
      {
        "name": "Pre-process Prompt Context",
        "type": "command",
        "command": "python3 .gemini/scripts/preprocess_model_input.py"
      }
    ],

    "BeforeTool": [
      {
        "matcher": "run_shell_command",
        "hooks": [
          {
            "name": "Validate Bash Commands",
            "type": "command",
            "command": "python3 .gemini/scripts/validate_bash.py"
          }
        ]
      },
      {
        "matcher": "write_file",
        "hooks": [
          {
            "name": "Pre-Write Code Style Prompt Evaluator",
            "type": "prompt",
            "prompt": "Analyze the proposed code modifications. Do they follow standard naming conventions for the project?"
          }
        ]
      }
    ],

    "AfterTool": [
      {
        "matcher": "write_file",
        "hooks": [
          {
            "name": "Post-Write Linter",
            "type": "command",
            "command": ".gemini/lint.sh"
          }
        ]
      }
    ],

    "AfterModel": [
      {
        "name": "Sensitive Content Redactor",
        "type": "command",
        "command": "python3 .gemini/scripts/redact_output.py"
      }
    ],

    "AfterAgent": [
      {
        "matcher": "*",
        "hooks": [
          {
            "name": "Trigger Complex Post-Check Subagent",
            "type": "agent",
            "prompt": "Evaluate the agent's work. Run unit tests using run_shell_command if code was changed, and fix any lint errors."
          }
        ]
      }
    ],

    "SessionEnd": [
      {
        "name": "Clean Up Cache",
        "type": "command",
        "command": "rm -rf .gemini/tmp/*"
      }
    ]
  }
}
```

---

## Detailed Hook Explanations

### 1. Session Lifecycle Hooks
* **`SessionStart`**: Fires once when the CLI starts a session (either a fresh session or resuming a saved conversation). Ideal for environment setup.
* **`SessionEnd`**: Fires when the CLI session terminates. Useful for clearing caches, writing final logs, or committing workspace status.

### 2. Agent Lifecycle Hooks
* **`BeforeAgent`**: Triggers after the user sends a message but before the agent performs any planning. Used for input validation.
* **`AfterAgent`**: Triggers when the agent completes its run loop and returns control to the user. Good for automated QA or notification systems.

### 3. Tool Lifecycle Hooks
* **`BeforeToolSelection`**: Runs before the agent selects tools. Allows hooks to filter or dynamically restrict list of available tools.
* **`BeforeTool`**: Runs immediately before a specific tool executes. Requires a `matcher` string indicating which tool to intercept (or `*` for all tools). Can halt tool execution.
* **`AfterTool`**: Runs after a tool executes, receiving the tool's stdout/stderr payload. Useful for logging or running secondary validation commands like linters.

### 4. Model Lifecycle Hooks
* **`BeforeModel`**: Runs right before a raw text prompt payload is sent to the LLM.
* **`AfterModel`**: Runs right after receiving a raw text response back from the LLM. Typically used to censor/filter sensitive keywords or outputs.
