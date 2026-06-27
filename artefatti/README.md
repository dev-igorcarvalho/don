<div align="center">
  <a href="https://github.com/dev-igorcarvalho/don">
    <img src="../docs/assets/artefatti_logo.png" alt="Don Logo" width="300" height="300" style="border-radius: 12px; border: 0px solid #f4f1ea; box-shadow: 0 8px 16px rgba(0,0,0,0.2); object-fit: cover; object-position: center;">
  </a>

  <h1>🌹 Artefatti </h1>

  <p>
    <b><i>"Laying the foundations of our operations, one artifact at a time."</i></b>
  </p>

  <p>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go Version"></a>
    <img src="https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-blue?style=flat-square" alt="Platform">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License">
  </p>
</div>

---

## 📖 Overview

**don-artefatti** is a lightweight, zero-dependency Node.js CLI tool that scaffolds the standard `.artefatti` workspace
structure required for orchestrating agentic pipelines in the **Don** monorepo. It automatically sets up directories,
copies baseline rules, and links rule files to developer tools like Claude Code (via `.claude/rules`).

---

## 🛠️ Features

- 🔋 **Zero Dependencies**: Relies solely on Node.js core APIs (`fs`, `path`, etc.) to guarantee reliability and speed.
- 📁 **Automated Structure**: Instantly provisions workspaces with standard execution folders (`agents/`, `artifacts/`,
  `pipelines/`, `rules/`).
- 📜 **Policy Distribution**: Injects baseline pipeline rule sets (`workflow.md`, `pipeline-workflow-rules.md`) to guide
  orchestrators.
- 🤖 **Developer Tool Hooking**: Automatically detects `.claude/` directories and propagates rules to
  `.claude/rules/pipeline-workflow-rules.md`.
- 🇮🇹 **Italian/Mafia-themed CLI**: The command-line output is customized with themed logging messages to fit the *Don*
  ecosystem.

---

## 📂 Generated Structure

When initialized, the tool creates the following directory structure in the target project folder:

```text
.
├── .artefatti/
│   ├── agents/                   # Staging/definition area for agents
│   ├── artifacts/                # Generated reports, execution outcomes, and data
│   ├── pipelines/                # Workflows and pipeline task scripts
│   └── rules/                    # Project rules and guidelines
│       ├── pipeline-workflow-rules.md
│       └── workflow.md
```

Additionally, if a `.claude/` directory is present in the target workspace, the tool writes rules to:

- `.claude/rules/pipeline-workflow-rules.md`

---

## 🚀 Installation & Usage

### 1. Linking Locally

To install and run `don-artefatti` globally from the monorepo root:

```bash
cd artefatti
npm link
```

### 2. Scaffold a Workspace

Navigate to your desired project directory and run:

```bash
don-artefatti
```

Alternatively, run the script directly with Node:

```bash
node /path/to/don/artefatti/bin/cli.js
```

### 3. Example CLI Run Output

```text
🇮🇹 Inizializzazione della struttura in corso...
Created folder: .artefatti
Created folder: .artefatti/agents
Created folder: .artefatti/artifacts
Created folder: .artefatti/pipelines
Created folder: .artefatti/rules
Copied file: .artefatti/README.md
Copied file: .artefatti/rules/pipeline-workflow-rules.md
Copied file: .artefatti/rules/workflow.md

🌹 Salutiamo il nostro nuovo amico!
La struttura `.artefatti` è stata creata con successo nel tuo territorio.
Abbiamo messo le regole al loro posto. Ricorda: una promessa fatta è un debito pagato.
Fai buon uso di questo potere... prima che ti facciamo un'offerta que non potrai rifiutare.
Benvenuto nella Famiglia.
```

---

## 📋 Pipeline Workflow Configuration Examples

The pipelines run by the orchestrator are configured using YAML files. Below are two examples showing how to define
pipelines, workflows, parallel agent execution, dependency resolution, output forwarding, and loop iterators.

### Example 1: Parallel Execution & Dependency Resolution

In this example:

- `A` and `B` run in parallel.
- `C` waits for both `A` and `B` to finish before running.
- The `greet_de` workflow runs only after the `greet_ab` workflow successfully finishes.
- Agents can read output artifacts from other workflows and agents via `${C.artifact}` or `${greet_ab.C.artifact}`.

```yaml
pipeline:
  name: greet-demo
  workflows:
    - name: greet_ab
      agents:
        - name: A
          prompt: "Print: I am agent ${name}. Hello world."
        - name: B
          prompt: "Print: I am agent ${name}. Hello world."
        - name: C
          depends: [ A, B ]
          prompt: "Gather the output of ${output.A} and ${output.B} and greet them both."
          artifact:
            output: ${pipeline.runDir}/greet-ab-C.md

    - name: greet_de
      depends: [ greet_ab ]
      agents:
        - name: D
          prompt: "Print: I am agent ${name}. Hello world. \n I can also read another agent's artifact: ${C.artifact}"
        - name: E
          prompt: "Print: I am agent ${name}. Hello world. Also i can get another workflow agent artifact with ${greet_ab.C.artifact}"
        - name: F
          depends: [ D, E ]
          prompt: "Gather the output of ${output.D} and ${output.E} and greet them both."
          artifact:
            output: ${pipeline.runDir}/greet-de-F.md
```

### Example 2: Implement → Verify → Fix Loop (Iterative Execution)

In this example:

- The `review_loop` workflow repeats up to 5 times.
- It continues running until Agent `B` outputs the stop signal `APPROVED`.
- In the first iteration, Agent `A` writes code from scratch.
- In subsequent iterations, Agent `A` receives its own previous output (`${loop.prev.A}`) and the reviewer's feedback (
  `${loop.prev.B}`) to correct the code.

```yaml
pipeline:
  workflows:
    - name: review_loop
      loop:
        max_iterations: 5
        until: "APPROVED"
      agents:
        - name: A
          prompt: |
            ${if loop.first}
            Write a Go function that reverses a string.
            ${else}
            Previous implementation:
            ${loop.prev.A}

            Reviewer feedback:
            ${loop.prev.B}

            Fix every issue and return the corrected code.
            ${endif}

        - name: B
          depends: [ A ]
          prompt: |
            Review this Go code:
            ${output.A}

            If it is correct and complete, respond exactly: APPROVED
            Otherwise, list every issue found.
```
