package mercadopago_test

import (
	"testing"

	"viabot.stream/sdk/domain"
	"viabot.stream/sdk/internal/mercadopago"
)

// ---------------------------------------------------------------------------
// Status mapping — table-driven test covering all known MP statuses plus
// unknown (fail-safe default).
// ---------------------------------------------------------------------------

func TestMapMPStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mpStatus string
		want     domain.PaymentStatus
	}{
		{name: "approved maps to PaymentApproved", mpStatus: "approved", want: domain.PaymentApproved},
		{name: "rejected maps to PaymentRejected", mpStatus: "rejected", want: domain.PaymentRejected},
		{name: "pending maps to PaymentPending", mpStatus: "pending", want: domain.PaymentPending},
		{name: "cancelled maps to PaymentCancelled", mpStatus: "cancelled", want: domain.PaymentCancelled},
		{name: "refunded maps to PaymentRefunded", mpStatus: "refunded", want: domain.PaymentRefunded},
		{name: "in_process maps to PaymentPending", mpStatus: "in_process", want: domain.PaymentPending},
		{name: "authorized maps to PaymentPending", mpStatus: "authorized", want: domain.PaymentPending},
		{name: "unknown status maps to PaymentPending (fail-safe)", mpStatus: "unknown_xyz", want: domain.PaymentPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mercadopago.MapMPStatus(tt.mpStatus)
			if got != tt.want {
				t.Errorf("MapMPStatus(%q) = %q, want %q", tt.mpStatus, got, tt.want)
			}
		})
	}
}
