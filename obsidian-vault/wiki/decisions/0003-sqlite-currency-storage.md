---
title: 0003 — SQLite Currency Storage Strategy
tags:
  - decision
  - architecture
  - sqlite
type: decision
created: '2026-05-16'
status: accepted
---
# 0003 — SQLite Currency Storage Strategy

## Context
When persisting monetary values to SQLite, we need to decide where to store the currency code ("BRL", "USD", etc.) for each aggregate.

## Options Considered

### Option A — Currency only on the aggregate root (Order/Product)
- `orders` table has `currency` column
- `products` table has `currency` column
- `order_items` and `cart_items` do NOT store currency — items inherit from parent

**Pros:** Less storage, no redundancy, single source of truth  
**Cons:** Cannot query item prices without the parent context

### Option B — Currency on every monetary value
- Every table with a `unit_price` or `amount` also stores `currency`

**Pros:** Self-contained rows, easier debugging  
**Cons:** Redundant, violates DRY

## Decision
**Hybrid approach:**

| Table | Currency stored? | Reason |
|-------|-----------------|--------|
| `orders` | ✅ `currency` column | Aggregate root defines currency context |
| `order_items` | ❌ Not stored | Inherits from order — always queried via order |
| `products` | ✅ `currency` column | Products are queried independently |
| `cart_items` | ✅ `currency` column (DEFAULT 'BRL') | Carts may need items without cart context |
| `carts` | ❌ Not stored | Cart items carry their own currency |
| `payments` | ❌ (in `orders` table) | Payment is embedded in the order row |

**Rationale:**
- Product prices need currency for standalone catalog display (prices may be in different currencies)
- Cart items need currency because the cart aggregate doesn't have its own currency column
- Order items don't need their own currency because the order's `currency` is the single source of truth for the transaction

## Consequences
- `scanOrderItems` must receive the currency as a parameter from the order context
- `scanCartItems` can read currency directly from the `cart_items` table
- Migration adds `currency TEXT NOT NULL DEFAULT 'BRL'` to `cart_items` table
