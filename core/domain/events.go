package domain

import "time"

// ---------------------------------------------------------------------------
// Domain Events
// Emitted by aggregates and consumed by the application layer.
// Each event is an immutable record of something that happened.
// ---------------------------------------------------------------------------

// OrderPlaced is emitted when a customer completes checkout.
type OrderPlaced struct {
	OrderID     OrderID
	CustomerID  CustomerID
	Total       Money
	PaymentLink string
	OccurredAt  time.Time
}

// PaymentConfirmed is emitted when a payment webhook confirms success.
type PaymentConfirmed struct {
	OrderID   OrderID
	PaymentID string
	Amount    Money
	OccurredAt time.Time
}

// PaymentFailed is emitted when a payment is declined or expires.
type PaymentFailed struct {
	OrderID    OrderID
	Reason     string
	OccurredAt time.Time
}

// OrderCancelled is emitted when an order is cancelled by customer or admin.
type OrderCancelled struct {
	OrderID    OrderID
	Reason     string
	OccurredAt time.Time
}

// CartExpired is emitted when a cart remains inactive past its expiry.
type CartExpired struct {
	CartID     CartID
	CustomerID CustomerID
	OccurredAt time.Time
}
