package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// SQLite adapter for CustomerRepository
// Rule 8: 1 instance variable (db).
// ---------------------------------------------------------------------------

// CustomerRepository implements domain.CustomerRepository using SQLite.
type CustomerRepository struct {
	db *sql.DB
}

// NewCustomerRepository creates a new SQLite-backed CustomerRepository.
func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Save inserts or updates a customer.
func (r *CustomerRepository) Save(ctx context.Context, customer *domain.Customer) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO customers (id, phone_country, phone_number, name, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			phone_country=excluded.phone_country,
			phone_number=excluded.phone_number,
			name=excluded.name`,
		customer.ID().Value(),
		customer.Phone().CountryCode(),
		customer.Phone().Number(),
		customer.Name(),
		customer.CreatedAt().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save customer: %w", err)
	}
	return nil
}

// FindByID retrieves a customer by their ID.
func (r *CustomerRepository) FindByID(ctx context.Context, id domain.CustomerID) (*domain.Customer, error) {
	return r.scanCustomer(ctx, `SELECT id, phone_country, phone_number, name, created_at
		FROM customers WHERE id = ?`, id.Value())
}

// FindByPhone retrieves a customer by phone number.
func (r *CustomerRepository) FindByPhone(ctx context.Context, phone domain.PhoneNumber) (*domain.Customer, error) {
	return r.scanCustomer(ctx, `SELECT id, phone_country, phone_number, name, created_at
		FROM customers WHERE phone_country = ? AND phone_number = ?`,
		phone.CountryCode(), phone.Number())
}

func (r *CustomerRepository) scanCustomer(ctx context.Context, query string, args ...interface{}) (*domain.Customer, error) {
	var id, name, phoneNumber, createdAtStr string
	var phoneCountry int

	err := r.db.QueryRowContext(ctx, query, args...).
		Scan(&id, &phoneCountry, &phoneNumber, &name, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCustomerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find customer: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
	phone := domain.NewPhoneNumber(phoneCountry, phoneNumber)

	return domain.ReconstructCustomer(
		domain.NewCustomerID(id), phone, name, createdAt,
	), nil
}
