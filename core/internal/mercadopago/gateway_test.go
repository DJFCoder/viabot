package mercadopago_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"viabot.stream/sdk/domain"
	"viabot.stream/sdk/internal/mercadopago"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// makeTestOrder creates a valid domain.Order for payment tests.
func makeTestOrder(t *testing.T) *domain.Order {
	t.Helper()

	customerID := domain.NewCustomerID("5511999999999@s.whatsapp.net")
	price, err := domain.NewMoney(15000, "BRL")
	if err != nil {
		t.Fatalf("NewMoney failed: %v", err)
	}
	pid := domain.NewProductID("prod_1")
	item, err := domain.NewOrderItem(pid, "Test Product", price, 1)
	if err != nil {
		t.Fatalf("NewOrderItem failed: %v", err)
	}
	items := domain.NewOrderItemCollection([]domain.OrderItem{item})
	orderID := domain.NewOrderID("ord_mp_test_1")
	order, err := domain.NewOrder(orderID, customerID, items)
	if err != nil {
		t.Fatalf("NewOrder failed: %v", err)
	}
	return order
}

// mpResponsePix returns a minimal MP API JSON response for a Pix payment.
func mpResponsePix(orderID string) map[string]interface{} {
	return map[string]interface{}{
		"id":                 123456789,
		"status":             "pending",
		"status_detail":      "pending_waiting_transfer",
		"payment_method_id":  "pix",
		"transaction_amount": 150.00,
		"external_reference": orderID,
		"date_of_expiration": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"point_of_interaction": map[string]interface{}{
			"transaction_data": map[string]interface{}{
				"qr_code_base64": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAA=",
				"qr_code":        "00020126580014BR.GOV.BCB.PIX0136123",
				"ticket_url":     "https://mercadopago.com.br/pay/123456789",
			},
		},
	}
}

// mpResponseBoleto returns a minimal MP API JSON response for a Boleto payment.
func mpResponseBoleto(orderID string) map[string]interface{} {
	return map[string]interface{}{
		"id":                 123456790,
		"status":             "pending",
		"status_detail":      "pending_waiting_payment",
		"payment_method_id":  "bolbradesco",
		"transaction_amount": 150.00,
		"external_reference": orderID,
		"date_of_expiration": time.Now().Add(72 * time.Hour).Format(time.RFC3339),
		"barcode": map[string]interface{}{
			"content": "34191790010123456789012345678901234567890123",
		},
		"transaction_details": map[string]interface{}{
			"external_resource_url": "https://mercadopago.com.br/pay/123456790/boleto",
		},
	}
}

// ---------------------------------------------------------------------------
// NewGateway — constructor tests
// ---------------------------------------------------------------------------

func TestMercadoPagoGatewayNewValid(t *testing.T) {
	t.Parallel()

	gateway, err := mercadopago.NewGateway("test_access_token_12345")
	if err != nil {
		t.Fatalf("NewGateway failed: %v", err)
	}
	if gateway == nil {
		t.Fatal("expected non-nil gateway")
	}
}

func TestMercadoPagoGatewayNewEmptyToken(t *testing.T) {
	t.Parallel()

	gateway, err := mercadopago.NewGateway("")
	if err == nil {
		t.Fatal("expected error for empty access token")
	}
	if gateway != nil {
		t.Fatal("expected nil gateway on error")
	}
}

// ---------------------------------------------------------------------------
// CreatePayment — happy path
// ---------------------------------------------------------------------------

func TestMercadoPagoGatewayCreatePaymentPix(t *testing.T) {
	t.Parallel()

	order := makeTestOrder(t)
	server := newMPServer(t, http.StatusCreated, mpResponsePix(order.ID().Value()))
	defer server.Close()

	gateway := newTestGateway(t, server.URL, domain.PaymentPix)

	payment, err := gateway.CreatePayment(context.Background(), order)
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if payment == nil {
		t.Fatal("expected non-nil payment")
	}
	if payment.Status() != domain.PaymentPending {
		t.Errorf("expected status pending, got %s", payment.Status())
	}
	if payment.QRCode() == "" {
		t.Error("expected non-empty QR code")
	}
	if payment.QRCodeText() == "" {
		t.Error("expected non-empty QR code text")
	}
	if payment.CheckoutLink() == "" {
		t.Error("expected non-empty checkout link")
	}
}

func TestMercadoPagoGatewayCreatePaymentBoleto(t *testing.T) {
	t.Parallel()

	order := makeTestOrder(t)
	server := newMPServer(t, http.StatusCreated, mpResponseBoleto(order.ID().Value()))
	defer server.Close()

	gateway := newTestGateway(t, server.URL, domain.PaymentBoleto)

	payment, err := gateway.CreatePayment(context.Background(), order)
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if payment == nil {
		t.Fatal("expected non-nil payment")
	}
	if payment.Status() != domain.PaymentPending {
		t.Errorf("expected status pending, got %s", payment.Status())
	}
	if payment.BoletoBarcode() == "" {
		t.Error("expected non-empty boleto barcode")
	}
	if payment.BoletoURL() == "" {
		t.Error("expected non-empty boleto URL")
	}
	if payment.CheckoutLink() == "" {
		t.Error("expected non-empty checkout link")
	}
}

// ---------------------------------------------------------------------------
// CreatePayment — error cases
// ---------------------------------------------------------------------------

func TestMercadoPagoGatewayCreatePaymentUnsupportedMethod(t *testing.T) {
	t.Parallel()

	order := makeTestOrder(t)
	gateway := newTestGateway(t, "http://localhost:1", domain.PaymentCard)

	_, err := gateway.CreatePayment(context.Background(), order)
	if err == nil {
		t.Fatal("expected error for unsupported payment method")
	}
	if !errors.Is(err, mercadopago.ErrInvalidPaymentMethod) {
		t.Errorf("expected ErrInvalidPaymentMethod, got %v", err)
	}
}

func TestMercadoPagoGatewayCreatePaymentUnauthorized(t *testing.T) {
	t.Parallel()

	order := makeTestOrder(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid access token","status":401}`))
	}))
	defer server.Close()

	gateway := newTestGateway(t, server.URL, domain.PaymentPix)

	_, err := gateway.CreatePayment(context.Background(), order)
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
	if !errors.Is(err, domain.ErrPaymentFailed) {
		t.Errorf("expected error wrapping ErrPaymentFailed, got %v", err)
	}
}

func TestMercadoPagoGatewayCreatePaymentTimeout(t *testing.T) {
	t.Parallel()

	order := makeTestOrder(t)
	// Server that hangs to trigger a timeout.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	gateway := newTestGatewayWithClient(t, server.URL, domain.PaymentPix, &http.Client{Timeout: 1 * time.Millisecond})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := gateway.CreatePayment(ctx, order)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, domain.ErrPaymentFailed) {
		t.Errorf("expected error wrapping ErrPaymentFailed, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetPaymentStatus — status mapping
// ---------------------------------------------------------------------------

func TestMercadoPagoGatewayGetPaymentStatusApproved(t *testing.T) {
	t.Parallel()

	testGetPaymentStatus(t, "approved", domain.PaymentApproved)
}

func TestMercadoPagoGatewayGetPaymentStatusRejected(t *testing.T) {
	t.Parallel()

	testGetPaymentStatus(t, "rejected", domain.PaymentRejected)
}

func TestMercadoPagoGatewayGetPaymentStatusPending(t *testing.T) {
	t.Parallel()

	testGetPaymentStatus(t, "pending", domain.PaymentPending)
}

func TestMercadoPagoGatewayGetPaymentStatusUnknown(t *testing.T) {
	t.Parallel()

	// Unknown MP statuses default to PaymentPending (fail-safe).
	testGetPaymentStatus(t, "some_garbage_status", domain.PaymentPending)
}

func TestMercadoPagoGatewayGetPaymentStatusNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"payment not found","status":404}`))
	}))
	defer server.Close()

	gateway := newTestGateway(t, server.URL, domain.PaymentPix)

	_, err := gateway.GetPaymentStatus(context.Background(), "nonexistent_id")
	if err == nil {
		t.Fatal("expected error for non-existent payment")
	}
	if !errors.Is(err, domain.ErrPaymentFailed) {
		t.Errorf("expected error wrapping ErrPaymentFailed, got %v", err)
	}
}

// testGetPaymentStatus is a helper for GetPaymentStatus tests.
func testGetPaymentStatus(t *testing.T, mpStatus string, expected domain.PaymentStatus) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     123456789,
			"status": mpStatus,
		})
	}))
	defer server.Close()

	gateway := newTestGateway(t, server.URL, domain.PaymentPix)

	status, err := gateway.GetPaymentStatus(context.Background(), "123456789")
	if err != nil {
		t.Fatalf("GetPaymentStatus failed: %v", err)
	}
	if status != expected {
		t.Errorf("GetPaymentStatus = %q, want %q", status, expected)
	}
}

// ---------------------------------------------------------------------------
// HandleWebhook
// ---------------------------------------------------------------------------

func TestMercadoPagoGatewayHandleWebhookValid(t *testing.T) {
	t.Parallel()

	order := makeTestOrder(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                 123456789,
			"status":             "approved",
			"external_reference": order.ID().Value(),
		})
	}))
	defer server.Close()

	gateway := newTestGateway(t, server.URL, domain.PaymentPix)

	// MP webhook payload with payment ID in data.id.
	payload := []byte(`{
		"action": "payment.updated",
		"type": "payment",
		"data": {"id": "123456789"}
	}`)

	event, err := gateway.HandleWebhook(context.Background(), payload)
	if err != nil {
		t.Fatalf("HandleWebhook failed: %v", err)
	}
	if event == nil {
		t.Fatal("expected non-nil webhook event")
	}
	if event.Action() != "payment.updated" {
		t.Errorf("expected action 'payment.updated', got %q", event.Action())
	}
	if event.PaymentID() != "123456789" {
		t.Errorf("expected paymentID '123456789', got %q", event.PaymentID())
	}
	if event.OrderID().Value() != order.ID().Value() {
		t.Errorf("expected orderID %q, got %q", order.ID().Value(), event.OrderID().Value())
	}
	if event.Status() != domain.PaymentApproved {
		t.Errorf("expected status approved, got %s", event.Status())
	}
}

func TestMercadoPagoGatewayHandleWebhookMalformed(t *testing.T) {
	t.Parallel()

	gateway := newTestGateway(t, "http://localhost:1", domain.PaymentPix)

	_, err := gateway.HandleWebhook(context.Background(), []byte(`{invalid json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

// ---------------------------------------------------------------------------
// Gateway test helpers
// ---------------------------------------------------------------------------

// newMPServer creates an httptest.Server that responds with the given status
// code and JSON body (mimicking the Mercado Pago API).
func newMPServer(t *testing.T, statusCode int, body interface{}) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

// newTestGateway creates a Gateway configured for the given test server URL
// and payment method, with a fast HTTP client.
func newTestGateway(t *testing.T, baseURL string, method domain.PaymentMethod) domain.PaymentGateway {
	t.Helper()

	gateway, err := mercadopago.NewGateway(
		"test_token_12345",
		mercadopago.WithBaseURL(baseURL),
		mercadopago.WithPaymentMethod(method),
		mercadopago.WithHTTPClient(&http.Client{Timeout: 2 * time.Second}),
	)
	if err != nil {
		t.Fatalf("NewGateway failed: %v", err)
	}
	return gateway
}

// newTestGatewayWithClient creates a Gateway with a custom HTTP client.
func newTestGatewayWithClient(
	t *testing.T,
	baseURL string,
	method domain.PaymentMethod,
	client *http.Client,
) domain.PaymentGateway {
	t.Helper()

	gateway, err := mercadopago.NewGateway(
		"test_token_12345",
		mercadopago.WithBaseURL(baseURL),
		mercadopago.WithPaymentMethod(method),
		mercadopago.WithHTTPClient(client),
	)
	if err != nil {
		t.Fatalf("NewGateway failed: %v", err)
	}
	return gateway
}
