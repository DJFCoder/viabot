---
title: "0004 — Multi-Agent TDD Enforcement Pipeline"
tags: [decision, agents, pipeline, tdd]
type: decision
created: 2026-05-16
updated: 2026-05-16
status: accepted
---

# 0004 — Multi-Agent TDD Enforcement Pipeline

## Context

The ViaBot project requires a rigorous quality enforcement mechanism. Previous sessions showed that without a pipeline, code was written without tests, compliance checks were manual, and knowledge was not consistently persisted to the vault. The team needed an automated system that:

1. Enforces TDD (tests before implementation)
2. Applies Object Calisthenics and SOLID rules consistently
3. Prevents PII exposure
4. Persists new knowledge to the Obsidian vault
5. Prevents infinite loops between review and implementation
6. Conserves tokens through compressed communication

## Decision

Implement a **multi-agent TDD enforcement pipeline** using OpenCode's custom agent system (`.opencode/agents/`):

### Architecture
- **7 specialized agents** with distinct responsibilities
- **Orchestrator pattern**: `viabot-specialist` coordinates the pipeline
- **Subagent dispatch**: Each phase runs as a separate `Task` call with isolated context
- **HITL gate**: Compliance issues always ask the user before re-implementation

### Agent Roles

| Agent | Role |
|-------|------|
| `viabot-specialist` | Orchestrator — routes, dispatches, manages flow, presents HITL decisions |
| `analyst` | Classifies request, consults vault for context, produces structured spec |
| `qa-analyst` | Debugs bugs, finds root cause, produces fix spec |
| `tdd-writer` | Writes failing tests first (TDD) |
| `builder` | Implements code to pass tests, enforces quality rules |
| `compliance` | Reviews code against all quality rules, never auto-approves |
| `vault-updater` | Syncs new knowledge to Obsidian vault |

### Configuration
- Agents defined in `opencode.json` using `{file:}` references to `prompts/` directory
- `prompts/` contains full agent instructions
- `agents/` contains responsibility summaries
- `default_agent` set to `viabot-specialist`

### Communication
- All agents use the `caveman` skill for compressed communication (token economy)
- User-facing output in pt-br
- Code, docs, prompts in English

## Consequences

### Positive
- ✅ TDD enforced at pipeline level — no code before tests
- ✅ Compliance catches violations before they reach production
- ✅ Knowledge automatically persisted to vault
- ✅ No infinite loops — HITL gate prevents auto-remediation
- ✅ Token savings via caveman mode
- ✅ Hybrid model: full pipeline or direct subagent calls

### Negative
- ❌ Pipeline adds latency to simple changes (e.g., one-line fix still goes through all phases)
- ❌ More subagent dispatches = more token consumption than manual approach
- ❌ Complexity: 7 agents to maintain and debug

### Mitigations
- Early exit paths for docs-only requests skip TDD/Builder/Compliance
- Responsibility summaries in `agents/` make agent roles discoverable
- Vault spec documents the full pipeline for onboarding

## Related
- [[wiki/specs/AGENT-PIPELINE]] — Full pipeline spec
- [[wiki/specs/ARCHITECTURE]] — Hexagonal Architecture
- [[wiki/decisions/0003-sqlite-currency-storage]] — Previous decision record
- `opencode.json` — Agent configuration
- `AGENTS.md` — Project agent rules
