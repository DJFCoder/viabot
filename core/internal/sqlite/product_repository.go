package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// SQLite adapter for ProductRepository
// Rule 8: 1 instance variable (db).
// ---------------------------------------------------------------------------

// ProductRepository implements domain.ProductRepository using SQLite.
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new SQLite-backed ProductRepository.
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Save inserts or replaces a product.
func (r *ProductRepository) Save(ctx context.Context, product *domain.Product) error {
	available := 0
	if product.IsAvailable() {
		available = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO products (id, name, description, price_amount, currency, category, available, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, description=excluded.description,
			price_amount=excluded.price_amount, currency=excluded.currency,
			category=excluded.category, available=excluded.available,
			image_url=excluded.image_url`,
		product.ID().Value(), product.Name(), product.Description(),
		product.Price().Amount(), product.Price().Currency(),
		product.Category(), available, product.ImageURL())
	if err != nil {
		return fmt.Errorf("save product: %w", err)
	}
	return nil
}

// FindByID retrieves a product by ID.
func (r *ProductRepository) FindByID(ctx context.Context, id domain.ProductID) (*domain.Product, error) {
	var name, desc, currency, category, imageURL string
	var priceAmount int64
	var available int

	err := r.db.QueryRowContext(ctx,
		`SELECT name, description, price_amount, currency, category, available, image_url
		 FROM products WHERE id = ?`, id.Value()).
		Scan(&name, &desc, &priceAmount, &currency, &category, &available, &imageURL)
	if err == sql.ErrNoRows {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find product: %w", err)
	}

	price, _ := domain.NewMoney(priceAmount, currency)
	return domain.ReconstructProduct(
		id, name, desc, price, category, available == 1, imageURL,
	), nil
}

// FindAll returns all products.
func (r *ProductRepository) FindAll(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, price_amount, currency, category, available, image_url
		 FROM products ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("find all: %w", err)
	}
	defer rows.Close()
	return r.scanProducts(rows)
}

// FindByCategory returns products matching the given category.
func (r *ProductRepository) FindByCategory(ctx context.Context, category string) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, price_amount, currency, category, available, image_url
		 FROM products WHERE category = ? ORDER BY name`, category)
	if err != nil {
		return nil, fmt.Errorf("find by category: %w", err)
	}
	defer rows.Close()
	return r.scanProducts(rows)
}

// UpdateAvailability sets the available flag for a product.
func (r *ProductRepository) UpdateAvailability(ctx context.Context, id domain.ProductID, available bool) error {
	val := 0
	if available {
		val = 1
	}
	result, err := r.db.ExecContext(ctx,
		`UPDATE products SET available = ? WHERE id = ?`, val, id.Value())
	if err != nil {
		return fmt.Errorf("update availability: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) scanProducts(rows *sql.Rows) ([]domain.Product, error) {
	var products []domain.Product
	for rows.Next() {
		var id, name, desc, currency, category, imageURL string
		var priceAmount int64
		var available int
		if err := rows.Scan(&id, &name, &desc, &priceAmount, &currency, &category, &available, &imageURL); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		price, _ := domain.NewMoney(priceAmount, currency)
		product := domain.ReconstructProduct(
			domain.NewProductID(id), name, desc, price, category, available == 1, imageURL,
		)
		products = append(products, *product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return products, nil
}
