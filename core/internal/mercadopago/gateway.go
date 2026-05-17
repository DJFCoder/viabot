package mercadopago

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"viabot.stream/sdk/domain"
)

const defaultBaseURL = "https://api.mercadopago.com"

// ---------------------------------------------------------------------------
// Configuration & Option
// ---------------------------------------------------------------------------

type gatewayConfig struct {
	accessToken   string
	baseURL       string
	paymentMethod domain.PaymentMethod
}

// Option is a functional option for Gateway construction.
type Option func(*Gateway)

// WithBaseURL overrides the MP API base URL (for testing with httptest.Server).
func WithBaseURL(url string) Option {
	return func(g *Gateway) {
		g.config.baseURL = url
	}
}

// WithPaymentMethod sets the default payment method.
func WithPaymentMethod(m domain.PaymentMethod) Option {
	return func(g *Gateway) {
		g.config.paymentMethod = m
	}
}

// WithHTTPClient injects a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(g *Gateway) {
		g.client = client
	}
}

// ---------------------------------------------------------------------------
// Gateway
// ---------------------------------------------------------------------------

// Gateway implements domain.PaymentGateway for Mercado Pago.
type Gateway struct {
	config gatewayConfig
	client *http.Client
}

// NewGateway creates a Gateway with the given access token and options.
// Returns error if accessToken is empty.
func NewGateway(accessToken string, opts ...Option) (*Gateway, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("mercadopago: access token must not be empty")
	}
	g := &Gateway{
		config: gatewayConfig{
			accessToken: accessToken,
			baseURL:     defaultBaseURL,
		},
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g, nil
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

// doRequest creates and executes an HTTP request with auth header.
func (g *Gateway) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := g.config.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("mercadopago: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.config.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mercadopago: %w", domain.ErrPaymentFailed)
	}
	return resp, nil
}

// ---------------------------------------------------------------------------
// CreatePayment
// ---------------------------------------------------------------------------

// mpCreatePaymentResponse models the MP API create-payment response.
type mpCreatePaymentResponse struct {
	ID                float64 `json:"id"`
	Status            string  `json:"status"`
	TransactionAmount float64 `json:"transaction_amount"`
	ExternalReference string  `json:"external_reference"`
	DateOfExpiration  string  `json:"date_of_expiration"`
	PointOfInteraction *struct {
		TransactionData *struct {
			QRCodeBase64 string `json:"qr_code_base64"`
			QRCode       string `json:"qr_code"`
			TicketURL    string `json:"ticket_url"`
		} `json:"transaction_data"`
	} `json:"point_of_interaction"`
	Barcode *struct {
		Content string `json:"content"`
	} `json:"barcode"`
	TransactionDetails *struct {
		ExternalResourceURL string `json:"external_resource_url"`
	} `json:"transaction_details"`
}

// paymentMethodID maps domain payment methods to MP API identifiers.
func paymentMethodID(method domain.PaymentMethod) string {
	switch method {
	case domain.PaymentPix:
		return "pix"
	case domain.PaymentBoleto:
		return "bolbradesco"
	default:
		return ""
	}
}

// buildPaymentRequest constructs the JSON body for a payment request.
func (g *Gateway) buildPaymentRequest(order *domain.Order) ([]byte, error) {
	amount := float64(order.Total().Amount()) / 100.0
	orderID := order.ID().Value()
	mpMethod := paymentMethodID(g.config.paymentMethod)
	payload := map[string]interface{}{
		"transaction_amount": amount,
		"description":        "Order " + orderID,
		"payment_method_id":  mpMethod,
		"payer": map[string]interface{}{
			"email": "buyer@example.com",
		},
		"external_reference": orderID,
	}
	return json.Marshal(payload)
}

// extractCheckoutLink returns the payment URL from a MP response.
func extractCheckoutLink(resp *mpCreatePaymentResponse) string {
	if resp.PointOfInteraction != nil && resp.PointOfInteraction.TransactionData != nil {
		return resp.PointOfInteraction.TransactionData.TicketURL
	}
	if resp.TransactionDetails != nil {
		return resp.TransactionDetails.ExternalResourceURL
	}
	return ""
}

// attachPixDetails sets QR code fields when the response has Pix data.
func attachPixDetails(payment *domain.Payment, resp *mpCreatePaymentResponse) {
	if resp.PointOfInteraction == nil || resp.PointOfInteraction.TransactionData == nil {
		return
	}
	payment.AttachPixDetails(
		resp.PointOfInteraction.TransactionData.QRCodeBase64,
		resp.PointOfInteraction.TransactionData.QRCode,
	)
}

// attachBoletoDetails sets barcode fields when the response has Boleto data.
func attachBoletoDetails(payment *domain.Payment, resp *mpCreatePaymentResponse) {
	if resp.Barcode == nil {
		return
	}
	boletoURL := ""
	if resp.TransactionDetails != nil {
		boletoURL = resp.TransactionDetails.ExternalResourceURL
	}
	payment.AttachBoletoDetails(resp.Barcode.Content, boletoURL)
}

// parsePaymentResponse decodes the MP API response and builds a domain.Payment.
func (g *Gateway) parsePaymentResponse(r io.Reader, order *domain.Order) (*domain.Payment, error) {
	var resp mpCreatePaymentResponse
	if err := json.NewDecoder(r).Decode(&resp); err != nil {
		return nil, fmt.Errorf("mercadopago: %w", err)
	}
	id := fmt.Sprintf("%.0f", resp.ID)
	expiresAt, _ := time.Parse(time.RFC3339, resp.DateOfExpiration)
	payment := domain.NewPayment(id, order.ID(), order.Total(), g.config.paymentMethod, extractCheckoutLink(&resp), expiresAt)
	attachPixDetails(payment, &resp)
	attachBoletoDetails(payment, &resp)
	return payment, nil
}

// CreatePayment creates a payment on Mercado Pago for the given order.
func (g *Gateway) CreatePayment(ctx context.Context, order *domain.Order) (*domain.Payment, error) {
	if g.config.paymentMethod == domain.PaymentCard {
		return nil, ErrInvalidPaymentMethod
	}
	body, err := g.buildPaymentRequest(order)
	if err != nil {
		return nil, fmt.Errorf("mercadopago: %w", err)
	}
	resp, err := g.doRequest(ctx, http.MethodPost, "/v1/payments", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("mercadopago: %w (status: %d)", domain.ErrPaymentFailed, resp.StatusCode)
	}
	return g.parsePaymentResponse(resp.Body, order)
}

// ---------------------------------------------------------------------------
// GetPaymentStatus
// ---------------------------------------------------------------------------

// mpPaymentStatusResponse models the MP API payment status response.
type mpPaymentStatusResponse struct {
	Status            string `json:"status"`
	ExternalReference string `json:"external_reference"`
}

// getPaymentData fetches payment status and external reference from MP API.
func (g *Gateway) getPaymentData(ctx context.Context, paymentID string) (domain.PaymentStatus, string, error) {
	resp, err := g.doRequest(ctx, http.MethodGet, "/v1/payments/"+paymentID, http.NoBody)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return "", "", fmt.Errorf("mercadopago: %w (status: %d)", domain.ErrPaymentFailed, resp.StatusCode)
	}
	var payResp mpPaymentStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&payResp); err != nil {
		return "", "", fmt.Errorf("mercadopago: %w", err)
	}
	return MapMPStatus(payResp.Status), payResp.ExternalReference, nil
}

// GetPaymentStatus retrieves the current payment status from Mercado Pago.
func (g *Gateway) GetPaymentStatus(ctx context.Context, paymentID string) (domain.PaymentStatus, error) {
	status, _, err := g.getPaymentData(ctx, paymentID)
	return status, err
}

// ---------------------------------------------------------------------------
// HandleWebhook
// ---------------------------------------------------------------------------

// HandleWebhook processes an incoming Mercado Pago webhook payload.
func (g *Gateway) HandleWebhook(ctx context.Context, payload []byte) (*domain.WebhookEvent, error) {
	var wh struct {
		Action string `json:"action"`
		Data   struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &wh); err != nil {
		return nil, fmt.Errorf("mercadopago: %w", err)
	}
	status, externalRef, err := g.getPaymentData(ctx, wh.Data.ID)
	if err != nil {
		return nil, err
	}
	return domain.NewWebhookEvent(wh.Action, wh.Data.ID, domain.NewOrderID(externalRef), status), nil
}
