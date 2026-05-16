package application

import (
	"context"
	"fmt"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// Application Service: Cart
// Manages the shopping cart lifecycle.
// Rule 8 compliance: CartService composes exactly 2 instance variables.
// ---------------------------------------------------------------------------

// CartService handles cart use cases for customers.
type CartService struct {
	cartRepo    domain.CartRepository
	productRepo domain.ProductRepository
}

// NewCartService creates a CartService with injected dependencies (DIP).
func NewCartService(
	cartRepo domain.CartRepository,
	productRepo domain.ProductRepository,
) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// AddToCart adds a product to the customer's cart.
// If the customer has no cart yet, a new one is created.
// Guard clause: product must exist and be available.
func (s *CartService) AddToCart(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID, quantity int) error {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("find product: %w", err)
	}
	if !product.IsAvailable() {
		return domain.ErrProductUnavailable
	}

	cart, err := s.loadOrCreateCart(ctx, customerID)
	if err != nil {
		return fmt.Errorf("load cart: %w", err)
	}
	if err := cart.AddItem(product.ID(), product.Name(), product.Price(), quantity); err != nil {
		return fmt.Errorf("add item: %w", err)
	}
	return s.cartRepo.Save(ctx, cart)
}

// RemoveFromCart removes a product from the customer's cart.
func (s *CartService) RemoveFromCart(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID) error {
	cart, err := s.cartRepo.FindByCustomer(ctx, customerID)
	if err != nil {
		return fmt.Errorf("find cart: %w", err)
	}
	cart.RemoveItem(productID)
	return s.cartRepo.Save(ctx, cart)
}

// UpdateCartItem changes the quantity of a product in the cart.
func (s *CartService) UpdateCartItem(ctx context.Context, customerID domain.CustomerID, productID domain.ProductID, quantity int) error {
	cart, err := s.cartRepo.FindByCustomer(ctx, customerID)
	if err != nil {
		return fmt.Errorf("find cart: %w", err)
	}
	if err := cart.UpdateQuantity(productID, quantity); err != nil {
		return fmt.Errorf("update quantity: %w", err)
	}
	return s.cartRepo.Save(ctx, cart)
}

// ClearCart empties the customer's cart.
func (s *CartService) ClearCart(ctx context.Context, customerID domain.CustomerID) error {
	cart, err := s.cartRepo.FindByCustomer(ctx, customerID)
	if err != nil {
		return fmt.Errorf("find cart: %w", err)
	}
	cart.Clear()
	return s.cartRepo.Save(ctx, cart)
}

// GetCart retrieves the customer's current cart.
func (s *CartService) GetCart(ctx context.Context, customerID domain.CustomerID) (*domain.Cart, error) {
	return s.cartRepo.FindByCustomer(ctx, customerID)
}

// loadOrCreateCart finds an existing cart or creates a new one.
func (s *CartService) loadOrCreateCart(ctx context.Context, customerID domain.CustomerID) (*domain.Cart, error) {
	cart, err := s.cartRepo.FindByCustomer(ctx, customerID)
	if err == nil {
		return cart, nil
	}
	cartID := domain.NewCartID(fmt.Sprintf("cart_%s", customerID.Value()))
	cart = domain.NewCart(cartID, customerID)
	return cart, nil
}
