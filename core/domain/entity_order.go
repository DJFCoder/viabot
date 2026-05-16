package domain

import (
	"time"
)

// ---------------------------------------------------------------------------
// Order — Aggregate Root (E-Commerce Bounded Context)
// Rule 8 compliance: Order composes exactly 2 instance variables.
// ---------------------------------------------------------------------------

// orderDetails groups all non-identity state of an Order.
// Each field is unexported; behaviour is exposed through Order methods.
type orderDetails struct {
	customerID  CustomerID
	status      OrderStatus
	items       OrderItemCollection
	total       Money
	payment     *Payment
	createdAt   time.Time
	confirmedAt *time.Time
}

// Order is the aggregate root for the purchase lifecycle.
type Order struct {
	id      OrderID
	details orderDetails
}

// NewOrder creates a new Order with the given identity, customer, and items.
// It computes the total from the item collection and sets status to pending.
func NewOrder(id OrderID, customerID CustomerID, items OrderItemCollection) (*Order, error) {
	if items.Count() == 0 {
		return nil, ErrEmptyCart
	}
	total, err := items.Total()
	if err != nil {
		return nil, err
	}
	return &Order{
		id: id,
		details: orderDetails{
			customerID: customerID,
			status:     OrderStatusPending,
			items:      items,
			total:      total,
			createdAt:  time.Now(),
		},
	}, nil
}

// MarkAsAwaitingPayment transitions from pending to awaiting_payment.
// Called when a payment link is generated for the order.
func (o *Order) MarkAsAwaitingPayment() error {
	if !o.details.status.CanTransitionTo(OrderStatusAwaitingPayment) {
		return ErrInvalidTransition
	}
	o.details.status = OrderStatusAwaitingPayment
	return nil
}

// ConfirmPayment transitions the order from awaiting_payment to confirmed
// and attaches the payment details.
func (o *Order) ConfirmPayment(payment *Payment) error {
	if !o.details.status.CanTransitionTo(OrderStatusConfirmed) {
		return ErrInvalidTransition
	}
	o.details.status = OrderStatusConfirmed
	o.details.payment = payment
	now := time.Now()
	o.details.confirmedAt = &now
	return nil
}

// Cancel transitions the order to cancelled if permitted.
func (o *Order) Cancel() error {
	if !o.details.status.CanTransitionTo(OrderStatusCancelled) {
		return ErrInvalidTransition
	}
	o.details.status = OrderStatusCancelled
	return nil
}

// MarkAsShipped transitions from confirmed to shipped.
func (o *Order) MarkAsShipped() error {
	if !o.details.status.CanTransitionTo(OrderStatusShipped) {
		return ErrInvalidTransition
	}
	o.details.status = OrderStatusShipped
	return nil
}

// MarkAsDelivered transitions from shipped to delivered.
func (o *Order) MarkAsDelivered() error {
	if !o.details.status.CanTransitionTo(OrderStatusDelivered) {
		return ErrInvalidTransition
	}
	o.details.status = OrderStatusDelivered
	return nil
}

// ---------------------------------------------------------------------------
// Read accessors — expose state for queries, not mutation (Rule 9).
// ---------------------------------------------------------------------------

// ID returns the order identifier.
func (o *Order) ID() OrderID { return o.id }

// CustomerID returns the customer who placed the order.
func (o *Order) CustomerID() CustomerID { return o.details.customerID }

// Status returns the current lifecycle status.
func (o *Order) Status() OrderStatus { return o.details.status }

// Items returns a defensive copy of the order items.
func (o *Order) Items() OrderItemCollection { return o.details.items }

// Total returns the order monetary total.
func (o *Order) Total() Money { return o.details.total }

// Payment returns the associated payment, if any.
func (o *Order) Payment() *Payment {
	if o.details.payment == nil {
		return nil
	}
	return o.details.payment
}

// CreatedAt returns when the order was placed.
func (o *Order) CreatedAt() time.Time { return o.details.createdAt }

// ConfirmedAt returns when the order was paid, or zero time if not confirmed.
func (o *Order) ConfirmedAt() time.Time {
	if o.details.confirmedAt == nil {
		return time.Time{}
	}
	return *o.details.confirmedAt
}
