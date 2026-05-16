package domain

// ---------------------------------------------------------------------------
// Value Objects
// ---------------------------------------------------------------------------

// Money represents an amount in a specific currency.
// Wraps primitive (int64 for cents) to prevent primitive obsession (Rule 3).
type Money struct {
	amount   int64  // amount in cents (e.g., R$ 10,50 = 1050)
	currency string // "BRL", "USD", "EUR"
}

// NewMoney creates a Money value object with validation.
func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}
	return Money{amount: amount, currency: currency}, nil
}

// Amount returns the amount in cents.
func (m Money) Amount() int64 { return m.amount }

// Currency returns the currency code.
func (m Money) Currency() string { return m.currency }

// Add returns a new Money with the sum, validating same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract returns a new Money with the difference, validating same currency.
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	if m.amount < other.amount {
		return Money{}, ErrNegativeAmount
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

// Multiply returns a new Money multiplied by an integer factor.
func (m Money) Multiply(factor int) Money {
	return Money{amount: m.amount * int64(factor), currency: m.currency}
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool { return m.amount == 0 }

// IsNegative returns true if the amount is negative.
func (m Money) IsNegative() bool { return m.amount < 0 }

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusConfirmed       OrderStatus = "confirmed"
	OrderStatusShipped         OrderStatus = "shipped"
	OrderStatusDelivered       OrderStatus = "delivered"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// validTransitions defines allowed state machine for OrderStatus.
// Guard clause pattern: only explicit transitions are permitted.
var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPending:         {OrderStatusAwaitingPayment, OrderStatusCancelled},
	OrderStatusAwaitingPayment: {OrderStatusConfirmed, OrderStatusCancelled},
	OrderStatusConfirmed:       {OrderStatusShipped},
	OrderStatusShipped:         {OrderStatusDelivered},
}

// CanTransitionTo checks if the current status allows moving to the target.
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	transitions, ok := validTransitions[s]
	if !ok {
		return false
	}
	return containsOrderStatus(transitions, target)
}

// containsOrderStatus checks if a status exists in a slice — helper extracted
// to comply with Rule 1 (one level of indentation per method).
func containsOrderStatus(statuses []OrderStatus, target OrderStatus) bool {
	for _, s := range statuses {
		if s == target {
			return true
		}
	}
	return false
}

// PhoneNumber wraps a phone number with country code.
type PhoneNumber struct {
	countryCode int
	number      string
}

// NewPhoneNumber creates a PhoneNumber.
func NewPhoneNumber(countryCode int, number string) PhoneNumber {
	return PhoneNumber{countryCode: countryCode, number: number}
}

// CountryCode returns the international dialing code.
func (p PhoneNumber) CountryCode() int { return p.countryCode }

// Number returns the local phone number without country code.
func (p PhoneNumber) Number() string { return p.number }

// ---------------------------------------------------------------------------
// Entity IDs
// ---------------------------------------------------------------------------

// OrderID uniquely identifies an Order.
type OrderID struct{ value string }

func NewOrderID(value string) OrderID { return OrderID{value: value} }
func (id OrderID) Value() string      { return id.value }

// ProductID uniquely identifies a Product.
type ProductID struct{ value string }

func NewProductID(value string) ProductID { return ProductID{value: value} }
func (id ProductID) Value() string        { return id.value }

// CartID uniquely identifies a Cart.
type CartID struct{ value string }

func NewCartID(value string) CartID { return CartID{value: value} }
func (id CartID) Value() string     { return id.value }
