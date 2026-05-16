# Customers — Entry Points

Each subdirectory is a standalone Go entry point for a specific client/tenant.

## Pattern

```markdown
customers/
├── client-a/
│   ├── go.mod
│   ├── main.go          # Wires adapters, starts the bot
│   └── config.go        # Environment-specific config
└── client-b/
    ├── go.mod
    ├── main.go
    └── config.go
```

Clients import the Core SDK and customize:

- Messenger adapter (WhatsApp, Telegram, etc.)
- Payment gateway (Mercado Pago, Stripe)
- Repository (SQLite, pgx)
- Product catalog
- Business rules / prompts

## Adding a New Client

1. Create directory `customers/<client-name>/`
2. Initialize module: `go mod init github.com/yourorg/<client-name>`
3. Import `viabot.stream/sdk`
4. Wire dependencies in `main.go`
5. Add `./customers/<client-name>` to `go.work`
6. Run `go work sync`
