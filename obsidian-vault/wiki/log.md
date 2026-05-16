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
