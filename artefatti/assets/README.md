# .artefatti

Home of the agentic pipeline orchestration system. Contains the DSL spec, reusable agent prompts, pipeline definitions, and all run artifacts.

---

## Structure

```
.artefatti/
├── workflow.md                  # DSL spec — read this before writing any pipeline
├── install-artifactory.sh       # Installs rules into .claude/rules/
├── agents/                      # Reusable agent prompt files (referenced via file://)
│   └── agent.example.md
├── artifacts/                   # Output files from pipeline runs (auto-generated)
│   └── <timestamp>-<uuid>-<pipeline_name>/
│       └── *.md
├── pipelines/                   # Pipeline YAML definitions
│   └── pipeline.example.yaml
└── rules/                       # Claude rule files — source of truth
    └── pipeline-workflow-rules.md
```

---

## How it works

You write a pipeline YAML and ask Claude to run it. Claude reads `workflow.md` as the orchestrator spec, then executes the pipeline — spawning agents, resolving dependencies, and writing artifacts.

### 1. Define a pipeline

Create a YAML file under `pipelines/`. See `pipelines/pipeline.example.yaml` for a working reference.

```yaml
pipeline:
  name: my-pipeline
  workflows:
    - name: analyse
      agents:
        - name: reader
          system: "You are a senior Go engineer. Be concise."
          prompt: "Summarise the main concerns in this file: file://./path/to/file.go"
          artifact:
            output: ${pipeline.runDir}/summary.md
```

### 2. Reuse agent prompts

Long or reused prompts can live in `agents/` and be referenced by path:

```yaml
prompt: "file://.artefatti/agents/my-prompt.md"
```

### 3. Run it

Tell Claude to run the pipeline file. It will read `workflow.md` first, then orchestrate the agents.

### 4. Find the output

Every run writes artifacts to an isolated directory:

```
.artefatti/artifacts/<timestamp>-<uuid>-<pipeline_name>/
```

Each run gets its own directory, so past runs are never overwritten.

---

## DSL quick reference

| Concept            | How                                                                 |
|--------------------|---------------------------------------------------------------------|
| Parallel agents    | Default — agents in a workflow run concurrently                     |
| Sequential agents  | `depends: [agentName]` on the agent that should run after          |
| Parallel workflows | Default — workflows in a pipeline run concurrently                  |
| Sequential workflows | `depends: [workflowName]` on the workflow that should run after   |
| Agent persona      | `system:` field — prepended to the prompt before execution          |
| Write a file       | `artifact.output: ${pipeline.runDir}/file.md`                       |
| Read another agent's file | `${agentName.artifact}` in the prompt                      |
| Loop until done    | `loop: { max_iterations: N, until: "STOP_SIGNAL" }` on a workflow  |

Full spec: [`workflow.md`](rules/workflow.md)

---
