# Pipeline Workflow Rules

Before reading, writing, executing, or reasoning about any pipeline YAML file in this project, read `.artfactory/rules/workflow.md`
using the Read tool.

That file is the authoritative instruction manual for the pipeline DSL used here. It defines the meaning and behavior of
every YAML tag.

## When This Rule Applies

- Creating or editing a pipeline YAML file
- Interpreting or explaining a pipeline YAML file
- Debugging a pipeline execution
- Answering questions about how `pipeline`, `workflow`, `agent`, `loop`, or `depends` work

## Quick Reference (from `.artfactory/rules/workflow.md`)

| Tag        | Scope                 | Key behavior                                                                                          |
|------------|-----------------------|-------------------------------------------------------------------------------------------------------|
| `pipeline` | top-level             | holds `workflows`; workflows run in parallel by default                                               |
| `workflow` | inside `pipeline`     | holds `agents`; agents run in parallel by default; supports `loop`                                    |
| `loop`     | inside `workflow`     | requires `max_iterations`; stops when any agent outputs the `until` string                            |
| `agent`    | inside `workflow`     | task via `prompt`; optional `system` prepended to prompt; supports `artifact.output` for file handoff |
| `depends`  | `agent` or `workflow` | agent deps → same workflow only; workflow deps → same pipeline only                                   |

### Variable interpolation

| Expression                               | Resolves to                                                                      |
|------------------------------------------|----------------------------------------------------------------------------------|
| `${name}`                                | The agent's own name                                                             |
| `${output.agentName}`                    | Text output of a completed agent in the same workflow                            |
| `${loop.first}`                          | `true` on the first loop iteration (use in `if` blocks)                          |
| `${loop.prev.agentName}`                 | Output of `agentName` from the previous loop iteration                           |
| `${agentName.artifact}`                  | Artifact written by `agentName` in the same workflow                             |
| `${workflowName.agentName.artifact}`     | Artifact written by `agentName` in a different workflow                          |
| `${pipeline.runDir}`                     | Isolated run directory: `.artfactory/<timestamp>-<uuid>-<pipeline_name>`         |

### Agent fields

| Field             | Meaning                                                                                          |
|-------------------|--------------------------------------------------------------------------------------------------|
| `system`          | Optional string prepended to the agent's prompt before execution; supports `${...}` interpolation |
| `artifact.output` | Path where the agent must write its result as a file                                             |

To consume another agent's artifact, reference it via `${agentName.artifact}` in the prompt — no `artifact.input` field needed.

### Error handling

- A failing agent stops its workflow; downstream `depends` agents are skipped.
- A failing workflow stops downstream `depends` workflows; independent parallel workflows are unaffected.

**Full semantics and examples (including the implement→verify→fix loop pattern) are in `.artfactory/rules/workflow.md` — always read that file first.**