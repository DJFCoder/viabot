// Package sdk provides public factory functions for the ViaBot Core SDK.
//
// It bridges the internal adapter implementations (whatsmeow, mercadopago, sqlite)
// to external entry points (customers/*) while keeping adapters hidden behind
// the domain interface (Hexagonal Architecture / DIP).
//
// Rule 6: No abbreviations — "sdk" is the well-known acronym from AGENTS.md.
package sdk

import (
	"database/sql"

	"viabot.stream/sdk/domain"
	"viabot.stream/sdk/internal/sqlite"
)

// ---------------------------------------------------------------------------
// Database
// ---------------------------------------------------------------------------

// OpenDatabase opens a SQLite connection using the provided DSN.
func OpenDatabase(dsn string) (*sql.DB, error) {
	return sqlite.OpenDB(dsn)
}

// RunMigrations executes the schema DDL.
func RunMigrations(database *sql.DB) error {
	return sqlite.Migrate(database)
}

// DefaultDatabaseDSN returns the recommended SQLite DSN for development.
func DefaultDatabaseDSN() string {
	return sqlite.DefaultDSN
}

// ---------------------------------------------------------------------------
// Repository factories
//
// Each factory returns the domain interface (port) instead of the concrete
// adapter type — callers depend on abstractions, never on implementations (DIP).
// ---------------------------------------------------------------------------

// NewCustomerRepository creates a SQLite-backed CustomerRepository.
func NewCustomerRepository(database *sql.DB) domain.CustomerRepository {
	return sqlite.NewCustomerRepository(database)
}

// NewProductRepository creates a SQLite-backed ProductRepository.
func NewProductRepository(database *sql.DB) domain.ProductRepository {
	return sqlite.NewProductRepository(database)
}

// NewCartRepository creates a SQLite-backed CartRepository.
func NewCartRepository(database *sql.DB) domain.CartRepository {
	return sqlite.NewCartRepository(database)
}

// NewOrderRepository creates a SQLite-backed OrderRepository.
func NewOrderRepository(database *sql.DB) domain.OrderRepository {
	return sqlite.NewOrderRepository(database)
}
