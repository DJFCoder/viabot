---
title: Multi-Agent TDD Pipeline
tags: [spec, agents, pipeline, tdd]
type: spec
created: 2026-05-16
updated: 2026-05-16
---

# Multi-Agent TDD Pipeline

> **Status:** Implemented
> **Default Agent:** `viabot-specialist`

## Overview

The ViaBot project uses a **multi-agent TDD enforcement pipeline** built on OpenCode's custom agent system. Seven specialized agents collaborate to ensure all code changes pass through analysis, test-first development, implementation, compliance review, and knowledge persistence.

## Architecture

```
.opencode/
├── agents/                        # Agent responsibility summaries
│   ├── viabot-specialist.md       # Orchestrator (mode: all)
│   ├── analyst.md                 # Request analyzer (mode: all)
│   ├── qa-analyst.md              # Bug diagnostician (mode: all)
│   ├── tdd-writer.md              # TDD test writer (mode: all)
│   ├── builder.md                 # Implementer (mode: all)
│   ├── compliance.md              # Code reviewer (mode: all)
│   └── vault-updater.md           # Vault sync (mode: subagent)
├── prompts/                       # Full prompt content
│   └── *.md                       # One prompt file per agent
```

## Pipeline Flow

```
User Request
    │
    ▼
┌─────────────────────────────────────────────────┐
│           viabot-specialist (Orchestrator)       │
│  Decides route, dispatches subagents, manages    │
│  flow, presents compliance issues to user        │
└──────┬──────────────────────────────────────┬────┘
       │                                      │
       ▼                                      ▼
┌──────────────┐                     ┌──────────────┐
│   analyst    │ ◄── consult vault    │  qa-analyst  │
│  (classify)  │     (RAG)            │  (bug only)  │
└──────┬───────┘                     └──────┬────────┘
       │                                    │
       ▼                                    ▼
┌──────────────┐                     ┌──────────────┐
│  tdd-writer  │                     │   builder    │
│  (tests)     │ ◄──── tests ────────│ (implement)  │
└──────────────┘                     └──────┬───────┘
                                           │
                                           ▼
                                    ┌──────────────┐
                                    │  compliance  │
                                    │  (review)    │
                                    └──────┬───────┘
                                           │
                              ┌────────────┴────────────┐
                              │                         │
                          ⚠ Issues                  ✅ Clear
                              │                         │
                          Ask User                  Continue
                        ┌────┴────┐                     │
                     Fix? ═══ No                       ▼
                        ┌──────────┐           ┌──────────────┐
                        │  builder  │           │vault-updater │
                        │ (re-fix)  │           │ (sync)       │
                        └──────────┘           └──────────────┘
                                                     │
                                                     ▼
                                            ┌────────────────┐
                                            │ Report to User │
                                            └────────────────┘
```

### Early Exit Paths
- **docs/consult** → analyst only, report results
- **feature** → analyst → tdd-writer → builder → compliance → vault-updater
- **bug** → analyst → qa-analyst → tdd-writer → builder → compliance → vault-updater
- **refactor** → same as feature

## Agent Roles

| Agent | Mode | Permission | Responsibility |
|-------|------|-----------|---------------|
| `viabot-specialist` | all | edit:ask, bash:ask | Pipeline orchestration, routing, user-facing HITL decisions |
| `analyst` | all | edit:deny, bash:ask, mcp.obsidian:allow | Request analysis, vault consultation, classification |
| `qa-analyst` | all | edit:deny, bash:ask | Bug investigation, root cause analysis |
| `tdd-writer` | all | edit:allow, bash:ask | TDD test writing, failure verification |
| `builder` | all | edit:allow, bash:ask | Code implementation, test execution |
| `compliance` | all | edit:deny, bash:ask | Code review, quality gate |
| `vault-updater` | subagent | edit:deny, bash:deny, mcp.obsidian:allow | Obsidian vault sync |

## Key Design Decisions

### Why No Auto-Loop
Compliance returns issues to the orchestrator, which presents them to the user. The user decides whether to fix. This prevents infinite loops and gives the user control over quality vs. speed trade-offs.

### Why vault-updater is Subagent-Only
Vault operations should only be triggered by the orchestrator at the end of a successful pipeline run. Making it `mode: subagent` prevents accidental direct invocation.

### Hybrid Interaction Model
Users can invoke the orchestrator (`viabot-specialist`) for the full pipeline, or call individual subagents directly for specific tasks (e.g., "analyst, analyze this").

## Configuration

Defined in `opencode.json` under the `agent` key with `{file:}` references:

```json
{
  "default_agent": "viabot-specialist",
  "agent": {
    "viabot-specialist": {
      "mode": "all",
      "prompt": "{file:.opencode/agents/prompts/viabot-specialist.md}",
      "permission": { "edit": "ask", "bash": "ask" }
    }
  }
}
```

## Communication Rules
- **User interaction:** pt-br (Brazilian Portuguese)
- **Code, comments, docs, prompts:** English
- **Caveman mode:** All agents load the `caveman` skill for ultra-compressed communication (token economy)

## Related
- [[wiki/decisions/0004-multi-agent-pipeline]] — Decision record
- [[wiki/specs/ARCHITECTURE]] — Hexagonal Architecture
- [[wiki/specs/PORTS-ADAPTERS]] — Ports & Adapters
- [[wiki/specs/MVP-SCOPE]] — MVP scope
- `opencode.json` — Agent configuration
