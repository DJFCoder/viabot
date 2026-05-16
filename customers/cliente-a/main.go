// Command cliente-a is a ViaBot entry point for a specific customer.
//
// It wires the Hexagonal Architecture layers:
//
//	Infrastructure → SDK factories (database, repositories)
//	Application   → Domain services (order, cart, product)
//	Domain        → Port interfaces (repository, payment gateway)
//
// Requirements:
//   - CGO_ENABLED=1
//   - gcc (mattn/go-sqlite3 is a CGO library)
package main

import (
	"context"
	"fmt"
	"log"

	"viabot.stream/sdk"
	"viabot.stream/sdk/application"
	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// Stub payment gateway — allows wiring the OrderService without a real
// payment provider. Replace with mercadopago.Adapter in production.
// ---------------------------------------------------------------------------

// stubPaymentGateway is a no-op PaymentGateway for development.
// It always returns a mocked Pix payment with 24h expiry.
type stubPaymentGateway struct{}

func (g *stubPaymentGateway) CreatePayment(
	ctx context.Context, order *domain.Order,
) (*domain.Payment, error) {
	// Stub: creates a fake Pix payment for testing
	payment := domain.NewPayment(
		fmt.Sprintf("pay_stub_%s", order.ID().Value()),
		order.ID(),
		order.Total(),
		domain.PaymentPix,
		"https://stub.payment/checkout",
		order.CreatedAt().Add(24*domain.CartExpirationHours*60*60*1000000000),
	)
	payment.Confirm(order.Total())
	return payment, nil
}

func (g *stubPaymentGateway) GetPaymentStatus(
	ctx context.Context, paymentID string,
) (domain.PaymentStatus, error) {
	return domain.PaymentApproved, nil
}

func (g *stubPaymentGateway) HandleWebhook(
	ctx context.Context, payload []byte,
) (*domain.WebhookEvent, error) {
	// Stub: not implemented for dev
	return nil, fmt.Errorf("webhook not implemented in stub gateway")
}

// ---------------------------------------------------------------------------
// Application entry point
// ---------------------------------------------------------------------------

func main() {
	ctx := context.Background()
	log.SetPrefix("[viabot] ")
	log.SetFlags(log.Lshortfile | log.Ltime)

	// ------------------------------------------------------------------
	// 1. Infrastructure — database
	// ------------------------------------------------------------------

	database, err := sdk.OpenDatabase(sdk.DefaultDatabaseDSN())
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if err := sdk.RunMigrations(database); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Println("database migrated successfully")

	// ------------------------------------------------------------------
	// 2. Infrastructure — repositories (via SDK public factories)
	// ------------------------------------------------------------------

	customerRepository := sdk.NewCustomerRepository(database)
	productRepository := sdk.NewProductRepository(database)
	cartRepository := sdk.NewCartRepository(database)
	orderRepository := sdk.NewOrderRepository(database)

	// ------------------------------------------------------------------
	// 3. Infrastructure — payment gateway (stub for development)
	// ------------------------------------------------------------------

	paymentGateway := &stubPaymentGateway{}

	// ------------------------------------------------------------------
	// 4. Application — domain services wired with repositories
	// ------------------------------------------------------------------

	productService := application.NewProductService(productRepository)
	_ = cartRepository
	cartService := application.NewCartService(cartRepository, productRepository)
	_ = orderRepository
	orderService := application.NewOrderService(
		orderRepository,
		cartRepository,
		paymentGateway,
	)

	// ------------------------------------------------------------------
	// 5. Demo: verify wiring with a smoke test
	// ------------------------------------------------------------------

	demoProducts(ctx, productService)
	demoCartFlow(ctx, cartService, productService, customerRepository)
	demoOrderFlow(ctx, orderService, cartService, customerRepository)

	log.Println("ViaBot cliente-a: initialization complete")
}

// ---------------------------------------------------------------------------
// Demo functions — simulate WhatsApp customer interactions
// ---------------------------------------------------------------------------

func demoProducts(ctx context.Context, productService *application.ProductService) {
	products, err := productService.ListProducts(ctx)
	if err != nil {
		log.Printf("list products: %v", err)
		return
	}
	log.Printf("product catalog: %d products available", products.Count())
}

func demoCartFlow(
	ctx context.Context,
	cartService *application.CartService,
	productService *application.ProductService,
	customerRepository domain.CustomerRepository,
) {
	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Cliente Teste",
	)
	if err := customerRepository.Save(ctx, customer); err != nil {
		log.Printf("save customer: %v", err)
	}

	products, err := productService.ListProducts(ctx)
	if err != nil || products.Count() == 0 {
		log.Printf("no products to add to cart: %v", err)
		return
	}

	firstProduct := products.Items()[0]
	if err := cartService.AddToCart(ctx, customer.ID(), firstProduct.ID(), 2); err != nil {
		log.Printf("add to cart: %v", err)
	}
	log.Printf("added 2x %s to cart", firstProduct.Name())
}

func demoOrderFlow(
	ctx context.Context,
	orderService *application.OrderService,
	cartService *application.CartService,
	customerRepository domain.CustomerRepository,
) {
	customer, err := customerRepository.FindByID(
		ctx, domain.NewCustomerID("5511999999999@s.whatsapp.net"),
	)
	if err != nil {
		log.Printf("find customer: %v", err)
		return
	}

	cart, err := cartService.GetCart(ctx, customer.ID())
	if err != nil {
		log.Printf("get cart: %v", err)
		return
	}

	order, err := orderService.PlaceOrder(ctx, customer.ID(), cart.ID())
	if err != nil {
		log.Printf("place order: %v", err)
		return
	}
	log.Printf("order placed: %s (total: %d %s)", order.ID().Value(), order.Total().Amount(), order.Total().Currency())
}
