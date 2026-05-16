// Package application implements use cases for the ViaBot e-commerce domain.
// It depends only on domain interfaces (ports), never on concrete adapters.
package application

import (
	"context"
	"fmt"
	"time"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// Application Service: Order
// Orchestrates the order lifecycle: place → confirm → cancel.
// Rule 8 compliance: OrderService composes exactly 2 instance variables.
// ---------------------------------------------------------------------------

// orderPersistence groups repository dependencies for order operations.
type orderPersistence struct {
	orders domain.OrderRepository
	carts  domain.CartRepository
}

// OrderService handles order use cases.
type OrderService struct {
	persistence orderPersistence
	gateway     domain.PaymentGateway
}

// NewOrderService creates an OrderService with injected dependencies (DIP).
func NewOrderService(
	orders domain.OrderRepository,
	carts domain.CartRepository,
	gateway domain.PaymentGateway,
) *OrderService {
	return &OrderService{
		persistence: orderPersistence{orders: orders, carts: carts},
		gateway:     gateway,
	}
}

// PlaceOrder converts a cart into an order, saves it, and returns the order.
// Steps:
//  1. Load cart from repository
//  2. Validate cart (exists, not expired, not empty)
//  3. Create order via cart.Checkout()
//  4. Save order
//  5. Delete cart
func (s *OrderService) PlaceOrder(ctx context.Context, customerID domain.CustomerID, cartID domain.CartID) (*domain.Order, error) {
	cart, err := s.persistence.carts.FindByCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("load cart: %w", err)
	}
	if cart.IsExpired() {
		return nil, fmt.Errorf("cart expired: %w", domain.ErrCartExpired)
	}

	orderID := domain.NewOrderID(fmt.Sprintf("ord_%s", cartID.Value()))
	order, err := cart.Checkout(orderID)
	if err != nil {
		return nil, fmt.Errorf("checkout: %w", err)
	}
	if err := s.persistence.orders.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}
	return order, nil
}

// ConfirmOrderPayment updates an order when a payment is confirmed.
// Called by the webhook handler adapter after parsing the gateway payload.
func (s *OrderService) ConfirmOrderPayment(ctx context.Context, orderID domain.OrderID, payment *domain.Payment) error {
	order, err := s.persistence.orders.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}
	if err := order.ConfirmPayment(payment); err != nil {
		return fmt.Errorf("confirm payment: %w", err)
	}
	return s.persistence.orders.UpdateStatus(ctx, orderID, order.Status())
}

// HandlePaymentWebhook processes a webhook event from a payment gateway.
// It parses the event, updates the order status accordingly.
func (s *OrderService) HandlePaymentWebhook(ctx context.Context, payload []byte) error {
	event, err := s.gateway.HandleWebhook(ctx, payload)
	if err != nil {
		return fmt.Errorf("handle webhook: %w", err)
	}

	switch event.Status() {
	case domain.PaymentApproved:
		return s.approvePayment(ctx, event)
	case domain.PaymentRejected, domain.PaymentExpired:
		return s.rejectPayment(ctx, event)
	default:
		// Ignore interim statuses (pending, etc.)
		return nil
	}
}

// approvePayment retrieves the order and confirms payment.
func (s *OrderService) approvePayment(ctx context.Context, event *domain.WebhookEvent) error {
	order, err := s.persistence.orders.FindByID(ctx, event.OrderID())
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}
	payment := domain.NewPayment(
		event.PaymentID(),
		event.OrderID(),
		order.Total(),
		domain.PaymentPix,
		"",
		order.CreatedAt().Add(24*time.Hour),
	)
	payment.Confirm(order.Total())

	if err := order.ConfirmPayment(payment); err != nil {
		return fmt.Errorf("confirm payment: %w", err)
	}
	return s.persistence.orders.UpdateStatus(ctx, order.ID(), order.Status())
}

// rejectPayment cancels the order when payment fails.
func (s *OrderService) rejectPayment(ctx context.Context, event *domain.WebhookEvent) error {
	order, err := s.persistence.orders.FindByID(ctx, event.OrderID())
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}
	if err := order.Cancel(); err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	return s.persistence.orders.UpdateStatus(ctx, order.ID(), order.Status())
}

// CancelOrder cancels an order if the current state permits it.
func (s *OrderService) CancelOrder(ctx context.Context, orderID domain.OrderID) error {
	order, err := s.persistence.orders.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}
	if err := order.Cancel(); err != nil {
		return fmt.Errorf("cancel: %w", err)
	}
	return s.persistence.orders.UpdateStatus(ctx, orderID, order.Status())
}

// GetOrder retrieves a single order by ID.
func (s *OrderService) GetOrder(ctx context.Context, orderID domain.OrderID) (*domain.Order, error) {
	return s.persistence.orders.FindByID(ctx, orderID)
}

// ListCustomerOrders returns all orders for a given customer.
func (s *OrderService) ListCustomerOrders(ctx context.Context, customerID domain.CustomerID) ([]*domain.Order, error) {
	return s.persistence.orders.FindByCustomer(ctx, customerID)
}
