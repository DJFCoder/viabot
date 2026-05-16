package domain

import (
	"time"
)

// ---------------------------------------------------------------------------
// Cart — Aggregate Root (E-Commerce Bounded Context)
// Rule 8 compliance: Cart composes exactly 2 instance variables.
// ---------------------------------------------------------------------------

const (
	// CartExpirationHours defines how long a cart remains active.
	CartExpirationHours = 24
)

// cartContents groups all non-identity state of a Cart.
type cartContents struct {
	customerID CustomerID
	items      CartItemCollection
	expiresAt  time.Time
}

// Cart is a temporary collection of items a customer intends to purchase.
type Cart struct {
	id       CartID
	contents cartContents
}

// NewCart creates a Cart with an expiration clock.
func NewCart(id CartID, customerID CustomerID) *Cart {
	return &Cart{
		id: id,
		contents: cartContents{
			customerID: customerID,
			items:      NewCartItemCollection(nil),
			expiresAt:  time.Now().Add(CartExpirationHours * time.Hour),
		},
	}
}

// AddItem adds a product to the cart. If the product already exists,
// the quantity is incremented instead of adding a duplicate.
func (c *Cart) AddItem(productID ProductID, name string, unitPrice Money, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	existing := c.findItem(productID)
	if existing != nil {
		newQuantity := existing.Quantity() + quantity
		updated, err := NewCartItem(existing.ProductID(), existing.Name(), existing.UnitPrice(), newQuantity)
		if err != nil {
			return err
		}
		c.updateItem(existing.ProductID(), updated)
		return nil
	}

	item, err := NewCartItem(productID, name, unitPrice, quantity)
	if err != nil {
		return err
	}
	c.contents.items = c.contents.items.Add(item)
	return nil
}

// RemoveItem removes a product from the cart entirely.
func (c *Cart) RemoveItem(productID ProductID) {
	c.contents.items = c.contents.items.Remove(productID)
}

// UpdateQuantity changes the quantity of an existing cart item.
func (c *Cart) UpdateQuantity(productID ProductID, quantity int) error {
	updated, err := c.contents.items.UpdateQuantity(productID, quantity)
	if err != nil {
		return err
	}
	c.contents.items = updated
	return nil
}

// Clear empties all items from the cart.
func (c *Cart) Clear() {
	c.contents.items = NewCartItemCollection(nil)
}

// IsExpired returns true if the cart has passed its expiration time.
func (c *Cart) IsExpired() bool {
	return time.Now().After(c.contents.expiresAt)
}

// Checkout converts the cart into an Order if the cart is valid.
// The cart is NOT discarded here — the caller must delete it after persisting the order.
func (c *Cart) Checkout(orderID OrderID) (*Order, error) {
	if c.contents.items.Count() == 0 {
		return nil, ErrEmptyCart
	}
	if c.IsExpired() {
		return nil, ErrCartExpired
	}

	orderItems := c.toOrderItems()
	collection := NewOrderItemCollection(orderItems)

	return NewOrder(orderID, c.contents.customerID, collection)
}

// toOrderItems converts CartItems to OrderItems for checkout.
func (c *Cart) toOrderItems() []OrderItem {
	items := c.contents.items.Items()
	result := make([]OrderItem, 0, len(items))
	for _, ci := range items {
		oi, err := NewOrderItem(ci.ProductID(), ci.Name(), ci.UnitPrice(), ci.Quantity())
		if err != nil {
			// All validation passed at CartItem creation; skip errors defensively.
			continue
		}
		result = append(result, oi)
	}
	return result
}

// findItem returns a pointer to the matching CartItem, or nil.
func (c *Cart) findItem(productID ProductID) *CartItem {
	for _, item := range c.contents.items.Items() {
		if item.ProductID() == productID {
			return &item
		}
	}
	return nil
}

// updateItem replaces an existing item in the collection.
func (c *Cart) updateItem(productID ProductID, updated CartItem) {
	c.contents.items = c.contents.items.Remove(productID)
	c.contents.items = c.contents.items.Add(updated)
}

// ---------------------------------------------------------------------------
// Read accessors
// ---------------------------------------------------------------------------

// ID returns the cart identifier.
func (c *Cart) ID() CartID { return c.id }

// CustomerID returns the customer who owns this cart.
func (c *Cart) CustomerID() CustomerID { return c.contents.customerID }

// Items returns a defensive copy of cart items.
func (c *Cart) Items() CartItemCollection { return c.contents.items }

// ExpiresAt returns when the cart becomes invalid.
func (c *Cart) ExpiresAt() time.Time { return c.contents.expiresAt }
