# ViaBot — Agent Knowledge Base (LLM Wiki)

> **Pattern:** Andrej Karpathy's LLM Wiki — agents build and maintain a compounding knowledge base.
> **Vault location:** `obsidian-vault/` (within the ViaBot workspace)
> **Viewer:** Obsidian (graph view shows cross-references)
> **Agents:** Read/write markdown. **Humans:** Curate sources, guide direction, audit.

---

## Architecture — Three Layers

```markdown
raw/                  # Immutable source documents (agents READ only)
├── docs/             # Project docs (ADRs, runbooks, policies)
└── external/         # Web-clipped articles, reference docs

wiki/                 # Agent-maintained knowledge (agents READ + WRITE)
├── index.md          # Catalog of all pages — updated on every ingest
├── log.md            # Chronological append-only log of operations
├── specs/            # Design documents (ARCHITECTURE, DOMAIN, PORTS-ADAPTERS, MVP-SCOPE)
├── entities/         # Entity pages (services, databases, gateways, channels)
├── concepts/         # Concept pages (Hexagonal Architecture, Value Objects, Calisthenics)
├── decisions/        # Architecture Decision Records (ADRs)
├── bugs/             # Documented bugs, quirks, and workarounds
└── templates/        # Page templates for consistency

AGENTS.md             # This file — schema for how agents maintain the wiki
```

---

## Page Format Conventions

Every wiki page MUST follow this template:

```markdown
---
title: "Page Title"
tags: [tag1, tag2]
type: entity|concept|decision|bug|spec
created: YYYY-MM-DD
updated: YYYY-MM-DD
sources: []
---

# Page Title

## Summary
One-paragraph summary of what this page covers.

## Details
...

## Related
- [[wiki/entities/...]]
- [[wiki/concepts/...]]

## References
- [Title](url)
```

**Rules:**

- `updated:` date MUST be refreshed when content changes.
- `tags:` use lowercase, no spaces, kebab-case for multi-word.
- `type:` MUST be one of: `entity`, `concept`, `decision`, `bug`, `spec`.

---

## Operations

### Ingest — Process a new source document

When a human drops a file into `raw/` and says "process this":

1. Read the source document from `raw/`.
2. Extract key information: entities, concepts, decisions, facts.
3. Create or update relevant wiki pages.
4. Update `wiki/index.md` with new entries.
5. Log the operation in `wiki/log.md`.

### Query — Find information

1. Read `wiki/index.md` to find relevant pages.
2. Use Obsidian wikilinks `[[wiki/...]]` to cross-reference.
3. When answering questions, cite sources.

### Lint — Health check

1. Verify all wikilinks resolve to existing pages.
2. Check for outdated `updated:` dates (> 30 days → flag).
3. Detect orphan pages (not linked from `index.md`).

---

## Wikilink Convention

Use Obsidian Flavored Markdown — `[[path/page]]` without `.md` extension:

| Correct | Incorrect |
| ------- | --------- |
| `[[wiki/specs/ARCHITECTURE]]` | `[[wiki/specs/ARCHITECTURE.md]]` |
| `[[wiki/concepts/value-objects]]` | `[[./value-objects]]` |
| `[[wiki/entities/mercadopago]]` | `[[entities/mercadopago]]` |

---

**Last Updated:** 2026-05-16
