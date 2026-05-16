package domain

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Money — Value Object Tests
// ---------------------------------------------------------------------------

func TestNewMoneyValid(t *testing.T) {
	m, err := NewMoney(1050, "BRL")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Amount() != 1050 {
		t.Errorf("expected amount 1050, got %d", m.Amount())
	}
	if m.Currency() != "BRL" {
		t.Errorf("expected currency BRL, got %s", m.Currency())
	}
}

func TestNewMoneyNegativeAmount(t *testing.T) {
	_, err := NewMoney(-1, "BRL")
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestNewMoneyEmptyCurrency(t *testing.T) {
	_, err := NewMoney(100, "")
	if err == nil {
		t.Fatal("expected error for empty currency")
	}
}

func TestMoneyAddSameCurrency(t *testing.T) {
	a, _ := NewMoney(100, "BRL")
	b, _ := NewMoney(200, "BRL")
	result, err := a.Add(b)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Amount() != 300 {
		t.Errorf("expected 300, got %d", result.Amount())
	}
}

func TestMoneyAddDifferentCurrency(t *testing.T) {
	a, _ := NewMoney(100, "BRL")
	b, _ := NewMoney(100, "USD")
	_, err := a.Add(b)
	if err == nil {
		t.Fatal("expected error for currency mismatch")
	}
}

func TestMoneySubtractValid(t *testing.T) {
	a, _ := NewMoney(300, "BRL")
	b, _ := NewMoney(100, "BRL")
	result, err := a.Subtract(b)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Amount() != 200 {
		t.Errorf("expected 200, got %d", result.Amount())
	}
}

func TestMoneySubtractNegativeResult(t *testing.T) {
	a, _ := NewMoney(100, "BRL")
	b, _ := NewMoney(300, "BRL")
	_, err := a.Subtract(b)
	if err == nil {
		t.Fatal("expected error when result would be negative")
	}
}

func TestMoneyMultiply(t *testing.T) {
	m, _ := NewMoney(150, "BRL")
	result := m.Multiply(3)
	if result.Amount() != 450 {
		t.Errorf("expected 450, got %d", result.Amount())
	}
	if result.Currency() != "BRL" {
		t.Errorf("expected currency BRL, got %s", result.Currency())
	}
}

func TestMoneyMultiplyZero(t *testing.T) {
	m, _ := NewMoney(150, "BRL")
	result := m.Multiply(0)
	if result.Amount() != 0 {
		t.Errorf("expected 0, got %d", result.Amount())
	}
}

func TestMoneyIsZero(t *testing.T) {
	zero, _ := NewMoney(0, "BRL")
	nonzero, _ := NewMoney(100, "BRL")
	if !zero.IsZero() {
		t.Error("expected zero to be zero")
	}
	if nonzero.IsZero() {
		t.Error("expected nonzero not to be zero")
	}
}

func TestMoneyIsNegative(t *testing.T) {
	// Money cannot be created with negative amount, but we can test zero and positive
	zero, _ := NewMoney(0, "BRL")
	positive, _ := NewMoney(100, "BRL")
	if zero.IsNegative() {
		t.Error("expected zero not to be negative")
	}
	if positive.IsNegative() {
		t.Error("expected positive not to be negative")
	}
}

// ---------------------------------------------------------------------------
// OrderStatus — Value Object Tests
// ---------------------------------------------------------------------------

func TestOrderStatusValidTransitions(t *testing.T) {
	tests := []struct {
		current OrderStatus
		target  OrderStatus
		valid   bool
	}{
		{OrderStatusPending, OrderStatusAwaitingPayment, true},
		{OrderStatusPending, OrderStatusCancelled, true},
		{OrderStatusPending, OrderStatusConfirmed, false},
		{OrderStatusAwaitingPayment, OrderStatusConfirmed, true},
		{OrderStatusAwaitingPayment, OrderStatusCancelled, true},
		{OrderStatusAwaitingPayment, OrderStatusDelivered, false},
		{OrderStatusConfirmed, OrderStatusShipped, true},
		{OrderStatusConfirmed, OrderStatusCancelled, false},
		{OrderStatusShipped, OrderStatusDelivered, true},
		{OrderStatusShipped, OrderStatusConfirmed, false},
		{OrderStatusDelivered, OrderStatusShipped, false},
		{OrderStatusCancelled, OrderStatusPending, false},
		// Self-transitions
		{OrderStatusPending, OrderStatusPending, false},
		{OrderStatusConfirmed, OrderStatusConfirmed, false},
	}
	for _, tc := range tests {
		got := tc.current.CanTransitionTo(tc.target)
		if got != tc.valid {
			t.Errorf("%s -> %s: expected %v, got %v", tc.current, tc.target, tc.valid, got)
		}
	}
}

// ---------------------------------------------------------------------------
// OrderItem — Value Object Tests
// ---------------------------------------------------------------------------

func TestNewOrderItemValid(t *testing.T) {
	price, _ := NewMoney(1000, "BRL")
	pid := NewProductID("prod_1")
	item, err := NewOrderItem(pid, "Test Product", price, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.Quantity() != 2 {
		t.Errorf("expected quantity 2, got %d", item.Quantity())
	}
}

func TestNewOrderItemInvalidQuantity(t *testing.T) {
	price, _ := NewMoney(1000, "BRL")
	pid := NewProductID("prod_1")
	_, err := NewOrderItem(pid, "Test", price, 0)
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestOrderItemSubTotal(t *testing.T) {
	price, _ := NewMoney(500, "BRL")
	pid := NewProductID("prod_1")
	item, _ := NewOrderItem(pid, "Test", price, 3)
	sub := item.SubTotal()
	if sub.Amount() != 1500 {
		t.Errorf("expected subtotal 1500, got %d", sub.Amount())
	}
}

// ---------------------------------------------------------------------------
// CartItem — Value Object Tests
// ---------------------------------------------------------------------------

func TestNewCartItemValid(t *testing.T) {
	price, _ := NewMoney(2000, "BRL")
	pid := NewProductID("prod_2")
	item, err := NewCartItem(pid, "Cart Product", price, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.Quantity() != 1 {
		t.Errorf("expected quantity 1, got %d", item.Quantity())
	}
}

func TestCartItemSubTotal(t *testing.T) {
	price, _ := NewMoney(1500, "BRL")
	pid := NewProductID("prod_2")
	item, _ := NewCartItem(pid, "Cart Product", price, 4)
	sub := item.SubTotal()
	if sub.Amount() != 6000 {
		t.Errorf("expected subtotal 6000, got %d", sub.Amount())
	}
}
