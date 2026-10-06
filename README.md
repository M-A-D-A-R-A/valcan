# Vulcan

> **A self-healing, self-evolving forge for coding agents.**

Vulcan is a harness designed to make coding agents **learn from their execution, recover from failures, and improve themselves over time**.

Inspired by **Vulcan**, the Roman god of fire and the forge, the project treats an agent harness like something that can be continuously **forged, tempered, repaired, and strengthened**.

The goal is not to build another coding agent.

The goal is to build the **system around the agent that makes the agent better.**

---

## Philosophy

Coding agents are already capable of writing code.

The harder problem is making them:

* remember what worked
* learn from what failed
* recover from mistakes
* select the right context
* choose the right model
* avoid repeating failed actions
* verify their own work
* improve their own workflows

Vulcan provides the feedback loop required for this.

```text
                    ┌─────────────────┐
                    │      HUMAN      │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │       PI        │
                    │  Coding Agent   │
                    └────────┬────────┘
                             │
                        execution
                             │
                             ▼
                    ┌─────────────────┐
                    │      VIGIL      │
                    │  Observe / Log  │
                    └────────┬────────┘
                             │
                        experiences
                             │
                             ▼
                    ┌─────────────────┐
                    │     VULCAN      │
                    │      Forge      │
                    ├─────────────────┤
                    │ Learn           │
                    │ Diagnose        │
                    │ Recover         │
                    │ Experiment      │
                    │ Evaluate        │
                    │ Evolve          │
                    └────────┬────────┘
                             │
                      improved harness
                             │
                             └──────────► PI
```

### The core loop

```text
Observe
   ↓
Understand
   ↓
Recover
   ↓
Learn
   ↓
Experiment
   ↓
Evaluate
   ↓
Forge
   ↓
Repeat
```

---

# Architecture

Vulcan is intentionally separated into distinct responsibilities.

### Pi — Execution

[Pi](https://github.com/badlogic/pi-mono) is the coding-agent execution engine.

It handles:

* reasoning
* tool calls
* file modifications
* shell commands
* tests
* interactive coding sessions

Vulcan does **not** replace Pi.

It works around Pi.

---

### Vigil — Observation

Vigil is the observability and experience layer.

It records what happened during agent execution:

* sessions
* prompts
* models
* context
* tool calls
* tool results
* failures
* retries
* files changed
* tests
* commands
* git state
* latency
* token usage
* behavioral signals

```text
Pi
 │
 ├── tool call
 ├── tool result
 ├── file edit
 ├── test
 └── error
       │
       ▼
     Vigil
```

Vigil provides the **evidence**.

---

### Vulcan — Evolution

Vulcan consumes that evidence and turns it into improvements.

```text
Vigil
  │
  ▼
Failure Mining
  │
  ▼
Pattern Detection
  │
  ▼
Hypothesis
  │
  ▼
Harness Modification
  │
  ▼
Behavioral Evaluation
  │
  ▼
Historical Replay
  │
  ├── Pass ──► Keep
  │
  └── Fail ──► Revert
```

Vigil watches.

**Vulcan forges.**

Pi executes.

---

# What Vulcan Controls

Vulcan is intended to become the control layer around coding agents.

## Context

Determine what the agent actually needs to know.

Instead of replaying an entire conversation:

```text
Task
+
Current State
+
Relevant Files
+
Relevant Experiences
+
Known Failures
+
Constraints
```

becomes the model context.

---

## Memory

Vulcan maintains structured memory rather than treating the conversation transcript as permanent memory.

```text
memory/
├── raw/
│   └── sessions/
│
├── experiences/
│   ├── successes/
│   └── failures/
│
├── strategies/
│
├── patterns/
│
├── experiments/
│
└── evaluations/
```

### Raw sessions

Permanent execution evidence.

### Experiences

Compressed lessons extracted from previous sessions.

### Strategies

Reusable approaches that worked across multiple tasks.

### Patterns

Recurring behavioral patterns.

### Experiments

Changes Vulcan has attempted.

### Evaluations

Evidence showing whether an experiment improved the system.

---

# Model Routing

Vulcan should remain model-agnostic.

The harness decides which model is appropriate for a task.

Example:

```text
                    Task
                      │
                      ▼
                 Vulcan Router
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
       Qwen 8B     Qwen 14B    Cloud Model
       cheap       complex      reasoning
       evals       coding       architecture
```

A possible strategy:

```text
Qwen 8B
   │
   ├── success ───────────────► continue
   │
   └── repeated failure
              │
              ▼
          Qwen 14B
              │
              └── repeated failure
                         │
                         ▼
                    stronger model
```

Model selection should itself become something Vulcan can learn from historical results.

---

# Self-Healing

Self-healing comes before self-evolution.

Vulcan should detect when an agent is stuck and attempt recovery.

Examples:

### Tool failure

```text
Tool fails
   ↓
Inspect error
   ↓
Correct arguments
   ↓
Retry
```

### Test failure

```text
Test fails
   ↓
Inspect failure
   ↓
Identify likely cause
   ↓
Modify code
   ↓
Run test again
```

### Repeated failure

```text
Failure
   ↓
Retry
   ↓
Failure
   ↓
Retry with modified strategy
   ↓
Failure
   ↓
Change model / escalate
```

### Context failure

```text
Context too large
       ↓
Compact
       ↓
Retrieve relevant memory
       ↓
Reconstruct context
       ↓
Continue
```

The goal is to avoid:

> **Doing the same thing repeatedly and expecting a different result.**

---

# Self-Evolution

Self-evolution is different from self-healing.

Self-healing fixes the **current task**.

Self-evolution improves the **harness itself**.

```text
Historical Sessions
        │
        ▼
Failure Mining
        │
        ▼
Recurring Pattern
        │
        ▼
Hypothesis
        │
        ▼
Modify Harness
        │
        ▼
Run Evaluation
        │
        ▼
Compare Against Baseline
        │
    ┌───┴───┐
    ▼       ▼
  Better   Worse
    │       │
    ▼       ▼
  Keep    Revert
```

Vulcan should never randomly modify itself.

Every evolution should be an **experiment with evidence**.

---

# Evolution Experiments

Every proposed change gets an experiment ID.

Example:

```yaml
experiment_id: EXP-0017

hypothesis:
  "The agent frequently edits source files without
   running the relevant validator."

target:
  component: verification_policy

baseline:
  task_success_rate: 0.71
  verification_rate: 0.54
  average_tokens: 12400

candidate:
  task_success_rate: 0.79
  verification_rate: 0.91
  average_tokens: 13100

regression:
  passed: true

decision:
  keep: true
```

The experiment becomes part of Vulcan's own history.

---

# Behavioral Evaluation

End-to-end success is not enough.

Vulcan should evaluate **individual agent behaviors**.

Examples:

* Did the agent verify a source-code change?
* Did it run the appropriate test?
* Did it choose the correct tool?
* Did it preserve evidence after failure?
* Did it avoid repeating a failed tool call?
* Did it ask for clarification when required?
* Did it recover from a failed command?
* Did it select an appropriate model?
* Did it use relevant context?
* Did it respect tool and sandbox policies?

This allows Vulcan to answer:

> **What behavior actually changed?**

rather than simply:

> **Did the final task pass?**

---

# Historical Replay

Before adopting an evolution, Vulcan should replay previous tasks.

```text
                Historical Tasks
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
        Easy        Medium        Hard
          │            │            │
          └────────────┼────────────┘
                       ▼
                  Candidate
                       │
                       ▼
                  Evaluation
                       │
             ┌─────────┴─────────┐
             ▼                   ▼
         Improvement          Regression
             │                   │
             ▼                   ▼
           Keep                Revert
```

The historical dataset should contain:

* successful tasks
* failed tasks
* repeated failures
* tool failures
* context failures
* recovery scenarios
* difficult coding tasks

This becomes Vulcan's **forge test suite**.

---

# Git Safety

Vulcan must never directly rewrite its production branch.

Every evolution happens in an isolated branch.

```text
main
 │
 └── evolution/EXP-0017
          │
          ├── modify
          ├── test
          ├── evaluate
          ├── replay
          │
          ├── better ──► merge
          │
          └── worse ───► delete / revert
```

Evolution must always have:

* a baseline
* a hypothesis
* a measurable evaluation
* regression protection
* rollback capability

The system should **never be able to disable its own safety mechanisms**.

---

# Observability

Every meaningful action should be traceable.

A common event envelope:

```json
{
  "event_id": "...",
  "session_id": "...",
  "trace_id": "...",
  "parent_id": "...",
  "timestamp": "...",
  "project": "...",
  "agent": "pi",
  "model": "qwen3-14b",
  "event_type": "tool_call",
  "duration_ms": 123,
  "input_tokens": 100,
  "output_tokens": 50,
  "success": true,
  "payload": {}
}
```

A complete session should be reconstructable as:

```text
User Prompt
    ↓
Model Selected
    ↓
Context Built
    ↓
Reasoning
    ↓
Tool Call
    ↓
Tool Result
    ↓
File Modification
    ↓
Test
    ↓
Failure / Success
    ↓
Recovery
    ↓
Git Diff
    ↓
Completion
```

This trace is the raw material from which Vulcan learns.

---

# Raspberry Pi Deployment

Vigil can run as a lightweight always-on observability server.

```text
                    Mac
        ┌─────────────────────────┐
        │                         │
        │ Pi Coding Agent         │
        │ Qwen 8B / 14B           │
        │ Cloud Models             │
        │                         │
        │ Vulcan                  │
        │ Evolution Engine        │
        │                         │
        └────────────┬────────────┘
                     │
                     │ HTTP / OTLP
                     ▼
        ┌─────────────────────────┐
        │     Raspberry Pi        │
        │                         │
        │        Vigil            │
        │     Go + SQLite         │
        │                         │
        │  Persistent observations │
        └─────────────────────────┘
```

The Raspberry Pi acts as the **black box recorder**.

The Mac performs expensive analysis and evolution.

---

# Privacy

Agent traces can contain sensitive information.

Vulcan should remain local-first.

Sensitive data should be redacted before persistence where appropriate:

* API keys
* access tokens
* passwords
* cookies
* private keys
* authorization headers
* `.env` values
* credentials embedded in commands

Raw traces should never leave the local environment unless explicitly exported.

---

# Research Inspiration

Vulcan draws inspiration from several emerging ideas around agent harness engineering:

* **Harness Engineering** — treating the harness itself as production software.
* **Behavioral evaluations** — evaluating individual agent behaviors rather than only final task success.
* **ReasoningBank** — extracting reusable strategies from successful and failed experiences.
* **Self-evolving agent systems** — using observed execution to generate and test improvements.
* **NVIDIA SoL-Pi** — improving Pi through extensions and systematic experimentation rather than modifying the underlying agent.
* **Sandboxing and policy layers** — keeping long-running/self-modifying agents constrained and reversible.

Vulcan is intended to combine these ideas into a **local, model-agnostic, experimentally driven harness**.

---

# Roadmap

## Phase 1 — Observe

Build the complete execution trace.

* [ ] Pi → Vigil event integration
* [ ] Session tracking
* [ ] Tool-call tracking
* [ ] Model tracking
* [ ] Token/latency tracking
* [ ] File-change tracking
* [ ] Test tracking
* [ ] Git tracking
* [ ] Failure classification

---

## Phase 2 — Understand

Turn traces into useful experiences.

* [ ] Failure mining
* [ ] Success extraction
* [ ] Experience generation
* [ ] Pattern detection
* [ ] Strategy extraction
* [ ] Cross-project memory
* [ ] Context retrieval

---

## Phase 3 — Evaluate

Build the forge's testing infrastructure.

* [ ] Behavioral evaluations
* [ ] Regression suite
* [ ] Historical replay
* [ ] Baseline metrics
* [ ] Candidate comparison
* [ ] Experiment tracking

---

## Phase 4 — Self-Heal

Teach Vulcan to recover during execution.

* [ ] Tool recovery
* [ ] Test recovery
* [ ] Context recovery
* [ ] Retry strategies
* [ ] Model escalation
* [ ] Failure-loop detection
* [ ] Human escalation

---

## Phase 5 — Self-Evolve

Allow Vulcan to modify its own harness.

* [ ] Failure → hypothesis pipeline
* [ ] Automated branch creation
* [ ] Code modification
* [ ] Behavioral evaluation
* [ ] Historical replay
* [ ] Automatic comparison
* [ ] Rollback
* [ ] Experiment ledger

Initially:

```text
Vulcan proposes
      ↓
Human approves
      ↓
Vulcan executes
```

Later:

```text
Vulcan proposes
      ↓
Vulcan evaluates
      ↓
Safety gates
      ↓
Automatic adoption
```

---

# Initial Evolution Experiments

The first experiments should be small and measurable.

### EXP-001 — Mandatory Verification

Detect whether agents modify source code without running relevant validation.

### EXP-002 — Failure Recovery

Improve recovery after failed commands and tool calls.

### EXP-003 — Context Selection

Determine whether retrieving targeted context performs better than sending larger transcripts.

### EXP-004 — Model Routing

Compare model selection strategies across task types.

### EXP-005 — Automatic Escalation

Escalate from smaller/local models after repeated failures.

### EXP-006 — Tool Repair

Learn common tool-call failures and automatically correct them.

---

# Design Principles

### 1. Evidence over intuition

Every improvement should have measurable evidence.

### 2. Small changes

Prefer incremental evolution over large uncontrolled rewrites.

### 3. Reversible by default

Every evolution must be easy to undo.

### 4. Model agnostic

Vulcan should not depend on one model provider.

### 5. Local first

Execution history and memory should remain under the user's control.

### 6. Observe before modifying

The system needs evidence before it attempts to improve itself.

### 7. Behavior over vibes

Measure concrete behaviors.

### 8. Failure is training data

A failed execution is not just an error.

It is an opportunity to discover how the system can become better.

### 9. The harness is software

The harness itself deserves tests, benchmarks, observability, versioning, and rollback.

### 10. Never remove the anvil

The system that evaluates evolution must itself remain protected from the evolution process.

---

# Long-Term Vision

The long-term goal is a coding environment where the agent does not remain static.

Instead:

```text
        ┌──────────────────────────┐
        │      Coding Agent        │
        └────────────┬─────────────┘
                     │
                  executes
                     │
                     ▼
        ┌──────────────────────────┐
        │         Vigil            │
        │     records reality      │
        └────────────┬─────────────┘
                     │
                  evidence
                     │
                     ▼
        ┌──────────────────────────┐
        │         Vulcan           │
        │                          │
        │    Learn → Experiment    │
        │       → Evaluate         │
        │          → Forge         │
        └────────────┬─────────────┘
                     │
                  improvement
                     │
                     ▼
              Better Harness
                     │
                     └──────► Better Agent
```

Every coding session becomes another source of evidence.

Every failure becomes a potential experiment.

Every successful experiment becomes a new capability.

And every accepted improvement makes the next coding session better.

---

# Definition of Success

Vulcan succeeds when it can demonstrate, with historical evidence, that:

* coding tasks become more reliable
* failures are recovered from automatically
* repeated mistakes decrease
* context becomes more efficient
* model selection improves
* unnecessary tokens decrease
* human intervention decreases
* the harness can safely improve itself
* every improvement remains measurable and reversible

The ultimate metric is not:

> **How intelligent is the model?**

It is:

> **How much better does the system become from experience?**

---

## Status

🚧 **Early development**

Current stack:

```text
Pi
├── Qwen3-8B
├── Qwen3-14B
└── Cloud Models

Vigil
├── Go
├── SQLite
├── FTS5
└── React / Vite

Vulcan
├── Context
├── Memory
├── Routing
├── Recovery
├── Evaluation
└── Evolution
```

---

## Name

**Vulcan** is inspired by the Roman god of fire and the forge.

The forge is where raw material is shaped, tempered, repaired, and strengthened.

That is the philosophy behind this project:

> **Forge a better harness.**
