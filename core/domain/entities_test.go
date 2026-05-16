package domain

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Order — Aggregate Tests
// ---------------------------------------------------------------------------

func makeTestOrder(t *testing.T) *Order {
	t.Helper()
	customerID := NewCustomerID("5511999999999@s.whatsapp.net")
	price, _ := NewMoney(1000, "BRL")
	pid := NewProductID("prod_1")
	item, _ := NewOrderItem(pid, "Test Item", price, 2)
	items := NewOrderItemCollection([]OrderItem{item})
	orderID := NewOrderID("ord_test_1")
	order, err := NewOrder(orderID, customerID, items)
	if err != nil {
		t.Fatalf("NewOrder failed: %v", err)
	}
	return order
}

func TestNewOrderValid(t *testing.T) {
	order := makeTestOrder(t)
	if order.Status() != OrderStatusPending {
		t.Errorf("expected status pending, got %s", order.Status())
	}
	if order.CustomerID().Value() != "5511999999999@s.whatsapp.net" {
		t.Errorf("unexpected customer ID")
	}
}

func TestNewOrderEmptyItems(t *testing.T) {
	customerID := NewCustomerID("test")
	items := NewOrderItemCollection(nil)
	orderID := NewOrderID("ord_empty")
	_, err := NewOrder(orderID, customerID, items)
	if err == nil {
		t.Fatal("expected error for empty order items")
	}
}

func makeOrderAwaitingPayment(t *testing.T, order *Order) {
	t.Helper()
	if err := order.MarkAsAwaitingPayment(); err != nil {
		t.Fatalf("MarkAsAwaitingPayment failed: %v", err)
	}
}

func TestOrderConfirmPayment(t *testing.T) {
	order := makeTestOrder(t)
	makeOrderAwaitingPayment(t, order)

	payment := NewPayment("pay_1", order.ID(), order.Total(), PaymentPix, "https://checkout", time.Now().Add(24*time.Hour))
	if err := order.ConfirmPayment(payment); err != nil {
		t.Fatalf("ConfirmPayment failed: %v", err)
	}
	if order.Status() != OrderStatusConfirmed {
		t.Errorf("expected status confirmed, got %s", order.Status())
	}
	if order.ConfirmedAt().IsZero() {
		t.Error("expected confirmedAt to be set")
	}
}

func TestOrderCannotConfirmTwice(t *testing.T) {
	order := makeTestOrder(t)
	makeOrderAwaitingPayment(t, order)

	payment := NewPayment("pay_1", order.ID(), order.Total(), PaymentPix, "", time.Now().Add(24*time.Hour))
	_ = order.ConfirmPayment(payment)

	payment2 := NewPayment("pay_2", order.ID(), order.Total(), PaymentPix, "", time.Now().Add(24*time.Hour))
	if err := order.ConfirmPayment(payment2); err == nil {
		t.Fatal("expected error confirming already confirmed order")
	}
}

func TestOrderCancelFromPending(t *testing.T) {
	order := makeTestOrder(t)
	if err := order.Cancel(); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}
	if order.Status() != OrderStatusCancelled {
		t.Errorf("expected cancelled, got %s", order.Status())
	}
}

func TestOrderCannotCancelAfterConfirmed(t *testing.T) {
	order := makeTestOrder(t)
	makeOrderAwaitingPayment(t, order)

	payment := NewPayment("pay_1", order.ID(), order.Total(), PaymentPix, "", time.Now().Add(24*time.Hour))
	_ = order.ConfirmPayment(payment)

	if err := order.Cancel(); err == nil {
		t.Fatal("expected error cancelling confirmed order")
	}
}

func TestOrderMarkAsShipped(t *testing.T) {
	order := makeTestOrder(t)
	makeOrderAwaitingPayment(t, order)

	payment := NewPayment("pay_1", order.ID(), order.Total(), PaymentPix, "", time.Now().Add(24*time.Hour))
	_ = order.ConfirmPayment(payment)

	if err := order.MarkAsShipped(); err != nil {
		t.Fatalf("MarkAsShipped failed: %v", err)
	}
	if order.Status() != OrderStatusShipped {
		t.Errorf("expected shipped, got %s", order.Status())
	}
}

func TestOrderFullLifecycle(t *testing.T) {
	order := makeTestOrder(t)
	makeOrderAwaitingPayment(t, order)

	// pending → awaiting_payment → confirmed → shipped → delivered
	payment := NewPayment("pay_1", order.ID(), order.Total(), PaymentPix, "", time.Now().Add(24*time.Hour))
	_ = order.ConfirmPayment(payment)
	_ = order.MarkAsShipped()
	if err := order.MarkAsDelivered(); err != nil {
		t.Fatalf("MarkAsDelivered failed: %v", err)
	}
	if order.Status() != OrderStatusDelivered {
		t.Errorf("expected delivered, got %s", order.Status())
	}
}

// ---------------------------------------------------------------------------
// Cart — Aggregate Tests
// ---------------------------------------------------------------------------

func makeTestCart(t *testing.T) *Cart {
	t.Helper()
	customerID := NewCustomerID("5511999999999@s.whatsapp.net")
	cartID := NewCartID("cart_test_1")
	return NewCart(cartID, customerID)
}

func TestNewCart(t *testing.T) {
	cart := makeTestCart(t)
	if cart.IsExpired() {
		t.Error("new cart should not be expired")
	}
	if cart.Items().Count() != 0 {
		t.Errorf("expected empty cart, got %d items", cart.Items().Count())
	}
}

func TestCartAddItem(t *testing.T) {
	cart := makeTestCart(t)
	price, _ := NewMoney(500, "BRL")
	pid := NewProductID("prod_1")

	if err := cart.AddItem(pid, "Product 1", price, 2); err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}
	if cart.Items().Count() != 1 {
		t.Errorf("expected 1 item, got %d", cart.Items().Count())
	}
}

func TestCartAddDuplicateItemIncrementsQuantity(t *testing.T) {
	cart := makeTestCart(t)
	price, _ := NewMoney(500, "BRL")
	pid := NewProductID("prod_1")

	_ = cart.AddItem(pid, "Product 1", price, 2)
	_ = cart.AddItem(pid, "Product 1", price, 3)

	if cart.Items().Count() != 1 {
		t.Errorf("expected 1 item (duplicates merged), got %d", cart.Items().Count())
	}
	// Expected quantity: 2 + 3 = 5
	for _, item := range cart.Items().Items() {
		if item.Quantity() != 5 {
			t.Errorf("expected quantity 5, got %d", item.Quantity())
		}
	}
}

func TestCartRemoveItem(t *testing.T) {
	cart := makeTestCart(t)
	price, _ := NewMoney(500, "BRL")
	pid1 := NewProductID("prod_1")
	pid2 := NewProductID("prod_2")

	_ = cart.AddItem(pid1, "Product 1", price, 1)
	_ = cart.AddItem(pid2, "Product 2", price, 2)

	cart.RemoveItem(pid1)
	if cart.Items().Count() != 1 {
		t.Errorf("expected 1 item after removal, got %d", cart.Items().Count())
	}
}

func TestCartUpdateQuantity(t *testing.T) {
	cart := makeTestCart(t)
	price, _ := NewMoney(500, "BRL")
	pid := NewProductID("prod_1")

	_ = cart.AddItem(pid, "Product 1", price, 2)
	_ = cart.UpdateQuantity(pid, 5)

	for _, item := range cart.Items().Items() {
		if item.Quantity() != 5 {
			t.Errorf("expected quantity 5, got %d", item.Quantity())
		}
	}
}

func TestCartUpdateQuantityInvalid(t *testing.T) {
	cart := makeTestCart(t)
	price, _ := NewMoney(500, "BRL")
	pid := NewProductID("prod_1")

	_ = cart.AddItem(pid, "Product 1", price, 2)
	if err := cart.UpdateQuantity(pid, 0); err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestCartClear(t *testing.T) {
	cart := makeTestCart(t)
	price, _ := NewMoney(500, "BRL")
	pid := NewProductID("prod_1")

	_ = cart.AddItem(pid, "Product 1", price, 2)
	cart.Clear()

	if cart.Items().Count() != 0 {
		t.Errorf("expected 0 items after clear, got %d", cart.Items().Count())
	}
}

func TestCartCheckout(t *testing.T) {
	cart := makeTestCart(t)
	price, _ := NewMoney(1000, "BRL")
	pid := NewProductID("prod_1")

	_ = cart.AddItem(pid, "Product 1", price, 2)
	orderID := NewOrderID("ord_from_cart")

	order, err := cart.Checkout(orderID)
	if err != nil {
		t.Fatalf("Checkout failed: %v", err)
	}
	if order.Status() != OrderStatusPending {
		t.Errorf("expected pending, got %s", order.Status())
	}
	if order.Total().Amount() != 2000 {
		t.Errorf("expected total 2000, got %d", order.Total().Amount())
	}
}

func TestCartCheckoutEmpty(t *testing.T) {
	cart := makeTestCart(t)
	orderID := NewOrderID("ord_empty_cart")

	_, err := cart.Checkout(orderID)
	if err == nil {
		t.Fatal("expected error checking out empty cart")
	}
}

// ---------------------------------------------------------------------------
// Customer — Entity Tests
// ---------------------------------------------------------------------------

func TestNewCustomer(t *testing.T) {
	id := NewCustomerID("5511999999999@s.whatsapp.net")
	phone := NewPhoneNumber(55, "11999999999")
	customer := NewCustomer(id, phone, "John Doe")

	if customer.Name() != "John Doe" {
		t.Errorf("expected name John Doe, got %s", customer.Name())
	}
	if customer.Phone().CountryCode() != 55 {
		t.Errorf("expected country code 55, got %d", customer.Phone().CountryCode())
	}
}

func TestCustomerUpdateName(t *testing.T) {
	id := NewCustomerID("5511999999999@s.whatsapp.net")
	phone := NewPhoneNumber(55, "11999999999")
	customer := NewCustomer(id, phone, "John")

	customer.UpdateName("Jane")
	if customer.Name() != "Jane" {
		t.Errorf("expected name Jane, got %s", customer.Name())
	}
}

// ---------------------------------------------------------------------------
// Product — Entity Tests
// ---------------------------------------------------------------------------

func TestNewProduct(t *testing.T) {
	price, _ := NewMoney(2990, "BRL")
	pid := NewProductID("prod_1")
	product := NewProduct(pid, "T-Shirt", "Cotton T-Shirt", price, "clothing", true, "https://img.url")

	if product.Name() != "T-Shirt" {
		t.Errorf("expected T-Shirt, got %s", product.Name())
	}
	if !product.IsAvailable() {
		t.Error("expected product to be available")
	}
}

func TestProductUpdateAvailability(t *testing.T) {
	price, _ := NewMoney(2990, "BRL")
	pid := NewProductID("prod_1")
	product := NewProduct(pid, "T-Shirt", "", price, "clothing", true, "")

	product.UpdateAvailability(false)
	if product.IsAvailable() {
		t.Error("expected product to be unavailable after update")
	}
}
