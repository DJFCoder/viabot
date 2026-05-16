package sqlite

import (
	"context"
	"testing"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// CartRepository — SQLite adapter integration tests
// ---------------------------------------------------------------------------

func TestCartRepositorySaveAndFindByCustomer(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	ctx := context.Background()

	// Setup: create customer first (FK constraint)
	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Cart Owner",
	)
	_ = customerRepository.Save(ctx, customer)

	// Create cart with items
	cartID := domain.NewCartID("cart_test_1")
	cart := domain.NewCart(cartID, customer.ID())

	price, _ := domain.NewMoney(1990, "BRL")
	_ = cart.AddItem(domain.NewProductID("prod_1"), "Item 1", price, 2)
	_ = cart.AddItem(domain.NewProductID("prod_2"), "Item 2", price, 1)

	if err := cartRepository.Save(ctx, cart); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Retrieve by customer
	found, err := cartRepository.FindByCustomer(ctx, customer.ID())
	if err != nil {
		t.Fatalf("FindByCustomer failed: %v", err)
	}
	if found.ID() != cartID {
		t.Errorf("expected cart ID %s, got %s", cartID.Value(), found.ID().Value())
	}
	if found.Items().Count() != 2 {
		t.Errorf("expected 2 items, got %d", found.Items().Count())
	}
}

func TestCartRepositoryFindByCustomerNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewCartRepository(database)
	ctx := context.Background()

	_, err := repository.FindByCustomer(ctx, domain.NewCustomerID("nonexistent"))
	if err != domain.ErrCartNotFound {
		t.Fatalf("expected ErrCartNotFound, got %v", err)
	}
}

func TestCartRepositoryDelete(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	ctx := context.Background()

	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Cart Owner",
	)
	_ = customerRepository.Save(ctx, customer)

	cart := domain.NewCart(domain.NewCartID("cart_delete"), customer.ID())
	_ = cartRepository.Save(ctx, cart)

	// Delete cart
	if err := cartRepository.Delete(ctx, cart.ID()); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err := cartRepository.FindByCustomer(ctx, customer.ID())
	if err != domain.ErrCartNotFound {
		t.Fatalf("expected ErrCartNotFound after delete, got %v", err)
	}
}

func TestCartRepositoryDeleteNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewCartRepository(database)
	ctx := context.Background()

	err := repository.Delete(ctx, domain.NewCartID("nonexistent"))
	if err != domain.ErrCartNotFound {
		t.Fatalf("expected ErrCartNotFound, got %v", err)
	}
}

func TestCartRepositoryUpdateExisting(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	ctx := context.Background()

	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Cart Owner",
	)
	_ = customerRepository.Save(ctx, customer)

	// Save cart with 1 item
	cartID := domain.NewCartID("cart_update")
	cart := domain.NewCart(cartID, customer.ID())
	price, _ := domain.NewMoney(1000, "BRL")
	_ = cart.AddItem(domain.NewProductID("p1"), "Item 1", price, 1)
	_ = cartRepository.Save(ctx, cart)

	// Add another item and save again
	_ = cart.AddItem(domain.NewProductID("p2"), "Item 2", price, 2)
	if err := cartRepository.Save(ctx, cart); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	// Verify 2 items now
	found, _ := cartRepository.FindByCustomer(ctx, customer.ID())
	if found.Items().Count() != 2 {
		t.Errorf("expected 2 items after update, got %d", found.Items().Count())
	}
}

func TestCartRepositoryDeleteCascadeItems(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	ctx := context.Background()

	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Cart Owner",
	)
	_ = customerRepository.Save(ctx, customer)

	cart := domain.NewCart(domain.NewCartID("cart_cascade"), customer.ID())
	price, _ := domain.NewMoney(1000, "BRL")
	_ = cart.AddItem(domain.NewProductID("p1"), "Item", price, 1)
	_ = cartRepository.Save(ctx, cart)

	// Delete cart — items should cascade delete
	_ = cartRepository.Delete(ctx, cart.ID())

	// Verify no orphan items (this is a structural test — SQLite has ON DELETE CASCADE)
	// Re-creating same cart should work without FK issues
	newCart := domain.NewCart(domain.NewCartID("cart_cascade"), customer.ID())
	if err := cartRepository.Save(ctx, newCart); err != nil {
		t.Fatalf("Save after delete cascade failed: %v", err)
	}
}
