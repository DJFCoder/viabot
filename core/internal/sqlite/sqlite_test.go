package sqlite

import (
	"database/sql"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Shared test helpers for SQLite repository tests
// All tests use :memory: database for isolation.
// ---------------------------------------------------------------------------

const testDatabaseDSN = "file::memory:?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=30000&cache=shared&_loc=UTC"

// openTestDatabase opens an in-memory SQLite database with migrations applied.
func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	database, err := OpenDB(testDatabaseDSN)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	if err := Migrate(database); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	// Allow time for WAL setup
	time.Sleep(10 * time.Millisecond)
	return database
}
