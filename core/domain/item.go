package domain

// ---------------------------------------------------------------------------
// OrderItem — a product line inside the Order aggregate
// ---------------------------------------------------------------------------

// OrderItem represents a single product entry within an Order.
// It is a value object inside the Order aggregate boundary.
type OrderItem struct {
	productID ProductID
	name      string
	unitPrice Money
	quantity  int
}

// NewOrderItem creates an OrderItem with validation.
func NewOrderItem(productID ProductID, name string, unitPrice Money, quantity int) (OrderItem, error) {
	if quantity <= 0 {
		return OrderItem{}, ErrInvalidQuantity
	}
	return OrderItem{
		productID: productID,
		name:      name,
		unitPrice: unitPrice,
		quantity:  quantity,
	}, nil
}

// SubTotal returns the line total (unitPrice × quantity).
func (i OrderItem) SubTotal() Money {
	return i.unitPrice.Multiply(i.quantity)
}

// ProductID returns the product identifier.
func (i OrderItem) ProductID() ProductID { return i.productID }

// Name returns the product name at time of ordering.
func (i OrderItem) Name() string { return i.name }

// UnitPrice returns the price per unit.
func (i OrderItem) UnitPrice() Money { return i.unitPrice }

// Quantity returns the number of units ordered.
func (i OrderItem) Quantity() int { return i.quantity }

// ---------------------------------------------------------------------------
// CartItem — a product entry inside the Cart aggregate
// ---------------------------------------------------------------------------

// CartItem represents a product the customer intends to purchase.
// Structurally identical to OrderItem but belongs to a different aggregate.
type CartItem struct {
	productID ProductID
	name      string
	unitPrice Money
	quantity  int
}

// NewCartItem creates a CartItem with validation.
func NewCartItem(productID ProductID, name string, unitPrice Money, quantity int) (CartItem, error) {
	if quantity <= 0 {
		return CartItem{}, ErrInvalidQuantity
	}
	return CartItem{
		productID: productID,
		name:      name,
		unitPrice: unitPrice,
		quantity:  quantity,
	}, nil
}

// SubTotal returns the line total (unitPrice × quantity).
func (i CartItem) SubTotal() Money {
	return i.unitPrice.Multiply(i.quantity)
}

// ProductID returns the product identifier.
func (i CartItem) ProductID() ProductID { return i.productID }

// Name returns the product name.
func (i CartItem) Name() string { return i.name }

// UnitPrice returns the price per unit.
func (i CartItem) UnitPrice() Money { return i.unitPrice }

// Quantity returns the number of units.
func (i CartItem) Quantity() int { return i.quantity }
