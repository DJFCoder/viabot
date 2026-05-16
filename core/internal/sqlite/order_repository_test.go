package sqlite

import (
	"context"
	"testing"
	"time"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// OrderRepository — SQLite adapter integration tests
// ---------------------------------------------------------------------------

func setupCustomerAndCart(t *testing.T, ctx context.Context) (*domain.Customer, *domain.Cart) {
	t.Helper()
	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Order Test",
	)
	cart := domain.NewCart(domain.NewCartID("cart_for_order"), customer.ID())
	price, _ := domain.NewMoney(1000, "BRL")
	_ = cart.AddItem(domain.NewProductID("prod_1"), "Item 1", price, 2)
	_ = cart.AddItem(domain.NewProductID("prod_2"), "Item 2", price, 1)
	return customer, cart
}

func createTestOrder(t *testing.T, ctx context.Context, customer *domain.Cart) *domain.Order {
	t.Helper()
	orderID := domain.NewOrderID("ord_test_1")
	order, err := customer.Checkout(orderID)
	if err != nil {
		t.Fatalf("Checkout failed: %v", err)
	}
	return order
}

func TestOrderRepositorySaveAndFindByID(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	order := createTestOrder(t, ctx, cart)
	if err := orderRepository.Save(ctx, order); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	found, err := orderRepository.FindByID(ctx, order.ID())
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.ID() != order.ID() {
		t.Errorf("expected order ID %s, got %s", order.ID().Value(), found.ID().Value())
	}
	if found.Status() != domain.OrderStatusPending {
		t.Errorf("expected pending, got %s", found.Status())
	}
	if found.Total().Amount() != 3000 {
		t.Errorf("expected total 3000, got %d", found.Total().Amount())
	}
	if found.Items().Count() != 2 {
		t.Errorf("expected 2 items, got %d", found.Items().Count())
	}
}

func TestOrderRepositoryFindByIDNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewOrderRepository(database)
	ctx := context.Background()

	_, err := repository.FindByID(ctx, domain.NewOrderID("nonexistent"))
	if err != domain.ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestOrderRepositoryFindByCustomer(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	order := createTestOrder(t, ctx, cart)
	_ = orderRepository.Save(ctx, order)

	orders, err := orderRepository.FindByCustomer(ctx, customer.ID())
	if err != nil {
		t.Fatalf("FindByCustomer failed: %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orders))
	}
}

func TestOrderRepositoryFindByCustomerEmpty(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewOrderRepository(database)
	ctx := context.Background()

	orders, err := repository.FindByCustomer(ctx, domain.NewCustomerID("nonexistent"))
	if err != nil {
		t.Fatalf("FindByCustomer failed: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}
}

func TestOrderRepositoryFindPendingByCustomer(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	order := createTestOrder(t, ctx, cart)
	_ = orderRepository.Save(ctx, order)

	found, err := orderRepository.FindPendingByCustomer(ctx, customer.ID())
	if err != nil {
		t.Fatalf("FindPendingByCustomer failed: %v", err)
	}
	if found.ID() != order.ID() {
		t.Errorf("expected order ID %s, got %s", order.ID().Value(), found.ID().Value())
	}
}

func TestOrderRepositoryFindPendingByCustomerNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewOrderRepository(database)
	ctx := context.Background()

	_, err := repository.FindPendingByCustomer(ctx, domain.NewCustomerID("nonexistent"))
	if err != domain.ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestOrderRepositoryUpdateStatus(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	order := createTestOrder(t, ctx, cart)
	_ = orderRepository.Save(ctx, order)

	if err := orderRepository.UpdateStatus(ctx, order.ID(), domain.OrderStatusCancelled); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	found, _ := orderRepository.FindByID(ctx, order.ID())
	if found.Status() != domain.OrderStatusCancelled {
		t.Errorf("expected cancelled, got %s", found.Status())
	}
}

func TestOrderRepositoryUpdateStatusNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewOrderRepository(database)
	ctx := context.Background()

	err := repository.UpdateStatus(ctx, domain.NewOrderID("nonexistent"), domain.OrderStatusCancelled)
	if err != domain.ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestOrderRepositorySaveWithPayment(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	order := createTestOrder(t, ctx, cart)
	_ = order.MarkAsAwaitingPayment()

	payment := domain.NewPayment(
		"pay_test_1",
		order.ID(),
		order.Total(),
		domain.PaymentPix,
		"https://checkout.example.com/pay_test_1",
		time.Now().Add(24*time.Hour),
	)
	payment.Confirm(order.Total())
	_ = order.ConfirmPayment(payment)

	if err := orderRepository.Save(ctx, order); err != nil {
		t.Fatalf("Save with payment failed: %v", err)
	}

	found, _ := orderRepository.FindByID(ctx, order.ID())
	if found.Status() != domain.OrderStatusConfirmed {
		t.Errorf("expected confirmed, got %s", found.Status())
	}
	if found.ConfirmedAt().IsZero() {
		t.Error("expected confirmedAt to be set")
	}
	if found.Payment() == nil {
		t.Fatal("expected payment to be present")
	}
	if found.Payment().ID() != "pay_test_1" {
		t.Errorf("expected payment ID pay_test_1, got %s", found.Payment().ID())
	}
	if found.Payment().Status() != domain.PaymentApproved {
		t.Errorf("expected payment approved, got %s", found.Payment().Status())
	}
}

func TestOrderRepositorySaveFullLifecycle(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	// Create, save, then update through full lifecycle
	order := createTestOrder(t, ctx, cart)
	_ = orderRepository.Save(ctx, order)

	// pending → awaiting_payment
	_ = order.MarkAsAwaitingPayment()
	_ = orderRepository.Save(ctx, order)

	// awaiting_payment → confirmed (with payment)
	payment := domain.NewPayment(
		"pay_lifecycle",
		order.ID(),
		order.Total(),
		domain.PaymentPix,
		"https://checkout.example.com/pay_lifecycle",
		time.Now().Add(24*time.Hour),
	)
	payment.Confirm(order.Total())
	_ = order.ConfirmPayment(payment)
	_ = orderRepository.Save(ctx, order)

	// Remove order from memory and re-fetch to verify persistence
	found, err := orderRepository.FindByID(ctx, order.ID())
	if err != nil {
		t.Fatalf("FindByID failed after lifecycle: %v", err)
	}
	if found.Status() != domain.OrderStatusConfirmed {
		t.Errorf("expected confirmed status, got %s", found.Status())
	}
	if found.Payment() == nil {
		t.Fatal("payment should be persisted")
	}
	if found.Payment().Status() != domain.PaymentApproved {
		t.Errorf("expected approved payment, got %s", found.Payment().Status())
	}
}

func TestOrderRepositoryUpdateExistingOrder(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	order := createTestOrder(t, ctx, cart)
	_ = orderRepository.Save(ctx, order)

	// Change items: remove one item
	updatedItems := order.Items().Remove(domain.NewProductID("prod_1"))
	// We can't modify order items directly, so we test via status update
	// and verify the original items are preserved
	_ = order.MarkAsAwaitingPayment()
	_ = orderRepository.Save(ctx, order)

	found, _ := orderRepository.FindByID(ctx, order.ID())
	// Items should remain unchanged from original save
	_ = updatedItems
	if found.Items().Count() != 2 {
		t.Errorf("expected 2 items (unchanged), got %d", found.Items().Count())
	}
}

func TestOrderRepositoryBoletoPaymentRoundTrip(t *testing.T) {
	database := openTestDatabase(t)
	customerRepository := NewCustomerRepository(database)
	cartRepository := NewCartRepository(database)
	orderRepository := NewOrderRepository(database)
	ctx := context.Background()

	customer, cart := setupCustomerAndCart(t, ctx)
	_ = customerRepository.Save(ctx, customer)
	_ = cartRepository.Save(ctx, cart)

	order := createTestOrder(t, ctx, cart)
	_ = order.MarkAsAwaitingPayment()

	// Simulate Mercado Pago boleto payment (initially pending)
	payment := domain.NewPayment(
		"pay_boleto_1",
		order.ID(),
		order.Total(),
		domain.PaymentBoleto,
		"https://checkout.example.com/boleto_1",
		time.Now().Add(72*time.Hour),
	)
	_ = order.ConfirmPayment(payment)

	if err := orderRepository.Save(ctx, order); err != nil {
		t.Fatalf("Save with boleto payment failed: %v", err)
	}

	found, _ := orderRepository.FindByID(ctx, order.ID())
	if found.Payment() == nil {
		t.Fatal("expected payment to be persisted")
	}
	if found.Payment().Method() != domain.PaymentBoleto {
		t.Errorf("expected boleto method, got %s", found.Payment().Method())
	}
	// Boleto starts as pending — confirmed only after manual payment
	if found.Payment().Status() != domain.PaymentPending {
		t.Errorf("expected pending, got %s", found.Payment().Status())
	}
}
