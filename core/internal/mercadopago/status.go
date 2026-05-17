package mercadopago

import (
	"errors"

	"viabot.stream/sdk/domain"
)

// ErrInvalidPaymentMethod indicates the payment method is not supported.
var ErrInvalidPaymentMethod = errors.New("mercadopago: invalid or unsupported payment method")

// MapMPStatus converts a Mercado Pago status string to a domain.PaymentStatus.
// Unknown statuses default to PaymentPending (fail-safe).
func MapMPStatus(mpStatus string) domain.PaymentStatus {
	switch mpStatus {
	case "approved":
		return domain.PaymentApproved
	case "rejected":
		return domain.PaymentRejected
	case "pending", "in_process", "authorized":
		return domain.PaymentPending
	case "cancelled":
		return domain.PaymentCancelled
	case "refunded":
		return domain.PaymentRefunded
	default:
		return domain.PaymentPending
	}
}
