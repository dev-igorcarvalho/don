# You are the agentic orchestrator.

## Composition Blocks

### `pipeline`
A pipeline is the top-level container. It holds one or more `workflow` blocks.
Workflows inside a pipeline run **in parallel by default** unless a workflow declares `depends` on another workflow.

### `workflow`
A named unit of work containing one or more `agent` blocks.
- Workflow can reference their own name via the field name.
- Agents inside a workflow run **in parallel by default**.
- Use `depends: [name1, name2]` to make an agent wait for specific agents to complete first.
- Use `depends: [name1, name2]` at the workflow level to make a workflow wait for other workflows.
- Use `loop` to repeat the workflow's agents until a stop signal is produced or `max_iterations` is reached.

### `loop`
Optional block on a workflow that enables iterative execution.
- `max_iterations` — hard cap on the number of loop cycles (required).
- `until` — stop signal string; the loop ends when any agent's output contains this exact value.
- Inside prompts, use `${loop.first}` / `${loop.prev.agentName}` to branch on first vs subsequent iterations.

### `agent`
A single agent that executes one task.
- Each agent receives its task via the `prompt` field. This field could be both a string of a file path to be read (file://./prompts/my-prompt.md)
- Agents can optionally declare a `system` field — a plain string prepended to the agent's prompt before execution. Use it to set a persona, restrict scope, or inject shared context without repeating it in every prompt. Variable interpolation (`${...}`) is supported in `system` the same as in `prompt`.
- Agents can reference their own name via `${name}`.
- Agents can reference another agent's output via `${agentName.output}`.
- Agents can reference another agent's output artifact via `${agentName.artifact}` in the prompt (same workflow).
- Agents can reference an artifact from a different workflow via `${workflowName.agentName.artifact}` in the prompt.
- Agents can write their result to a file via the `artifact.output` field, which specifies the destination path.
- An agent with `depends` will only start after all listed agents have completed successfully.

### Pipeline Run Directory

Every pipeline execution creates an isolated run directory at:

```
.artefatti/artifacts/<timestamp>-<uuid>-<pipeline_name>/
```

- Created automatically before any agent runs, once per pipeline execution.
- All workflows and agents in the pipeline share this directory.
- Use `${pipeline.runDir}` in `artifact.output` paths and prompts to reference this directory.
- Relative paths in `artifact.output` are resolved relative to `${pipeline.runDir}`, not the project root.
- The `<pipeline_name>` segment is the value of the top-level `name` field on the `pipeline` block (kebab-cased). If the pipeline has no `name`, it defaults to `pipeline`.

Example:
```yaml
artifact:
  output: ${pipeline.runDir}/my-file.md
```

This resolves to something like:
```
.artefatti/artifacts/20260619T143012-1de9fe3d-57db-4b40-bfe5-feb1b5239d44-my-pipeline/my-file.md
```

### Dependencies
- Syntax: `depends: [agentOrWorkflowName1, agentOrWorkflowName2]`
- Names must match the `name` of another agent or workflow within the same scope.
- Agent dependencies must refer to agents within the same workflow.
- Workflow dependencies must refer to workflows within the same pipeline.

### Error Handling
- If an agent fails, its workflow stops and downstream agents that depend on it are skipped.
- If a workflow fails, downstream workflows that depend on it are skipped.
- Independent parallel agents/workflows are not affected by a sibling's failure.

---

## Example

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
          depends: [A, B]
          prompt: "Gather the output of ${output.A} and ${output.B} and greet them both."
          artifact:
            output: ${pipeline.runDir}/greet-ab-C.md

    - name: greet_de
      depends: [greet_ab]
      agents:
        - name: D
          prompt: "Print: I am agent ${name}. Hello world. \n I can also read another agent's artifact: ${C.artifact}"
        - name: E
          prompt: "Print: I am agent ${name}. Hello world. Also i can get another workflow agent artifact with ${greet_ab.C.artifact}"
        - name: F
          depends: [D, E]
          prompt: "Gather the output of ${output.D} and ${output.E} and greet them both."
          artifact:
            output: ${pipeline.runDir}/greet-de-F.md
```
---

## Example 2: Implement → Verify → Fix loop

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
          depends: [A]
          prompt: |
            Review this Go code:
            ${output.A}

            If it is correct and complete, respond exactly: APPROVED
            Otherwise, list every issue found.
```

**How it works:**
- Iteration 1 — A implements from scratch; B reviews.
- Iteration 2+ — A receives its own previous output and B's feedback, fixes issues; B reviews again.
- Loop stops as soon as B outputs `APPROVED`, or after 5 iterations at most.

