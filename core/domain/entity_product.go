package domain

// ---------------------------------------------------------------------------
// Product — Entity
// Rule 8 compliance: Product composes exactly 2 instance variables.
// ---------------------------------------------------------------------------

// productDetails groups all descriptive and commercial attributes of a product.
type productDetails struct {
	name        string
	description string
	price       Money
	category    string
	available   bool
	imageURL    string
}

// Product represents an item available for purchase.
type Product struct {
	id      ProductID
	details productDetails
}

// NewProduct creates a new Product with full attributes.
func NewProduct(id ProductID, name, description string, price Money, category string, available bool, imageURL string) *Product {
	return &Product{
		id: id,
		details: productDetails{
			name:        name,
			description: description,
			price:       price,
			category:    category,
			available:   available,
			imageURL:    imageURL,
		},
	}
}

// UpdateAvailability marks the product as available or unavailable.
func (p *Product) UpdateAvailability(available bool) {
	p.details.available = available
}

// ---------------------------------------------------------------------------
// Read accessors
// ---------------------------------------------------------------------------

// ID returns the product identifier.
func (p *Product) ID() ProductID { return p.id }

// Name returns the product name.
func (p *Product) Name() string { return p.details.name }

// Description returns the product description.
func (p *Product) Description() string { return p.details.description }

// Price returns the current price.
func (p *Product) Price() Money { return p.details.price }

// Category returns the product category.
func (p *Product) Category() string { return p.details.category }

// IsAvailable returns whether the product can be purchased.
func (p *Product) IsAvailable() bool { return p.details.available }

// ImageURL returns the product image URL.
func (p *Product) ImageURL() string { return p.details.imageURL }
