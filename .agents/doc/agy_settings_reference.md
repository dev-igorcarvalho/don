# Antigravity CLI (`agy`) Configuration Reference

This document provides a comprehensive reference for configuring the Antigravity CLI (`agy`).

## Configuration File Locations
* **Global Configuration:** [~/.gemini/antigravity-cli/settings.json](file:///home/igor/.gemini/antigravity-cli/settings.json)
* **Workspace / Project Configuration:** [.agents/settings.json](file:///home/igor/Documents/projetos/don/.agents/settings.json)

---

## All Configurable Options (`settings.json`)

Below is a complete `settings.json` template showing all available configurations with representative values:

```json
{
  "allowNonWorkspaceAccess": false,
  "altScreenMode": "default",
  "artifactReviewPolicy": "asks-for-review",
  "colorScheme": "terminal",
  "editor": "auto",
  "enableTelemetry": true,
  "enableTerminalSandbox": false,
  "gcp": {
    "projectId": "your-gcp-project-id",
    "location": "us-central1",
    "zone": "us-central1-a"
  },
  "historySize": 2000,
  "model": "gemini-3.5-flash",
  "notifications": false,
  "permissions": {
    "allow": [
      "Read",
      "Write",
      "Bash(git *)"
    ],
    "deny": [
      "Bash(rm -rf *)"
    ],
    "ask": [
      "Bash"
    ]
  },
  "runningLightSpeed": "medium",
  "showFeedbackSurvey": true,
  "showTips": true,
  "statusLine": {
    "type": "command",
    "command": "/home/user/.gemini/antigravity-cli/statusline.sh",
    "padding": 0
  },
  "title": {
    "type": "command",
    "command": "/home/user/.gemini/antigravity-cli/title.sh"
  },
  "toolPermission": "request-review",
  "trustedWorkspaces": [
    "/home/igor/Documents/projetos/don"
  ],
  "useG1Credits": false,
  "verbosity": "high",
  "hooks": {
    "BeforeAgent": [],
    "AfterAgent": [
      {
        "matcher": "*",
        "hooks": [
          {
            "name": "Linter",
            "type": "command",
            "command": ".gemini/lint.sh"
          }
        ]
      }
    ]
  }
}
```

---

## Configuration Reference Table

| JSON Key | Type | Valid Values / Example | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `allowNonWorkspaceAccess` | `bool` | `true`, `false` | `false` | Permits the agent to read or write files outside the workspace root. |
| `altScreenMode` | `string` | `"default"`, `"always"`, `"never"` | `"default"` | Controls alternate screen buffer usage in the terminal. |
| `artifactReviewPolicy` | `string` | `"always-proceed"`, `"agent-decides"`, `"asks-for-review"` | `"asks-for-review"` | Controls when the agent asks for artifact review. |
| `colorScheme` | `string` | `"terminal"`, `"dark"`, `"light"`, `"solarized dark"`, `"tokyo night"` | `"terminal"` | The CLI color scheme (`terminal` inherits terminal's color theme). |
| `editor` | `string` | `"auto"`, `"vim"`, `"nano"`, `"code"` | `"auto"` | Preferred editor command for file opening. |
| `enableTelemetry` | `bool` | `true`, `false` | `true` | Toggles anonymous usage and crash reporting. |
| `enableTerminalSandbox` | `bool` | `true`, `false` | `false` | Runs terminal commands inside a restricted sandbox environment. |
| `gcp` | `object` | See [GCP Configuration](#gcp-configuration) | `null` | GCP project, location, and zone configurations. |
| `historySize` | `int` | `2000`, `-1` (unlimited) | `2000` | Maximum number of history entries persisted to disk. |
| `model` | `string` | `"gemini-3.5-flash"`, `"gemini-3.5-pro"` | `"gemini-3.5-flash"` | The active model identifier used by the main agent. |
| `notifications` | `bool` | `true`, `false` | `false` | Enables system notifications on task completion. |
| `permissions` | `object` | See [Permissions Rules](#permissions-rules) | `null` | Global allow/deny/ask rules for tools. |
| `runningLightSpeed` | `string` | `"off"`, `"fast"`, `"medium"`, `"slow"` | `"medium"` | Artificial typing delays to control thought visualization. |
| `showFeedbackSurvey` | `bool` | `true`, `false` | `true` | Displays periodic product feedback surveys. |
| `showTips` | `bool` | `true`, `false` | `true` | Displays helpful tips and shortcuts in the CLI. |
| `statusLine` | `object` | See [Status Line](#status-line-customization) | `null` | Configuration for routing TUI status bar info to a script. |
| `title` | `object` | See [Title Customization](#title-customization) | `null` | Configuration for routing terminal tab title to a script. |
| `toolPermission` | `string` | `"always-proceed"`, `"request-review"`, `"strict"`, `"proceed-in-sandbox"` | `"request-review"` | Confirmation mode for tools. |
| `trustedWorkspaces` | `array` | `["/home/user/my-project"]` | `[]` | List of directory paths trusted by the user for execution. |
| `useG1Credits` | `bool` | `true`, `false` | `false` | Toggles usage of Google One AI premium quotas. |
| `verbosity` | `string` | `"high"`, `"low"` | `"high"` | Detail level of agent trace rendering. |
| `hooks` | `object` | See [Lifecycle Hooks](#lifecycle-hooks) | `null` | Event hooks to run custom scripts during agent cycles. |

---

## Detailed Section Explanations

### GCP Configuration
When routing queries or utilizing GCP resources, define the target environment:
```json
"gcp": {
  "projectId": "my-enterprise-project",
  "location": "us-central1",
  "zone": "us-central1-a"
}
```

### Permissions Rules
Rules use format `ToolName` or `ToolName(pattern)`. Evaluation order is `deny` > `ask` > `allow`.
```json
"permissions": {
  "allow": ["Read", "Write", "Bash(git *)"],
  "deny": ["Bash(rm *)"],
  "ask": ["Bash"]
}
```

### Status Line Customization
Route the TUI status bar state (delivered as JSON on `stdin`) into an external script:
```json
"statusLine": {
  "type": "command",
  "command": "/home/user/.gemini/antigravity-cli/statusline.sh",
  "padding": 0
}
```

### Title Customization
Update terminal tab titles dynamically:
```json
"title": {
  "type": "command",
  "command": "/home/user/.gemini/antigravity-cli/title.sh"
}
```

### Lifecycle Hooks
Define custom commands to trigger automatically before or after agent tasks:
```json
"hooks": {
  "BeforeAgent": [],
  "AfterAgent": [
    {
      "matcher": "*",
      "hooks": [
        {
          "name": "Linter",
          "type": "command",
          "command": ".gemini/lint.sh"
        }
      ]
    }
  ]
}
```
