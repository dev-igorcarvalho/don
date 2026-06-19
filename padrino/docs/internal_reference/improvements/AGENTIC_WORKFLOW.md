# Agentic Coding Workflow: Plan & Roadmap

This document outlines the design, implementation strategy, and roadmap for a sequential multi-agent pipeline within the Gemini CLI environment.

## 1. Goal
To automate high-quality code delivery by orchestrating a specialized sequence of "agents" (sub-agents) that handle different phases of the software development lifecycle (SDLC).

## 2. Architecture
The workflow uses a **Manager-Worker** pattern:
- **Orchestrator (Main Agent):** Managed via a Gemini CLI **Skill**. It triggers on coding requests and manages the state machine.
- **Workers (Sub-agents):** Isolated LLM instances defined in `.gemini/agents/`. Each worker has a specific system prompt and toolset to prevent context bias.

### The Pipeline Sequence
1.  **Planner:** Analyzes requirements and drafts a technical strategy.
2. 
3. **Developer:** Implements the core logic/features.
4. **Refactor:** Optimizes the code (Martin Fowler style).
5. **Unit Tester:** Writes and validates tests.
6. **Reviewer:** Conducts a final quality/security check. and Create a report
7. **TechWriter** Conducts documentations, go-doc, etc
8. **IssueManager:** consumes Reviewer report and create issues / tasks documentation.

---

## 3. Roadmap

### Phase 1: Scaffolding (Infrastructure)
*   [ ] Create `.gemini/agents/` directory.
*   [ ] Create "Empty" agent definitions (`.md` files with frontmatter) for all 6 personas.
*   [ ] Initialize the `coding-pipeline` skill structure.

### Phase 2: Orchestration (The Skill)
*   [ ] Define the state machine logic in `SKILL.md`.
*   [ ] Implement sequential `invoke_agent` calls.
*   [ ] Package and install the skill locally (`.skill` file).

### Phase 3: Persona Refinement (Prompt Engineering)
*   [ ] Define the "Developer" standards.
*   [ ] Define the "Refactor" principles (Martin Fowler reference).
*   [ ] Define the "Reviewer" checklist (Go standards, security).
*   [ ] Refine the "IssueManager" output format.

### Phase 4: Automation & Loopback
*   [ ] Implement basic error handling (e.g., if Unit Tester fails, loop back to Developer).
*   [ ] Enable "silent" mode for agents to reduce terminal noise.
*   [ ] Finalize the "trigger" mechanism so it activates seamlessly on coding requests.

---

## 4. Implementation Details

### Sub-Agent Template
Each agent in `.gemini/agents/` follows this format:
```markdown
---
name: agent-name
description: Brief description for the orchestrator.
tools: [read_file, write_file, grep_search, run_shell_command]
---
# System Prompt
You are the [Role Name]. Your specific mission is to...
```

### Orchestrator Logic (Skill)
The `SKILL.md` will contain the following procedural guidance:
1.  **Receive** user coding request.
2.  **Invoke `planner`** to generate a roadmap.
3.  **Invoke `developer`** with the roadmap.
4.  **Invoke `refactor`** on modified files.
5.  **Invoke `unit-tester`** to ensure 100% coverage.
6.  **Invoke `reviewer`** for final approval.
7.  **Invoke `issue-manager`** to wrap up.

---

## 5. Next Steps
1.  Review this roadmap.
2.  Execute Phase 1 by creating the `.gemini/agents` directory and scaffolding the files.
