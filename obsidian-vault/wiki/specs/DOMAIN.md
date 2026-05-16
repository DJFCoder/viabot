---
title: "Domain Model"
tags: [spec, domain, ddd]
type: spec
created: 2026-05-16
updated: 2026-05-16
---

# Domain Model — ViaBot

## Summary

This document defines the Ubiquitous Language, aggregates, entities, value objects, and domain events for ViaBot's e-commerce domain (MVP). Future domains (Finance, Mentor IA) will have their own bounded contexts.

---

## Ubiquitous Language

| Term | Definition |
|------|-----------|
| **Customer** | A person who interacts with the bot via WhatsApp to browse and purchase products |
| **Product** | An item available for purchase (physical or digital) |
| **Catalog** | A curated collection of products visible to customers |
| **Cart** | A temporary collection of items a customer intends to purchase |
| **Order** | A confirmed request to purchase items, with payment status tracking |
| **Payment** | A financial transaction to pay for an order |
| **Pix** | Instant payment method in BRL (Brazilian real-time payment) |
| **Boleto** | Brazilian bank slip payment method (paid in up to 3 days) |
| **Receipt** | Confirmation sent to customer after payment is confirmed |
| **Messenger** | The communication channel (WhatsApp, Telegram, etc.) |

---

## Aggregates

### Order Aggregate (MVP)

**Root Entity:** `Order`

```
┌─────────────────────────────────┐
│           Order                  │
│  ID: OrderID                     │
│  CustomerID: CustomerID          │
│  Status: OrderStatus             │
│  Items: OrderItemCollection      │
│  Total: Money                    │
│  Payment: *Payment               │
│  CreatedAt: time.Time            │
│  ConfirmedAt: *time.Time         │
│                                  │
│  Public Behavior:                │
│  - Place() → Payment             │
│  - ConfirmPayment(payment)       │
│  - Cancel()                      │
└─────────────────────────────────┘
```

**Invariants:**
- Order must have at least 1 item
- Order total must equal sum of item totals
- Status transitions: `pending` → `awaiting_payment` → `confirmed` → `shipped` → `delivered`
- Only cancel if status is `pending` or `awaiting_payment`
- A confirmed order cannot be modified

### Cart Aggregate (MVP)

**Root Entity:** `Cart`

```
┌─────────────────────────────────┐
│            Cart                  │
│  ID: CartID                      │
│  CustomerID: CustomerID          │
│  Items: CartItemCollection       │
│  Total: Money                    │
│  ExpiresAt: time.Time            │
│                                  │
│  Public Behavior:                │
│  - AddItem(product, quantity)    │
│  - RemoveItem(productID)         │
│  - UpdateQuantity(productID, qty)│
│  - Clear()                       │
│  - Checkout() → Order            │
└─────────────────────────────────┘
```

**Invariants:**
- Quantity per item > 0
- Cart expires after 24 hours of inactivity
- Checkout converts Cart to Order (Cart is then discarded)

---

## Entities

### Customer

```go
type Customer struct {
    id        CustomerID
    phone     PhoneNumber
    name      string
    createdAt time.Time
}
```

**Identity:** `CustomerID` (derived from WhatsApp JID)

### Product

```go
type Product struct {
    id          ProductID
    name        string
    description string
    price       Money
    category    string
    available   bool
    imageURL    string
}
```

**Identity:** `ProductID` (UUID)

---

## Value Objects

### Money

```go
type Money struct {
    amount   Decimal
    currency Currency  // "BRL", "USD", "EUR"
}

func NewMoney(amount Decimal, currency Currency) (Money, error)
func (m Money) Add(other Money) (Money, error)
func (m Money) Subtract(other Money) (Money, error)
func (m Money) Multiply(factor int) Money
func (m Money) IsZero() bool
func (m Money) IsNegative() bool
```

**Invariants:**
- `amount` must not be negative
- Operations validate currency match (`BRL` + `BRL` = OK, `BRL` + `USD` = error)

### OrderStatus

```go
type OrderStatus string

const (
    OrderStatusPending         OrderStatus = "pending"
    OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
    OrderStatusConfirmed       OrderStatus = "confirmed"
    OrderStatusShipped         OrderStatus = "shipped"
    OrderStatusDelivered       OrderStatus = "delivered"
    OrderStatusCancelled       OrderStatus = "cancelled"
)

func (s OrderStatus) CanTransitionTo(target OrderStatus) bool
```

### PhoneNumber

```go
type PhoneNumber struct {
    countryCode int
    number      string
}

func NewPhoneNumber(countryCode int, number string) (PhoneNumber, error)
// Validates format, country code must exist
```

### CustomerID

```go
type CustomerID struct {
    value string  // WhatsApp JID format: "5511999999999@s.whatsapp.net"
}
```

### OrderID, ProductID, CartID

```go
type OrderID struct{ value string }   // UUID
type ProductID struct{ value string } // UUID
type CartID struct{ value string }    // UUID
```

---

## Domain Events

Emitted by aggregates and consumed by application layer (or future EventBus).

| Event | When | Payload |
|-------|------|---------|
| `OrderPlaced` | Customer completes checkout | orderID, customerID, total, paymentLink |
| `PaymentConfirmed` | Webhook confirms payment | orderID, paymentID, amount |
| `PaymentFailed` | Payment declined/expired | orderID, reason |
| `OrderCancelled` | Customer or admin cancels | orderID, reason |
| `CartExpired` | Cart inactive for 24h | cartID, customerID |

### Example

```go
type OrderPlaced struct {
    OrderID     OrderID
    CustomerID  CustomerID
    Total       Money
    PaymentLink string
    OccurredAt  time.Time
}
```

---

## First-Class Collections (Rule 4)

### OrderItemCollection

```go
type OrderItemCollection struct {
    items []OrderItem
}

func (c OrderItemCollection) Total() Money
func (c OrderItemCollection) Count() int
func (c OrderItemCollection) Items() []OrderItem  // returns copy, not reference
```

### CartItemCollection

```go
type CartItemCollection struct {
    items []CartItem
}

func (c CartItemCollection) Total() Money
func (c CartItemCollection) Count() int
```

### ProductCollection

```go
type ProductCollection struct {
    items []Product
}

func (c ProductCollection) FilterByCategory(category string) ProductCollection
func (c ProductCollection) Search(term string) ProductCollection
```

---

## Bounded Contexts (Future)

```
┌───────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│   E-Commerce      │   │    Finance        │   │   Mentor IA      │
│  (MVP)            │   │  (Phase 2)        │   │  (Phase 3+)      │
│                   │   │                   │   │                  │
│  Order            │   │  Invoice          │   │  FAQ             │
│  Cart             │   │  Subscription     │   │  KnowledgeBase   │
│  Product          │   │  SplitPayment     │   │  ChatSession     │
│  Catalog          │   │  Wallet           │   │  Intent          │
└───────────────────┘   └──────────────────┘   └──────────────────┘
```

Each bounded context has its own:
- Aggregates and entities
- Ubiquitous Language
- Repository interfaces
- Domain events

---

## Related

- [[wiki/specs/ARCHITECTURE]] — Layer overview and rationale
- [[wiki/specs/PORTS-ADAPTERS]] — Interfaces catalog
- [[wiki/specs/MVP-SCOPE]] — What's included in each phase

## References

- [DDD Quickly](https://www.infoq.com/minibooks/domain-driven-design-quickly/)
- [Martin Fowler — DDD](https://martinfowler.com/tags/domain%20driven%20design.html)
- [Refactoring Guru — Value Object](https://refactoring.guru/design-patterns/value-object)
