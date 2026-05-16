---
title: Change Log
tags:
  - meta
  - log
type: concept
created: 2026-05-16T00:00:00.000Z
updated: '2026-05-16'
---

# Change Log

> Chronological log of all agent operations on the knowledge base.

## [2026-05-16] build | Initial project scaffold and design documents

- Created: [[wiki/specs/ARCHITECTURE]] — Hexagonal Architecture design doc
- Created: [[wiki/specs/DOMAIN]] — Domain model design doc
- Created: [[wiki/specs/PORTS-ADAPTERS]] — Ports & Adapters catalog
- Created: [[wiki/specs/MVP-SCOPE]] — MVP scope definition
- Created: `opencode.json`, `AGENTS.md`, `.opencode/instructions/`
- Created: RAG pipeline (copied from AscendEd, adapted paths)
- Verified deps via Context7: whatsmeow, mercadopago/sdk-go, stripe/stripe-go, mattn/go-sqlite3, jackc/pgx

## [2026-05-16] build | Rebranded to ViaBot

- Updated platform name from "WhatsApp Bot" to "ViaBot"
- Updated Go module path from `go.mau.fi/whatsbot-core` to `viabot.stream/sdk`
- Updated all design specs, AGENTS.md, and instructions with new naming

## [2026-05-16] build | Domain entities, collections, events, and application layer

- Added `Money.Multiply()`, `Money.Subtract()`, `OrderStatus.CanTransitionTo()`, `PhoneNumber` accessors
- Created: `item.go` — `OrderItem`, `CartItem` value objects with creation and subtotal calculation
- Created: `collections.go` — First-Class Collections (Rule 4): `OrderItemCollection`, `CartItemCollection`, `ProductCollection`
- Created: `events.go` — Domain events: `OrderPlaced`, `PaymentConfirmed`, `PaymentFailed`, `OrderCancelled`, `CartExpired`
- Created: `entity_order.go` — `Order` aggregate with status lifecycle (pending → awaiting_payment → confirmed → shipped → delivered)
- Created: `entity_cart.go` — `Cart` aggregate with AddItem, RemoveItem, UpdateQuantity, Clear, Checkout, expiration
- Created: `entity_customer.go` — `Customer` entity with name updates
- Created: `entity_product.go` — `Product` entity with availability management
- Expanded: `payment.go` — `Payment` constructor, `WebhookEvent` constructor + accessors, `Payment` confirm/reject behaviour
- Created: `core/application/` — 3 services: `OrderService`, `CartService`, `ProductService` (Hexagonal, DIP-compliant)
- Tests: 44 total (36 domain + 8 application with in-memory mocks), all passing

## [2026-05-16] build | SQLite adapter layer + public SDK + entry point

- Fixed: `reconstructOrder` now properly uses passed-in `items` instead of re-scanning DB (eliminated N+1 in `FindByID`)
- Created: `core/internal/sqlite/order_repository.go` — full OrderRepository with transaction-based Save, status machine persistence, payment round-trip
- Created: `core/internal/sqlite/product_repository.go` — ProductRepository with FindAll, FindByCategory, UpdateAvailability
- Created: `core/internal/sqlite/cart_repository.go` — CartRepository with ON DELETE CASCADE for cart_items
- Created: `core/internal/sqlite/customer_repository.go` — CustomerRepository with FindByPhone
- Created: `core/adapters.go` (`package sdk`) — public factory functions bridging internal adapters to entry points (Hexagonal DIP)
- Created: `customers/cliente-a/` — first customer entry point with full DI wiring (database → repositories → services → stub gateway)
- Added: `github.com/mattn/go-sqlite3 v1.14.44` dependency
- Updated: `go.work` includes `./customers/cliente-a`
- Tests: 17 new SQLite integration tests across all 4 repositories (with isolated :memory: databases)
- Fixed: `scanOrderItems` now accepts currency parameter — prevents silent Money creation with empty currency
- Fixed: `cart_items` schema now includes `currency DEFAULT 'BRL'` column — cart items properly track currency
- Created: [[wiki/decisions/0003-sqlite-currency-storage]] — Currency storage strategy decision record

## [2026-05-16] build | Multi-agent TDD pipeline with 7 agents

- Created: `.opencode/agents/` — 7 OpenCode agent definitions
- Created: `.opencode/agents/prompts/` — Full prompt files for each agent
- Agents: `viabot-specialist` (orchestrator), `analyst`, `qa-analyst`, `tdd-writer`, `builder`, `compliance`, `vault-updater`
- Pipeline: User → viabot-specialist → analyst (+qa if bug) → tdd-writer → builder → compliance (HITL) → vault-updater
- Updated: `opencode.json` — `default_agent` → `viabot-specialist`, all 7 agents with `{file:}` prompt references
- Created: [[wiki/specs/AGENT-PIPELINE]] — Multi-agent pipeline design spec
- Created: [[wiki/decisions/0004-multi-agent-pipeline]] — Multi-agent TDD enforcement decision record
