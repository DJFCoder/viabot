# ViaBot — Workspace AGENTS.md

> Sessão aberta em `whatsapp-bot/`. Leia antes de qualquer ação.
> Resposta ao usuário **sempre em pt-br**. Código, comentários, docs, commits **sempre em inglês**.

## ⛔ COMMITS ARE FORBIDDEN — HUMAN ONLY

**NENHUM agent, subagent ou skill pode executar commits, pushes ou merges. Apenas humanos commitam. Esta regra é inegociável e aplica-se a TODOS os agentes do projeto.**

```markdown
PROIBIDO: git commit, git push, git merge, gh pr merge, gh pr create
PERMITIDO: git status, git diff, git log, git branch, gh pr view, gh pr list
```

---

## 🧠 Obsidian LLM Wiki — Knowledge Base

> **Vault**: `obsidian-vault/` (abra no Obsidian para navegação visual com Graph View)
> **Schema**: `obsidian-vault/AGENTS.md` (leia com Read tool ao trabalhar com conhecimento)

**Regra #1 — SEMPRE atualizar o vault.** Toda operação que produza conhecimento novo DEVE registrar no vault:

1. **Atualizar `wiki/log.md`** — entrada cronológica
2. **Criar/atualizar páginas de conceito** em `wiki/concepts/`
3. **Criar/atualizar decisões** em `wiki/decisions/`
4. **Criar/atualizar entidades** em `wiki/entities/`

---

## 1. Estrutura do Projeto

```markdown
whatsapp-bot/
├── core/                     # Core SDK — public Go package (Hexagonal Architecture)
│   ├── go.mod
│   ├── domain/               # Ports (interfaces) — Messenger, PaymentGateway, Repository
│   └── internal/             # Adapters (implementações) — whatsmeow, mercadopago, sqlite
├── customers/                # Entry points — per-client customizations (future multi-repo)
├── admin-ui/                 # React admin SPA (future — not in MVP)
├── obsidian-vault/           # Vault Obsidian com wiki de conhecimento
│   └── wiki/
│       ├── specs/            # Design documents (ARCHITECTURE, DOMAIN, MVP)
│       ├── concepts/
│       ├── decisions/
│       └── entities/
├── .opencode/                # OpenCode config
│   ├── instructions/         # Regras de qualidade e arquitetura
│   └── rag/                  # Pipeline RAG (BM25 + semântico)
├── opencode.json             # Config OpenCode
├── go.work                   # Go workspace (core + customers/*/)
└── AGENTS.md                 # This file
```

### Arquitetura: Hexagonal (Ports & Adapters) + DDD

```markdown
┌──────────────────────────────────────────────┐
│                  Messenger                    │
│  (WhatsApp / Telegram / Email / SMS)         │
└─────────────────┬────────────────────────────┘
                  │
┌─────────────────▼────────────────────────────┐
│            Application Layer                  │
│  Use Cases / Handlers / Services              │
└────┬──────────────────────┬──────────────────┘
     │                      │
┌────▼──────────┐   ┌──────▼───────────┐
│   Domain      │   │  Infrastructure  │
│  Entities     │   │  Repositories    │
│  Value Objects│   │  PaymentGateway  │
│  Aggregates   │   │  Messenger       │
└───────────────┘   └──────────────────┘
```

### Core SDK Pattern

- `core/` — public Go package importado via `viabot.stream/sdk`
- `customers/*/` — per-client entry points com customizações
- `go.work` resolve `core/` localmente no monorepo MVP
- Futuro: split para repositórios separados com `git submodule`

---

## 2. Stack Tecnológica (Verificada via Context7)

| Camada | Tecnologia | Module Path | Context7 ID |
| ------ | ---------- | ----------- | ----------- |
| **Runtime** | Go 1.22+ | — | — |
| **WhatsApp** | whatsmeow | `go.mau.fi/whatsmeow` | `/tulir/whatsmeow` |
| **Payments (MVP)** | Mercado Pago SDK | `github.com/mercadopago/sdk-go` | `/mercadopago/sdk-go` |
| **Payments (future)** | Stripe Go SDK | `github.com/stripe/stripe-go/v85` | `/stripe/stripe-go` |
| **DB (dev)** | SQLite (mattn) | `github.com/mattn/go-sqlite3` | `/mattn/go-sqlite3` |
| **DB (prod)** | pgx v5 | `github.com/jackc/pgx/v5` | `/jackc/pgx` |
| **RAG** | Python 3.10+ | — | — |
| **Embeddings** | fastembed (all-MiniLM-L6-v2) | — | — |

> **Regra:** Toda dependência externa deve ser verificada via Context7 antes de incluir no `go.mod` ou referenciar em docs.

---

## 3. RAG Pipeline (`.opencode/rag/`)

Pipeline híbrido BM25 + semântico para agents. Usa modelos ONNX locais:

- **Indexador** (`index.py`) — chunk → embed (all-MiniLM-L6-v2) → SQLite + FTS5
- **Retriever** (`query.py`) — BM25 + semantic search + score fusion → top-5 passagens
- **Auditoria** (`audit.py`) — relatório de qualidade do índice

**Corpus:** `obsidian-vault/wiki/`, `.opencode/instructions/`, `core/` (future)

---

## 4. Regras de Código

### Object Calisthenics (Obrigatório para código com comportamento)

| # | Regra | Prática |
| - | ----- | ------- |
| 1 | One level of indentation per method | Extrair helper methods |
| 2 | **No `else` keyword** — ever | Early returns / guard clauses |
| 3 | Wrap primitives (Value Objects) | `Money{amount, currency}`, `PhoneNumber`, `OrderStatus` |
| 4 | First class collections | `ProductCollection` em vez de `[]Product` |
| 5 | One dot per line (Law of Demeter) | Delegar ao objeto intermediário |
| 6 | No abbreviations | `CalculateTotal`, não `CalcTot` |
| 7 | Small entities | ≤50 linhas/classe, ≤20 linhas/método, ≤10 métodos/classe |
| 8 | Max 2 instance variables | Usar composição (exceto logger) |
| 9 | No getters/setters in domain | Private constructor + static factory |

### SOLID

| Princípio | Aplicação no Bot |
| --------- | ---------------- |
| **SRP** | 1 interface = 1 responsabilidade (Messenger, Payment, Repository) |
| **OCP** | Novo canal = novo adapter, sem modificar domínio |
| **LSP** | Todo adapter de Messenger deve ser substituível |
| **ISP** | Interfaces pequenas: `MessengerSender`, `MessengerReceiver` |
| **DIP** | Depender de abstrações, injetar concreções via construtor |

---

## 5. Domínios Planejados (Fases Futuras)

| Domínio | MVP | Fase 2 | Fase 3+ |
| ------- | --- | ------ | ------- |
| **E-commerce** | Catálogo + carrinho + Pix/Boleto/Card | Entregas, tracking, estoque | Marketplace |
| **Finance** | — | Extrato, split payment | Assinaturas, recorrente |
| **Mentor IA** | — | FAQ automático | Chat IA contextual |

---

## 6. Multi-Channel Architecture

```go
// Messenger interface — single abstraction for all channels
type Messenger interface {
    Send(ctx context.Context, to Recipient, msg *Message) error
    Receive(ctx context.Context, handler MessageHandler) error
}
```

| Channel | Adapter | Status |
| ------- | ------- | ------ |
| WhatsApp | whatsmeow (go.mau.fi) | ✅ MVP |
| Telegram | telegram-bot-api | 📋 Future |
| Email | SMTP / SendGrid | 📋 Future |
| SMS | Twilio / AWS SNS | 📋 Future |

---

## 7. Multi-Gateway Payments

```go
// PaymentGateway interface — single abstraction for all gateways
type PaymentGateway interface {
    CreatePayment(ctx context.Context, order *Order) (*Payment, error)
    GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)
    WebhookHandler(ctx context.Context, payload []byte) (*WebhookEvent, error)
}
```

| Gateway | Métodos | Moeda | Status |
| ------- | ------- | ----- | ------ |
| Mercado Pago | Pix (0.99%), Boleto (R$1.99), Card (3.99%) | BRL | ✅ MVP |
| Stripe | Card, international | USD, EUR | 📋 Future |

---

## 8. Repository Pattern (SQLite ↔ PostgreSQL)

```go
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    FindByCustomer(ctx context.Context, customerID CustomerID) ([]*Order, error)
}
```

| Driver | Uso | Module |
| ------ | --- | ------ |
| SQLite (mattn) | Dev, testes unitários | `github.com/mattn/go-sqlite3` |
| pgx v5 | Produção | `github.com/jackc/pgx/v5` |

---

## 9. Controle de Sessão e Estados (BDD-style)

```markdown
Feature: Customer checks out via WhatsApp
  Scenario: Complete purchase flow
    Given a customer with items in cart
    When they request checkout
    Then they receive a payment link
    And the order status is "awaiting_payment"
    When payment is confirmed
    Then the order status is "confirmed"
    And the customer receives a receipt
```

---

**Last Updated:** 2026-05-16
