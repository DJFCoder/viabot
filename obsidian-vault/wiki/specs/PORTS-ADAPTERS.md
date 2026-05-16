---
title: "Ports & Adapters Catalog"
tags: [spec, hexagonal, ports, adapters]
type: spec
created: 2026-05-16
updated: 2026-05-16
---

# Ports & Adapters Catalog

## Summary

Complete catalog of all Ports (interfaces in `core/domain/`) and their Adapters (implementations in `core/internal/`). This is the contract between the domain and the outside world — new channels, gateways, and databases are added by writing new Adapters.

---

## Port: Messenger

**File:** `core/domain/messenger.go`

### Interface

```go
// Messenger handles bidirectional message exchange with customers.
// Segregated into sender/receiver for ISP compliance.
package domain

type MessageSender interface {
    Send(ctx context.Context, to Recipient, message *OutboundMessage) error
}

type MessageReceiver interface {
    Receive(ctx context.Context, handler MessageHandler) error
}

type Messenger interface {
    MessageSender
    MessageReceiver
}
```

### Supporting Types

```go
type Recipient struct {
    id   CustomerID   // JID for WhatsApp, chat ID for Telegram, email for Email
    channel Channel   // which channel this recipient uses
}

type OutboundMessage struct {
    text     string
    media    *MediaAttachment
    buttons  []Button
}

type InboundMessage struct {
    from     CustomerID
    text     string
    media    *MediaAttachment
    received time.Time
}

type MessageHandler func(ctx context.Context, msg *InboundMessage) error

type MediaAttachment struct {
    mimeType string
    data     []byte
    filename string
}

type Button struct {
    id    string
    label string
}

type Channel string
const (
    ChannelWhatsApp Channel = "whatsapp"
    ChannelTelegram Channel = "telegram"
)
```

**Why this shape:**
- ISP: A component that only sends (notifications) doesn't need to know about receiving
- `Recipient` encapsulates routing info per channel
- `MediaAttachment` handles image/products/catalog messages

### Adapters

#### 1. WhatsApp — whatsmeow (MVP)

**Package:** `core/internal/whatsmeow/`

| Technology | whatsmeow |
|-----------|-----------|
| Module | `go.mau.fi/whatsmeow` |
| Context7 | `/tulir/whatsmeow` |
| Status | ✅ MVP |
| Session | SQLite-backed (container.GetFirstDevice) |
| Messaging | Text, media, interactive buttons |
| QR Pairing | via GetQRChannel on first connect |

**Implementation approach:**
```go
type Messenger struct {
    client *whatsmeow.Client
}

func NewMessenger(dbPath string, logLevel string) (*Messenger, error)
// - Opens SQLite container
// - Gets or creates device store
// - Creates whatsmeow client
// - Registers event handler for incoming messages

func (m *Messenger) Send(ctx context.Context, to Recipient, msg *OutboundMessage) error
// - Parses JID from to.id
// - Builds whatsmeow proto.Message (Conversation, ImageMessage, etc.)
// - Calls client.SendMessage

func (m *Messenger) Receive(ctx context.Context, handler MessageHandler) error
// - Starts goroutine for client.AddEventHandler
// - Calls client.Connect()
// - Routes events.Message to handler
// - Blocks until context cancelled

func (m *Messenger) Disconnect() error
```

#### 2. Telegram — Future

**Status:** 📋 Future (Phase 2)

**Adapter idea:**
```go
// core/internal/telegram/
type Messenger struct {
    bot *tgbotapi.BotAPI
}
// Same Messenger interface, Telegram-specific implementation
```

#### 3. Email / SMS — Future

**Status:** 📋 Future (Phase 3+)

---

## Port: PaymentGateway

**File:** `core/domain/payment.go`

### Interface

```go
package domain

type PaymentGateway interface {
    CreatePayment(ctx context.Context, order *Order) (*Payment, error)
    GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)
    HandleWebhook(ctx context.Context, payload []byte) (*WebhookEvent, error)
}
```

### Supporting Types

```go
type Payment struct {
    id            string       // gateway payment ID
    orderID       OrderID
    status        PaymentStatus
    amount        Money
    paidAmount    Money
    method        PaymentMethod
    checkoutLink  string       // URL to complete payment (Pix QR, boleto barcode)
    qrCode        string       // Pix: base64 QR image
    qrCodeText    string       // Pix: copy-paste code
    boletoBarcode string       // Boleto: barcode number
    boletoURL     string       // Boleto: printable slip URL
    expiresAt     time.Time
    createdAt     time.Time
    confirmedAt   *time.Time
}

type PaymentStatus string
const (
    PaymentPending   PaymentStatus = "pending"
    PaymentApproved  PaymentStatus = "approved"
    PaymentRejected  PaymentStatus = "rejected"
    PaymentRefunded  PaymentStatus = "refunded"
    PaymentCancelled PaymentStatus = "cancelled"
    PaymentExpired   PaymentStatus = "expired"
)

type PaymentMethod string
const (
    PaymentPix    PaymentMethod = "pix"
    PaymentBoleto PaymentMethod = "boleto"
    PaymentCard   PaymentMethod = "credit_card"
)

type WebhookEvent struct {
    action    string       // "payment.created", "payment.updated"
    paymentID string
    orderID   OrderID
    status    PaymentStatus
}
```

**Why this shape:**
- Payment carries all possible payment method details (Pix QR, boleto barcode)
- Status enum covers Mercado Pago and Stripe statuses
- WebhookEvent normalizes gateway-specific webhooks into domain events

### Adapters

#### 1. Mercado Pago (MVP)

**Package:** `core/internal/mercadopago/`

| Technology | Mercado Pago SDK |
|-----------|-----------------|
| Module | `github.com/mercadopago/sdk-go` |
| Context7 | `/mercadopago/sdk-go` |
| Status | ✅ MVP |
| Methods | Pix (0.99%), Boleto (R$1.99), Card (3.99%) |
| Currency | BRL |

**Implementation approach:**
```go
type Gateway struct {
    client *payment.Client   // for card payments
    // preference client for checkout preference (Pix)
    // order client for unified Orders API (future)
}

func NewGateway(accessToken string) (*Gateway, error)
// - Creates config.New(accessToken)
// - Creates payment client
// - Preloads available payment methods

func (g *Gateway) CreatePayment(ctx context.Context, order *Order) (*Payment, error)
// - Creates payment preference (Pix/boleto)
// - Returns checkout URL and QR code

func (g *Gateway) GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)
// - Calls payment.Get(paymentID)
// - Maps MP status to domain PaymentStatus

func (g *Gateway) HandleWebhook(ctx context.Context, payload []byte) (*WebhookEvent, error)
// - Parses MP webhook payload
// - Extracts action, paymentID, status
// - Returns domain WebhookEvent
```

**Important Mercado Pago knowledge:**
- Pix: uses `payment` client with `payment_method_id = "pix"`
- Boleto: uses `payment` client with `payment_method_id = "bolbradesco"` (1-3 day settlement)
- Card: requires frontend card tokenization or can use checkout preference
- Webhooks: configured in MP dashboard, sent to our `/webhooks/mercadopago` endpoint
- **Access token**: Production vs test — differentiate via env

#### 2. Stripe — Future

**Status:** 📋 Future (Phase 2)

| Technology | Stripe Go SDK |
|-----------|--------------|
| Module | `github.com/stripe/stripe-go/v85` |
| Context7 | `/stripe/stripe-go` |
| Status | 📋 Future |
| Methods | Card, international |
| Currency | USD, EUR |

---

## Port: Repository

**File:** `core/domain/repository.go`

### Interfaces

```go
package domain

type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    FindByCustomer(ctx context.Context, customerID CustomerID) ([]*Order, error)
    FindPendingByCustomer(ctx context.Context, customerID CustomerID) (*Order, error)
    UpdateStatus(ctx context.Context, id OrderID, status OrderStatus) error
}

type ProductRepository interface {
    Save(ctx context.Context, product *Product) error
    FindByID(ctx context.Context, id ProductID) (*Product, error)
    FindAll(ctx context.Context) (ProductCollection, error)
    FindByCategory(ctx context.Context, category string) (ProductCollection, error)
    UpdateAvailability(ctx context.Context, id ProductID, available bool) error
}

type CartRepository interface {
    Save(ctx context.Context, cart *Cart) error
    FindByCustomer(ctx context.Context, customerID CustomerID) (*Cart, error)
    Delete(ctx context.Context, id CartID) error
}

type CustomerRepository interface {
    Save(ctx context.Context, customer *Customer) error
    FindByID(ctx context.Context, id CustomerID) (*Customer, error)
    FindByPhone(ctx context.Context, phone PhoneNumber) (*Customer, error)
}
```

### Adapters

#### 1. SQLite (MVP — Dev)

**Package:** `core/internal/sqlite/`

| Technology | go-sqlite3 |
|-----------|-----------|
| Module | `github.com/mattn/go-sqlite3` |
| Context7 | `/mattn/go-sqlite3` |
| Status | ✅ MVP (dev) |
| Build | Requires CGO_ENABLED=1 + gcc |

**Implementation approach:**
```go
type OrderRepository struct {
    db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository
// - Expects DB already opened and migrated
// - All methods use database/sql directly (no ORM)

func (r *OrderRepository) Save(ctx context.Context, order *Order) error
// - INSERT OR REPLACE into orders table
// - INSERT items into order_items table
// - Uses transaction

func (r *OrderRepository) FindByID(ctx context.Context, id OrderID) (*Order, error)
// - SELECT from orders JOIN order_items
// - Reconstructs Order aggregate
```

**Schema outline:**
```sql
CREATE TABLE orders (
    id           TEXT PRIMARY KEY,
    customer_id  TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    total_amount INTEGER NOT NULL,  -- in cents
    currency     TEXT NOT NULL DEFAULT 'BRL',
    payment_id   TEXT,
    payment_link TEXT,
    qr_code      TEXT,
    boleto_url   TEXT,
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    confirmed_at TEXT,
    cancelled_at TEXT
);

CREATE TABLE order_items (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id   TEXT NOT NULL REFERENCES orders(id),
    product_id TEXT NOT NULL,
    name       TEXT NOT NULL,
    quantity   INTEGER NOT NULL,
    unit_price INTEGER NOT NULL,  -- in cents
    total      INTEGER NOT NULL   -- in cents
);
```

#### 2. PostgreSQL / pgx — Future

**Status:** 📋 Future (Phase 2)

| Technology | pgx v5 |
|-----------|-------|
| Module | `github.com/jackc/pgx/v5` |
| Context7 | `/jackc/pgx` |
| Status | 📋 Future |
| Reason | Higher performance, native pgx interface + database/sql compatibility |

> **Note on lib/pq vs pgx:** Context7 confirmed both. We choose pgx for production because it's actively maintained, has higher benchmark (88.25 vs 78), better performance, and supports native pgx interface as well as `database/sql` compatibility.

---

## Port: EventBus (Future)

For Phase 2+ — decoupled event propagation.

```go
package domain

type EventBus interface {
    Publish(ctx context.Context, event DomainEvent) error
    Subscribe(eventType string, handler EventHandler) error
}

type DomainEvent interface {
    EventType() string
    OccurredAt() time.Time
}

type EventHandler func(ctx context.Context, event DomainEvent) error
```

---

## Related

- [[wiki/specs/ARCHITECTURE]] — How layers connect
- [[wiki/specs/DOMAIN]] — What types appear in these interfaces
- [[wiki/specs/MVP-SCOPE]] — Which adapters are built first

## References

- [Context7: whatsmeow](https://context7.com/tulir/whatsmeow) — WhatsApp multi-device lib
- [Context7: Mercado Pago SDK Go](https://context7.com/mercadopago/sdk-go) — Payment gateway
- [Context7: Stripe Go](https://context7.com/stripe/stripe-go) — International payments
- [Context7: go-sqlite3](https://context7.com/mattn/go-sqlite3) — SQLite driver
- [Context7: pgx](https://context7.com/jackc/pgx) — PostgreSQL driver
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
