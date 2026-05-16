---
title: "MVP Scope"
tags: [spec, mvp, scope]
type: spec
created: 2026-05-16
updated: 2026-05-16
---

# MVP Scope — ViaBot

## Summary

Defines what is included in the MVP (Minimum Viable Product), what comes in Phase 2, and what is explicitly out of scope. The MVP goal is: **a customer can browse products and complete a purchase end-to-end via WhatsApp, paying with Pix or Boleto.**

---

## MVP Features — ✅ In Scope

### 1. WhatsApp Onboarding
- QR code pairing via whatsmeow
- Session persistence (SQLite-backed)
- Auto-reconnect on disconnect
- Single-device instance (MVP)

### 2. Customer Discovery
- Customer identified by WhatsApp JID (no registration needed)
- Customer profile created on first interaction
- Conversation state persisted per customer

### 3. Catalog Browsing
- Admin pre-loads products in database
- Customer views product list via "Catálogo" command
- Products show: name, description, price, image
- Categories for filtering

### 4. Cart Operations
- Add item to cart: `Adicionar [product]`
- Remove item: `Remover [product]`
- View cart: `Carrinho`
- Clear cart: `Limpar carrinho`
- Cart persisted per customer (survives disconnect)

### 5. Checkout Flow
- Customer requests checkout: `Finalizar compra`
- Bot presents payment options: `Pix (recomendado)` or `Boleto`
- Customer selects payment method
- Bot generates payment (Pix QR code or boleto link)
- Bot sends payment instructions + deadline

### 6. Payment Confirmation (Webhook)
- Mercado Pago webhook endpoint: `POST /webhooks/mercadopago`
- On `payment.approved`:
  - Order status → `confirmed`
  - Send receipt to customer
  - Store Payment details in order

### 7. Order History
- Customer views their orders: `Meus pedidos`
- Shows status: `Aguardando pagamento`, `Confirmado`, `Enviado`, `Entregue`, `Cancelado`

### 8. Admin API (API-Only MVP)
- List products (CRUD)
- List orders
- Update order status
- View customer list
- Simple API key auth

### 9. Error Handling & UX
- Friendly error messages in pt-BR
- Timeout handling (payment expiration)
- Invalid command handling
- Session timeout (24h cart expiry)

---

## Phase 2 Features — 📋 Future

| Feature | Description |
|---------|-------------|
| **Card Payments** | Credit/debit card via Mercado Pago |
| **Stripe Gateway** | International payments (USD) |
| **Admin React UI** | Full admin SPA with dashboard |
| **Delivery Tracking** | Shipping status, tracking codes |
| **Inventory Mgmt** | Stock tracking, low-stock alerts |
| **Multi-Tenant** | Separate customers/databases per client |
| **Telegram Channel** | Multi-channel via same codebase |
| **Order Notifications** | Proactive status push to customer |
| **Refund Flow** | Customer requests refund via bot |
| **Analytics Dashboard** | Sales, customers, popular products |

---

## Phase 3+ Features — 🚀 Future

| Feature | Description |
|---------|-------------|
| **Finance Domain** | Invoices, subscriptions, split payments |
| **Mentor IA Domain** | FAQ automation, AI chat |
| **Marketplace** | Multiple sellers |
| **Email/SMS Channels** | Additional messenger adapters |
| **Web Admin** | Full admin panel |
| **Stripe Connect** | Multi-seller payouts |
| **Conversation UI** | Rich message history viewer |

---

## What's NOT in MVP (Explicitly Out of Scope)

| Item | Reason |
|------|--------|
| **Telegram/Email/SMS** | Keep channel count = 1 for focus |
| **Stripe** | Only BRL needed for MVP; Mercado Pago covers Pix/Boleto |
| **Credit Card** | Card requires PCI compliance scope; Pix is instant and cheaper |
| **React Admin UI** | API-only admin (Postman/curl) reduces frontend complexity |
| **Multi-tenant** | Single-tenant MVP → multi-tenant when needed |
| **Inventory tracking** | MVP assumes unlimited stock; inventory is Phase 2 |
| **Delivery tracking** | Physical delivery details defined by seller after sale |
| **Refunds** | Admin manually refunds via MP dashboard in MVP |
| **User registration** | WhatsApp JID is the identity — no email/password |
| **Marketing features** | No coupons, discounts, loyalty points |
| **Reports & analytics** | Raw data available via API; dashboards are Phase 2 |
| **Unit/Integration tests for adapters** | MVP tests focus on domain logic; adapter tests are Phase 2 |
| **Docker/CI-CD** | Manual deployment in MVP; CI/CD is Phase 2 |
| **Rate limiting** | MVP assumes low volume; rate limiting is Phase 2 |
| **Encryption at rest** | SQLite file permissions suffice for MVP; production requires encryption |

---

## Architecture Decisions for MVP

### Why SQLite (not pgx) for MVP?

| Factor | SQLite | pgx/PostgreSQL |
|--------|--------|----------------|
| Setup | Zero — file-based | Requires PostgreSQL server |
| Dev speed | Instant | Need Docker/install |
| whatsmeow requirement | Already uses SQLite | Adds another DB |
| Performance for MVP | ✅ Sufficient | Overkill |
| Migration to prod | Export SQL → pg import | Native |

**Decision:** SQLite for MVP. pgx in Phase 2 when multi-instance/schema-per-tenant is needed.

### Why Mercado Pago Pix/Boleto (not Card)?

| Factor | Pix | Boleto | Card |
|--------|-----|--------|------|
| Fee | 0.99% | R$1.99 flat | 3.99% |
| Settlement | Instant | 1-3 days | 30 days |
| Chargeback risk | Near zero | Near zero | Present |
| PCI scope | None | None | Requires PCI SAQ |
| User convenience | QR code scan | Print/online pay | Enter card details |

**Decision:** Pix primary (recommended) + Boleto secondary. Card in Phase 2 (after PCI assessment).

### Why whatsmeow (not Cloud API) for MVP?

| Factor | whatsmeow | Cloud API |
|--------|-----------|-----------|
| Setup | QR code pair | Meta approval needed |
| Cost | Free | Pay per conversation |
| Multi-device | ✅ Native | ✅ Supported |
| Production readiness | OK for MVP | Better for scale |
| Rate limits | Lower | Higher |

**Decision:** whatsmeow for MVP (develop locally, minimal cost). Cloud API evaluation for production rollout.

---

## MVP User Journeys

### Happy Path: Pix Purchase

```
Customer                        Bot
  │                              │
  │  "Catálogo"                  │
  │ ──────────────────────────►  │
  │                    List products
  │  ◄────────────────────────── │
  │                              │
  │  "Adicionar Camiseta"        │
  │ ──────────────────────────►  │
  │                    Item added to cart
  │  ◄────────────────────────── │
  │                              │
  │  "Finalizar compra"          │
  │ ──────────────────────────►  │
  │                    Offer Pix/Boleto
  │  ◄────────────────────────── │
  │                              │
  │  "Pix"                       │
  │ ──────────────────────────►  │
  │                    Generate Pix QR code
  │  ◄────────────────────────── │
  │      [QR Code Image]         │
  │      [Copy-paste code]       │
  │      "Pague até 15min"       │
  │                              │
  │  (Customer pays via Pix)     │
  │                              │
  │        Webhook: approved     │
  │ ──────────────────────────►  │
  │                    "Pagamento confirmado!"
  │  ◄────────────────────────── │
  │      "Pedido #123 confirmado"
  │      [Receipt with amount]   │
```

### Error Path: Payment Expires

```
Customer                        Bot
  │                              │
  │  (selects Boleto)            │
  │ ──────────────────────────►  │
  │                    Generate boleto
  │  ◄────────────────────────── │
  │      [Boleto link]           │
  │      "Pague até 3 dias"      │
  │                              │
  │  (3 days pass — no payment)  │
  │                              │
  │        Webhook: expired      │
  │ ──────────────────────────►  │
  │                    "Pagamento expirou"
  │  ◄────────────────────────── │
  │      "Deseja gerar novo?"    │
```

---

## Related

- [[wiki/specs/ARCHITECTURE]] — How everything fits together
- [[wiki/specs/DOMAIN]] — What domain objects exist
- [[wiki/specs/PORTS-ADAPTERS]] — Which interfaces/adapters are built

## References

- [Mercado Pago: Pix](https://www.mercadopago.com.br/pix) — Pix integration guide
- [Mercado Pago: Boleto](https://www.mercadopago.com.br/boletos) — Boleto integration
- [whatsmeow: Getting Started](https://github.com/tulir/whatsmeow) — WhatsApp library
