package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// SQLite adapter for CartRepository
// Rule 8: 1 instance variable (db).
// ---------------------------------------------------------------------------

// CartRepository implements domain.CartRepository using SQLite.
type CartRepository struct {
	db *sql.DB
}

// NewCartRepository creates a new SQLite-backed CartRepository.
func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

// Save persists a cart and its items within a transaction.
func (r *CartRepository) Save(ctx context.Context, cart *domain.Cart) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO carts (id, customer_id, expires_at)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			customer_id=excluded.customer_id, expires_at=excluded.expires_at`,
		cart.ID().Value(), cart.CustomerID().Value(),
		cart.ExpiresAt().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("upsert cart: %w", err)
	}

	// Replace cart items: delete old, insert new
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = ?`, cart.ID().Value()); err != nil {
		return fmt.Errorf("delete items: %w", err)
	}
		for _, item := range cart.Items().Items() {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO cart_items (cart_id, product_id, name, unit_price, currency, quantity)
				VALUES (?, ?, ?, ?, ?, ?)`,
				cart.ID().Value(), item.ProductID().Value(), item.Name(),
				item.UnitPrice().Amount(), item.UnitPrice().Currency(), item.Quantity()); err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}
	return tx.Commit()
}

// FindByCustomer retrieves the customer's current cart.
func (r *CartRepository) FindByCustomer(ctx context.Context, customerID domain.CustomerID) (*domain.Cart, error) {
	var id, expiresAtStr string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, expires_at FROM carts WHERE customer_id = ?`, customerID.Value()).
		Scan(&id, &expiresAtStr)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCartNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find cart: %w", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	// Load items
	items, err := r.scanCartItems(ctx, id)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructCart(
		domain.NewCartID(id),
		customerID,
		domain.NewCartItemCollection(items),
		expiresAt,
	), nil
}

// Delete removes a cart and its items (ON DELETE CASCADE handles items).
func (r *CartRepository) Delete(ctx context.Context, id domain.CartID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM carts WHERE id = ?`, id.Value())
	if err != nil {
		return fmt.Errorf("delete cart: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.ErrCartNotFound
	}
	return nil
}

func (r *CartRepository) scanCartItems(ctx context.Context, cartID string) ([]domain.CartItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT product_id, name, unit_price, currency, quantity FROM cart_items WHERE cart_id = ?`, cartID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []domain.CartItem
	for rows.Next() {
		var productID, name, currency string
		var unitPrice, quantity int
		if err := rows.Scan(&productID, &name, &unitPrice, &currency, &quantity); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		price, err := domain.NewMoney(int64(unitPrice), currency)
		if err != nil {
			return nil, fmt.Errorf("invalid item price: %w", err)
		}
		item, err := domain.NewCartItem(domain.NewProductID(productID), name, price, quantity)
		if err != nil {
			return nil, fmt.Errorf("reconstruct item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
