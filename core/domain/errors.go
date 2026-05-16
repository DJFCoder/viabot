package domain

import "errors"

// ---------------------------------------------------------------------------
// Domain errors — defined as sentinel errors for zero-dependency error handling.
// ---------------------------------------------------------------------------

var (
	ErrNegativeAmount    = errors.New("amount must not be negative")
	ErrInvalidCurrency   = errors.New("currency must not be empty")
	ErrCurrencyMismatch  = errors.New("currency mismatch between values")
	ErrInvalidOrder      = errors.New("order is invalid")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrPaymentFailed     = errors.New("payment processing failed")
	ErrProductNotFound   = errors.New("product not found")
	ErrCartNotFound      = errors.New("cart not found")
	ErrCartExpired       = errors.New("cart has expired")
	ErrCustomerNotFound  = errors.New("customer not found")
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
	ErrEmptyCart         = errors.New("cart is empty")
	ErrProductUnavailable = errors.New("product is not available")
)
