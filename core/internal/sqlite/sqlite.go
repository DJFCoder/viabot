// Package sqlite implements domain repository interfaces using SQLite.
// Requires CGO_ENABLED=1 and gcc (mattn/go-sqlite3 is a CGO library).
package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// SQLite database setup — shared across all repository implementations.
// ---------------------------------------------------------------------------

const (
	// DefaultDSN is the recommended SQLite DSN for ViaBot.
	// WAL mode for concurrent reads, foreign keys for referential integrity,
	// 30s busy timeout to avoid "database is locked" errors.
	DefaultDSN = "file:./data/bot.db?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=30000&cache=shared&_loc=UTC"
)

// Migrate creates all required tables if they do not exist.
// Idempotent — safe to call on every startup.
func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS customers (
		id              TEXT PRIMARY KEY,
		phone_country   INTEGER NOT NULL DEFAULT 55,
		phone_number    TEXT NOT NULL,
		name            TEXT NOT NULL DEFAULT '',
		created_at      TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS products (
		id              TEXT PRIMARY KEY,
		name            TEXT NOT NULL,
		description     TEXT NOT NULL DEFAULT '',
		price_amount    INTEGER NOT NULL,
		currency        TEXT NOT NULL DEFAULT 'BRL',
		category        TEXT NOT NULL DEFAULT '',
		available       INTEGER NOT NULL DEFAULT 1,
		image_url       TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS carts (
		id              TEXT PRIMARY KEY,
		customer_id     TEXT NOT NULL,
		expires_at      TEXT NOT NULL,
		FOREIGN KEY (customer_id) REFERENCES customers(id)
	);

	CREATE TABLE IF NOT EXISTS cart_items (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		cart_id         TEXT NOT NULL,
		product_id      TEXT NOT NULL,
		name            TEXT NOT NULL,
		unit_price      INTEGER NOT NULL,
		currency        TEXT NOT NULL DEFAULT 'BRL',
		quantity        INTEGER NOT NULL,
		FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS orders (
		id              TEXT PRIMARY KEY,
		customer_id     TEXT NOT NULL,
		status          TEXT NOT NULL DEFAULT 'pending',
		total_amount    INTEGER NOT NULL,
		currency        TEXT NOT NULL DEFAULT 'BRL',
		payment_id      TEXT,
		payment_status  TEXT,
		payment_method  TEXT,
		paid_amount     INTEGER NOT NULL DEFAULT 0,
		checkout_link   TEXT,
		qr_code         TEXT,
		qr_code_text    TEXT,
		boleto_barcode  TEXT,
		boleto_url      TEXT,
		payment_expires_at TEXT,
		payment_created_at TEXT,
		payment_confirmed_at TEXT,
		created_at      TEXT NOT NULL DEFAULT (datetime('now')),
		confirmed_at    TEXT,
		FOREIGN KEY (customer_id) REFERENCES customers(id)
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id        TEXT NOT NULL,
		product_id      TEXT NOT NULL,
		name            TEXT NOT NULL,
		quantity        INTEGER NOT NULL,
		unit_price      INTEGER NOT NULL,
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
	CREATE INDEX IF NOT EXISTS idx_carts_customer ON carts(customer_id);
	CREATE INDEX IF NOT EXISTS idx_cart_items_cart ON cart_items(cart_id);
	CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
	CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// OpenDB opens a SQLite database, enables WAL mode, and runs migrations.
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	return db, nil
}
