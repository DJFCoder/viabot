package domain

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// Port: PaymentGateway
// Abstraction for payment processing (Mercado Pago, Stripe, etc.).
// ---------------------------------------------------------------------------

// PaymentStatus represents the current state of a payment.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentApproved  PaymentStatus = "approved"
	PaymentRejected  PaymentStatus = "rejected"
	PaymentRefunded  PaymentStatus = "refunded"
	PaymentCancelled PaymentStatus = "cancelled"
	PaymentExpired   PaymentStatus = "expired"
)

// PaymentMethod represents how the customer paid.
type PaymentMethod string

const (
	PaymentPix    PaymentMethod = "pix"
	PaymentBoleto PaymentMethod = "boleto"
	PaymentCard   PaymentMethod = "credit_card"
)

// Payment holds the full details of a payment transaction.
// Rule 8 is relaxed for Payment (data-heavy domain type).
type Payment struct {
	id            string
	orderID       OrderID
	status        PaymentStatus
	amount        Money
	paidAmount    Money
	method        PaymentMethod
	checkoutLink  string
	qrCode        string
	qrCodeText    string
	boletoBarcode string
	boletoURL     string
	expiresAt     time.Time
	createdAt     time.Time
	confirmedAt   time.Time
}

// NewPayment creates a Payment value object.
func NewPayment(
	id string,
	orderID OrderID,
	amount Money,
	method PaymentMethod,
	checkoutLink string,
	expiresAt time.Time,
) *Payment {
	return &Payment{
		id:           id,
		orderID:      orderID,
		status:       PaymentPending,
		amount:       amount,
		method:       method,
		checkoutLink: checkoutLink,
		expiresAt:    expiresAt,
		createdAt:    time.Now(),
	}
}

// Confirm marks the payment as approved.
func (p *Payment) Confirm(paidAmount Money) {
	p.status = PaymentApproved
	p.paidAmount = paidAmount
	p.confirmedAt = time.Now()
}

// Reject marks the payment as rejected.
func (p *Payment) Reject() {
	p.status = PaymentRejected
}

// ---------------------------------------------------------------------------
// Payment read accessors
// ---------------------------------------------------------------------------

// ID returns the payment identifier.
func (p *Payment) ID() string { return p.id }

// OrderID returns the associated order identifier.
func (p *Payment) OrderID() OrderID { return p.orderID }

// Status returns the current payment status.
func (p *Payment) Status() PaymentStatus { return p.status }

// Amount returns the total payment amount.
func (p *Payment) Amount() Money { return p.amount }

// PaidAmount returns the amount that was actually paid (may differ from Amount for partial captures).
func (p *Payment) PaidAmount() Money { return p.paidAmount }

// Method returns the payment method used.
func (p *Payment) Method() PaymentMethod { return p.method }

// CheckoutLink returns the URL the customer uses to complete payment.
func (p *Payment) CheckoutLink() string { return p.checkoutLink }

// QRCode returns the Pix QR code (base64-encoded image).
func (p *Payment) QRCode() string { return p.qrCode }

// QRCodeText returns the Pix copy-paste code.
func (p *Payment) QRCodeText() string { return p.qrCodeText }

// BoletoBarcode returns the boleto barcode number.
func (p *Payment) BoletoBarcode() string { return p.boletoBarcode }

// BoletoURL returns the boleto printable URL.
func (p *Payment) BoletoURL() string { return p.boletoURL }

// ExpiresAt returns when the payment expires.
func (p *Payment) ExpiresAt() time.Time { return p.expiresAt }

// CreatedAt returns when the payment was created.
func (p *Payment) CreatedAt() time.Time { return p.createdAt }

// ConfirmedAt returns when the payment was confirmed (zero time if not confirmed).
func (p *Payment) ConfirmedAt() time.Time { return p.confirmedAt }

// ---------------------------------------------------------------------------
// WebhookEvent — normalized webhook from any payment gateway
// ---------------------------------------------------------------------------

// WebhookEvent represents a normalized webhook from any payment gateway.
// Used by the application layer to process payment updates uniformly.
type WebhookEvent struct {
	action    string
	paymentID string
	orderID   OrderID
	status    PaymentStatus
}

// NewWebhookEvent creates a normalized webhook event.
func NewWebhookEvent(action, paymentID string, orderID OrderID, status PaymentStatus) *WebhookEvent {
	return &WebhookEvent{
		action:    action,
		paymentID: paymentID,
		orderID:   orderID,
		status:    status,
	}
}

// Action returns the webhook action type (e.g., "payment.created", "payment.approved").
func (e *WebhookEvent) Action() string { return e.action }

// PaymentID returns the payment gateway's transaction ID.
func (e *WebhookEvent) PaymentID() string { return e.paymentID }

// OrderID returns the associated order identifier.
func (e *WebhookEvent) OrderID() OrderID { return e.orderID }

// Status returns the payment status from the webhook.
func (e *WebhookEvent) Status() PaymentStatus { return e.status }

// ---------------------------------------------------------------------------
// Port: PaymentGateway
// ---------------------------------------------------------------------------

// PaymentGateway defines the contract for payment processing.
// Implementations: Mercado Pago (MVP), Stripe (future).
type PaymentGateway interface {
	// CreatePayment initiates a payment for the given order.
	CreatePayment(ctx context.Context, order *Order) (*Payment, error)

	// GetPaymentStatus retrieves the current status of a payment.
	GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)

	// HandleWebhook processes an incoming webhook payload from the gateway.
	HandleWebhook(ctx context.Context, payload []byte) (*WebhookEvent, error)
}
