# Customers — Entry Points

Each subdirectory is a standalone Go entry point for a specific Customer/tenant.

## Pattern

```markdown
customers/
├── customer-a/
│   ├── go.mod
│   ├── main.go          # Wires adapters, starts the bot
│   └── config.go        # Environment-specific config
└── customer-b/
    ├── go.mod
    ├── main.go
    └── config.go
```

Customer import the Core SDK and customize:

- Messenger adapter (WhatsApp, Telegram, etc.)
- Payment gateway (Mercado Pago, Stripe)
- Repository (SQLite, pgx)
- Product catalog
- Business rules / prompts

## Adding a New Customer

1. Create directory `customers/<customer-name>/`
2. Initialize module: `go mod init github.com/yourorg/<customer-name>`
3. Import `viabot.stream/sdk`
4. Wire dependencies in `main.go`
5. Add `./customers/<customer-name>` to `go.work`
6. Run `go work sync`
