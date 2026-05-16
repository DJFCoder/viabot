---
title: "Architecture Overview"
tags: [spec, architecture, hexagonal]
type: spec
created: 2026-05-16
updated: 2026-05-16
---

# Architecture Overview — ViaBot

## Summary

**ViaBot** uses **Hexagonal Architecture (Ports & Adapters)** combined with **Domain-Driven Design (DDD)** to create a multi-domain chatbot platform that can expand to multiple messaging channels, payment gateways, and database backends without modifying core business logic.

---

## Rationale

### Why Hexagonal Architecture?

| Requirement | Hexagonal Solution |
|-------------|-------------------|
| Multi-channel (WhatsApp, Telegram, Email, SMS) | Each channel is an Adapter implementing the same `Messenger` Port |
| Multi-gateway (Mercado Pago, Stripe) | Each gateway is an Adapter implementing the same `PaymentGateway` Port |
| Multi-database (SQLite dev, PostgreSQL prod) | Each database is an Adapter implementing the same `Repository` Port |
| Testability | Domain logic depends on interfaces — mock adapters in tests |
| Independent evolution | Add a new channel without touching domain code (OCP) |

### Why DDD?

| Requirement | DDD Solution |
|-------------|-------------|
| Complex domains (e-commerce, finance, mentor IA) | Aggregates, Value Objects, Domain Events keep each domain isolated |
| Ubiquitous Language | Shared vocabulary between devs, PMs, and stakeholders |
| Bounded Contexts | E-commerce and Finance don't leak into each other |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Entry Points                             │
│  customers/client-a/main.go                                  │
│  customers/client-b/main.go                                  │
│  (Future: admin-ui/, API handlers)                           │
└───────────┬─────────────────────────────────────┬───────────┘
            │                                     │
            ▼                                     ▼
┌───────────────────────┐         ┌───────────────────────────┐
│   Application Layer   │         │   Application Layer       │
│  (core/application/)  │         │  (shared services)        │
│  Use Cases / Services │         │                           │
└──────┬────────────────┘         └───────────┬───────────────┘
       │                                      │
       ▼                                      ▼
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer (core/domain/)              │
│                                                             │
│  Ports (Interfaces):       Domain Types:                    │
│  ┌──────────────────┐     ┌──────────────────┐              │
│  │ Messenger        │     │ Order            │              │
│  │ PaymentGateway   │     │ Customer         │              │
│  │ OrderRepository  │     │ Product          │              │
│  │ CustomerRepo     │     │ Money (VO)       │              │
│  │ ProductRepo      │     │ OrderStatus (VO) │              │
│  └──────────────────┘     │ PhoneNumber (VO) │              │
│                            └──────────────────┘              │
└─────────────────────────────────────────────────────────────┘
            │                        │
            ▼                        ▼
┌─────────────────────────────────────────────────────────────┐
│                 Infrastructure Layer (core/internal/)        │
│                                                             │
│  Messenger Adapters:    Payment Adapters:    Repository:    │
│  ┌────────────────┐    ┌────────────────┐   ┌────────────┐  │
│  │ whatsmeow      │    │ mercadopago    │   │ sqlite     │  │
│  │ (WhatsApp)     │    │ (Pix/Boleto/   │   │ pgx        │  │
│  │                │    │  Card)         │   │ (futuro)   │  │
│  │ (future:       │    │                │   └────────────┘  │
│  │  Telegram,     │    │ (future:       │                    │
│  │  Email, SMS)   │    │  Stripe)       │                    │
│  └────────────────┘    └────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

---

## Layer Responsibilities

### 1. Domain Layer (`core/domain/`)

**Zero external dependencies.** Pure Go types and interfaces.

| Element | Purpose |
|---------|---------|
| **Ports (interfaces)** | Contracts for external interactions: `Messenger`, `PaymentGateway`, `Repository` |
| **Entities** | Types with identity: `Order`, `Customer`, `Product` |
| **Value Objects** | Immutable types without identity: `Money`, `OrderStatus`, `PhoneNumber`, `CustomerID` |
| **Domain Events** | Things that happened: `OrderPlaced`, `PaymentConfirmed` |

### 2. Application Layer (`core/application/` or inline with handlers)

Orchestrates domain objects to fulfill use cases.

| Use Case | Flow |
|----------|------|
| Place Order | Customer → creates cart → validates items → calculates total → creates Order → returns payment link |
| Confirm Payment | Webhook → finds Order → marks as paid → sends confirmation → updates inventory |
| List Products | Query → returns Product collection |

**Dependencies:** Only domain layer interfaces.

### 3. Infrastructure Layer (`core/internal/`)

Implements domain interfaces. May have external dependencies.

| Adapter Package | Implements | External Dep |
|----------------|------------|-------------|
| `whatsmeow/` | `Messenger` (send/receive) | `go.mau.fi/whatsmeow` |
| `mercadopago/` | `PaymentGateway` | `github.com/mercadopago/sdk-go` |
| `sqlite/` | `Repository` interfaces | `github.com/mattn/go-sqlite3` |

### 4. Entry Points (`customers/*/`)

Wires everything together — creates adapters, injects dependencies, starts the bot.

```go
// customers/client-a/main.go
func main() {
    db := openSQLite()
    messenger := whatsmeow.NewMessenger(db)
    gateway := mercadopago.NewGateway(os.Getenv("MP_TOKEN"))
    repo := sqlite.NewOrderRepository(db)

    handler := application.NewOrderHandler(gateway, repo, messenger)
    // ...
}
```

---

## Core SDK Pattern

### Monorepo (MVP Phase)

```
whatsapp-bot/
├── core/          # viabot.stream/sdk — public Go package
├── customers/
│   ├── client-a/  # go.work → replace viabot.stream/sdk => ./core
│   └── client-b/
└── go.work        # join core, customers/client-a, customers/client-b
```

### Multi-Repo (Future Phase)

```
viabot-sdk/           # Separate GitHub repo
├── go.mod                   # module viabot.stream/sdk
├── domain/
└── internal/

client-a-repo/               # Separate repo
├── go.mod                   # require viabot.stream/sdk
├── git submodule viabot-sdk
├── main.go
```

---

## Design Patterns Applied

| Pattern | Location | Why |
|---------|----------|-----|
| **Repository** | `core/domain/` | Abstract data access (SQLite ↔ pgx swap) |
| **Strategy** | `core/domain/` | Swappable payment gateways |
| **Adapter** | `core/internal/` | Wrap external SDKs in domain interfaces |
| **Factory** | Entry points | Create adapters based on config |
| **Observer** | `core/domain/` (EventBus) | Decoupled webhook → handler propagation |
| **Command** | Application layer | Order state change undo/redo (future) |
| **Facade** | Application layer | `CheckoutFacade` hides payment + messaging + persistence |

---

## Dependency Injection

**No DI framework.** Manual constructor injection at entry points.

```go
type CheckoutHandler struct {
    gateway   PaymentGateway
    messenger MessageSender
    orders    OrderRepository
}

func NewCheckoutHandler(
    gateway PaymentGateway,
    messenger MessageSender,
    orders OrderRepository,
) *CheckoutHandler {
    return &CheckoutHandler{
        gateway:   gateway,
        messenger: messenger,
        orders:    orders,
    }
}
```

---

## Related

- [[wiki/specs/DOMAIN]] — Ubiquitous Language, aggregates, entities, value objects
- [[wiki/specs/PORTS-ADAPTERS]] — Interface definitions and adapter implementations
- [[wiki/specs/MVP-SCOPE]] — What's in and out of scope for MVP
- [[wiki/concepts/hexagonal-architecture]] — Deep dive into Ports & Adapters

## References

- [Alistair Cockburn — Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Context7: whatsmeow](https://context7.com/tulir/whatsmeow) — WhatsApp multi-device library
- [Context7: Mercado Pago SDK Go](https://context7.com/mercadopago/sdk-go) — Payment gateway
- [Context7: pgx](https://context7.com/jackc/pgx) — PostgreSQL driver
