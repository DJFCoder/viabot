package domain

// ---------------------------------------------------------------------------
// First-Class Collections (Rule 4)
// Each collection type wraps a slice and exposes behaviour.
// ---------------------------------------------------------------------------

// OrderItemCollection wraps a slice of OrderItem with aggregate behaviour.
// It is the ONLY attribute on this class (Rule 4 compliance).
type OrderItemCollection struct {
	items []OrderItem
}

// NewOrderItemCollection creates a collection from items.
func NewOrderItemCollection(items []OrderItem) OrderItemCollection {
	if items == nil {
		return OrderItemCollection{items: []OrderItem{}}
	}
	return OrderItemCollection{items: items}
}

// Add appends an item and returns a new collection (immutable pattern).
func (c OrderItemCollection) Add(item OrderItem) OrderItemCollection {
	newItems := make([]OrderItem, len(c.items)+1)
	copy(newItems, c.items)
	newItems[len(c.items)] = item
	return OrderItemCollection{items: newItems}
}

// Remove filters out the item matching the given productID.
func (c OrderItemCollection) Remove(productID ProductID) OrderItemCollection {
	var filtered []OrderItem
	for _, item := range c.items {
		if item.ProductID() != productID {
			filtered = append(filtered, item)
		}
	}
	return OrderItemCollection{items: filtered}
}

// Total sums all item subtotals.
func (c OrderItemCollection) Total() (Money, error) {
	var total Money
	var zeroInit bool

	for _, item := range c.items {
		sub := item.SubTotal()
		if !zeroInit {
			total = sub
			zeroInit = true
			continue
		}
		var err error
		total, err = total.Add(sub)
		if err != nil {
			return Money{}, err
		}
	}
	return total, nil
}

// Count returns the number of items.
func (c OrderItemCollection) Count() int {
	return len(c.items)
}

// Items returns a defensive copy of the inner slice.
func (c OrderItemCollection) Items() []OrderItem {
	result := make([]OrderItem, len(c.items))
	copy(result, c.items)
	return result
}

// ---------------------------------------------------------------------------
// CartItemCollection
// ---------------------------------------------------------------------------

// CartItemCollection wraps a slice of CartItem.
type CartItemCollection struct {
	items []CartItem
}

// NewCartItemCollection creates a collection from items.
func NewCartItemCollection(items []CartItem) CartItemCollection {
	if items == nil {
		return CartItemCollection{items: []CartItem{}}
	}
	return CartItemCollection{items: items}
}

// Add appends an item and returns a new collection.
func (c CartItemCollection) Add(item CartItem) CartItemCollection {
	newItems := make([]CartItem, len(c.items)+1)
	copy(newItems, c.items)
	newItems[len(c.items)] = item
	return CartItemCollection{items: newItems}
}

// Remove filters out the item matching the given productID.
func (c CartItemCollection) Remove(productID ProductID) CartItemCollection {
	var filtered []CartItem
	for _, item := range c.items {
		if item.ProductID() != productID {
			filtered = append(filtered, item)
		}
	}
	return CartItemCollection{items: filtered}
}

// UpdateQuantity replaces the quantity for a given product.
// Returns the updated collection; if the product is not found returns the
// original collection unchanged.
func (c CartItemCollection) UpdateQuantity(productID ProductID, quantity int) (CartItemCollection, error) {
	if quantity <= 0 {
		return CartItemCollection{}, ErrInvalidQuantity
	}
	updated := make([]CartItem, len(c.items))
	var found bool
	for i, item := range c.items {
		if item.ProductID() == productID {
			newItem, err := NewCartItem(item.ProductID(), item.Name(), item.UnitPrice(), quantity)
			if err != nil {
				return CartItemCollection{}, err
			}
			updated[i] = newItem
			found = true
		} else {
			updated[i] = item
		}
	}
	if !found {
		return c, nil
	}
	return CartItemCollection{items: updated}, nil
}

// Total sums all item subtotals.
func (c CartItemCollection) Total() (Money, error) {
	var total Money
	var zeroInit bool

	for _, item := range c.items {
		sub := item.SubTotal()
		if !zeroInit {
			total = sub
			zeroInit = true
			continue
		}
		var err error
		total, err = total.Add(sub)
		if err != nil {
			return Money{}, err
		}
	}
	return total, nil
}

// Count returns the number of items.
func (c CartItemCollection) Count() int {
	return len(c.items)
}

// Items returns a defensive copy of the inner slice.
func (c CartItemCollection) Items() []CartItem {
	result := make([]CartItem, len(c.items))
	copy(result, c.items)
	return result
}

// ---------------------------------------------------------------------------
// ProductCollection
// ---------------------------------------------------------------------------

// ProductCollection wraps a slice of Product with query behaviour.
type ProductCollection struct {
	items []Product
}

// NewProductCollection creates a collection from products.
func NewProductCollection(items []Product) ProductCollection {
	if items == nil {
		return ProductCollection{items: []Product{}}
	}
	return ProductCollection{items: items}
}

// FilterByCategory returns products matching the given category.
func (c ProductCollection) FilterByCategory(category string) ProductCollection {
	var filtered []Product
	for _, p := range c.items {
		if p.Category() == category {
			filtered = append(filtered, p)
		}
	}
	return ProductCollection{items: filtered}
}

// Search returns products whose name contains the term (case-insensitive).
func (c ProductCollection) Search(term string) ProductCollection {
	var matched []Product
	for _, p := range c.items {
		if containsIgnoreCase(p.Name(), term) {
			matched = append(matched, p)
		}
	}
	return ProductCollection{items: matched}
}

// Items returns a defensive copy of the inner slice.
func (c ProductCollection) Items() []Product {
	result := make([]Product, len(c.items))
	copy(result, c.items)
	return result
}

// Count returns the number of products.
func (c ProductCollection) Count() int {
	return len(c.items)
}

// containsIgnoreCase checks if s contains substr, ignoring case.
func containsIgnoreCase(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if len(substr) == 0 {
		return true
	}
	// Simple case-insensitive containment check
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalFold compares two strings case-insensitively (ASCII only).
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if toLower(a[i]) != toLower(b[i]) {
			return false
		}
	}
	return true
}

// toLower converts an ASCII byte to lowercase.
func toLower(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return c + 32
	}
	return c
}
