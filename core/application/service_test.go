package application

import (
	"context"
	"sync"
	"testing"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// In-memory mock repositories — satisfy domain interfaces for testing
// ---------------------------------------------------------------------------

type inMemoryCartRepository struct {
	mu    sync.Mutex
	carts map[string]*domain.Cart
}

func newInMemoryCartRepository() *inMemoryCartRepository {
	return &inMemoryCartRepository{carts: make(map[string]*domain.Cart)}
}

func (r *inMemoryCartRepository) Save(_ context.Context, cart *domain.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.carts[cart.ID().Value()] = cart
	return nil
}

func (r *inMemoryCartRepository) FindByCustomer(_ context.Context, customerID domain.CustomerID) (*domain.Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cart := range r.carts {
		if cart.CustomerID() == customerID {
			return cart, nil
		}
	}
	return nil, domain.ErrCartNotFound
}

func (r *inMemoryCartRepository) Delete(_ context.Context, id domain.CartID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.carts, id.Value())
	return nil
}

type inMemoryOrderRepository struct {
	mu     sync.Mutex
	orders map[string]*domain.Order
}

func newInMemoryOrderRepository() *inMemoryOrderRepository {
	return &inMemoryOrderRepository{orders: make(map[string]*domain.Order)}
}

func (r *inMemoryOrderRepository) Save(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID().Value()] = order
	return nil
}

func (r *inMemoryOrderRepository) FindByID(_ context.Context, id domain.OrderID) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, ok := r.orders[id.Value()]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

func (r *inMemoryOrderRepository) FindByCustomer(_ context.Context, customerID domain.CustomerID) ([]*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*domain.Order
	for _, order := range r.orders {
		if order.CustomerID() == customerID {
			result = append(result, order)
		}
	}
	return result, nil
}

func (r *inMemoryOrderRepository) FindPendingByCustomer(_ context.Context, customerID domain.CustomerID) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, order := range r.orders {
		if order.CustomerID() == customerID && order.Status() == domain.OrderStatusPending {
			return order, nil
		}
	}
	return nil, domain.ErrOrderNotFound
}

func (r *inMemoryOrderRepository) UpdateStatus(_ context.Context, id domain.OrderID, status domain.OrderStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, ok := r.orders[id.Value()]
	if !ok {
		return domain.ErrOrderNotFound
	}
	// Since Order fields are unexported, we simulate by replacing
	// In a real scenario the repository would use SQL UPDATE.
	r.orders[id.Value()] = order
	return nil
}

type inMemoryProductRepository struct {
	mu       sync.Mutex
	products map[string]*domain.Product
}

func newInMemoryProductRepository() *inMemoryProductRepository {
	return &inMemoryProductRepository{products: make(map[string]*domain.Product)}
}

func (r *inMemoryProductRepository) Save(_ context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[product.ID().Value()] = product
	return nil
}

func (r *inMemoryProductRepository) FindByID(_ context.Context, id domain.ProductID) (*domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	product, ok := r.products[id.Value()]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return product, nil
}

func (r *inMemoryProductRepository) FindAll(_ context.Context) ([]domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []domain.Product
	for _, p := range r.products {
		result = append(result, *p)
	}
	return result, nil
}

func (r *inMemoryProductRepository) FindByCategory(_ context.Context, category string) ([]domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []domain.Product
	for _, p := range r.products {
		if p.Category() == category {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (r *inMemoryProductRepository) UpdateAvailability(_ context.Context, id domain.ProductID, available bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	product, ok := r.products[id.Value()]
	if !ok {
		return domain.ErrProductNotFound
	}
	product.UpdateAvailability(available)
	return nil
}

// ---------------------------------------------------------------------------
// Mock PaymentGateway
// ---------------------------------------------------------------------------

type mockPaymentGateway struct{}

func (m *mockPaymentGateway) CreatePayment(_ context.Context, order *domain.Order) (*domain.Payment, error) {
	return domain.NewPayment(
		"pay_mock",
		order.ID(),
		order.Total(),
		domain.PaymentPix,
		"https://mock.checkout",
		order.CreatedAt().Add(24*15),
	), nil
}

func (m *mockPaymentGateway) GetPaymentStatus(_ context.Context, paymentID string) (domain.PaymentStatus, error) {
	return domain.PaymentApproved, nil
}

func (m *mockPaymentGateway) HandleWebhook(_ context.Context, payload []byte) (*domain.WebhookEvent, error) {
	return domain.NewWebhookEvent("payment.approved", "pay_mock", domain.OrderID{}, domain.PaymentApproved), nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCartServiceAddAndGetCart(t *testing.T) {
	ctx := context.Background()
	cartRepo := newInMemoryCartRepository()
	productRepo := newInMemoryProductRepository()

	svc := NewCartService(cartRepo, productRepo)

	// Setup: create a product
	price, _ := domain.NewMoney(1990, "BRL")
	pid := domain.NewProductID("prod_1")
	product := domain.NewProduct(pid, "T-Shirt", "Cool shirt", price, "clothing", true, "")
	_ = productRepo.Save(ctx, product)

	// Add to cart
	customerID := domain.NewCustomerID("test_user")
	if err := svc.AddToCart(ctx, customerID, pid, 2); err != nil {
		t.Fatalf("AddToCart failed: %v", err)
	}

	// Retrieve cart
	cart, err := svc.GetCart(ctx, customerID)
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}
	if cart.Items().Count() != 1 {
		t.Errorf("expected 1 item, got %d", cart.Items().Count())
	}
}

func TestCartServiceRemoveFromCart(t *testing.T) {
	ctx := context.Background()
	cartRepo := newInMemoryCartRepository()
	productRepo := newInMemoryProductRepository()
	svc := NewCartService(cartRepo, productRepo)

	price, _ := domain.NewMoney(1990, "BRL")
	pid := domain.NewProductID("prod_1")
	product := domain.NewProduct(pid, "T-Shirt", "Cool shirt", price, "clothing", true, "")
	_ = productRepo.Save(ctx, product)

	customerID := domain.NewCustomerID("test_user")
	_ = svc.AddToCart(ctx, customerID, pid, 2)
	_ = svc.RemoveFromCart(ctx, customerID, pid)

	cart, _ := svc.GetCart(ctx, customerID)
	if cart.Items().Count() != 0 {
		t.Errorf("expected 0 items after removal, got %d", cart.Items().Count())
	}
}

func TestCartServiceAddUnavailableProduct(t *testing.T) {
	ctx := context.Background()
	cartRepo := newInMemoryCartRepository()
	productRepo := newInMemoryProductRepository()
	svc := NewCartService(cartRepo, productRepo)

	price, _ := domain.NewMoney(1990, "BRL")
	pid := domain.NewProductID("prod_1")
	// Product is NOT available
	product := domain.NewProduct(pid, "T-Shirt", "Cool shirt", price, "clothing", false, "")
	_ = productRepo.Save(ctx, product)

	customerID := domain.NewCustomerID("test_user")
	err := svc.AddToCart(ctx, customerID, pid, 1)
	if err == nil {
		t.Fatal("expected error adding unavailable product")
	}
}

func TestProductServiceListProducts(t *testing.T) {
	ctx := context.Background()
	productRepo := newInMemoryProductRepository()
	svc := NewProductService(productRepo)

	price1, _ := domain.NewMoney(1990, "BRL")
	price2, _ := domain.NewMoney(2990, "BRL")
	_ = productRepo.Save(ctx, domain.NewProduct(domain.NewProductID("p1"), "T-Shirt", "", price1, "clothing", true, ""))
	_ = productRepo.Save(ctx, domain.NewProduct(domain.NewProductID("p2"), "Mug", "", price2, "accessories", true, ""))

	collection, err := svc.ListProducts(ctx)
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}
	if collection.Count() != 2 {
		t.Errorf("expected 2 products, got %d", collection.Count())
	}
}

func TestProductServiceSearchProducts(t *testing.T) {
	ctx := context.Background()
	productRepo := newInMemoryProductRepository()
	svc := NewProductService(productRepo)

	price, _ := domain.NewMoney(1990, "BRL")
	_ = productRepo.Save(ctx, domain.NewProduct(domain.NewProductID("p1"), "Blue T-Shirt", "", price, "clothing", true, ""))
	_ = productRepo.Save(ctx, domain.NewProduct(domain.NewProductID("p2"), "Red Mug", "", price, "accessories", true, ""))

	// Search should be case-insensitive
	collection, err := svc.SearchProducts(ctx, "t-shirt")
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}
	if collection.Count() != 1 {
		t.Errorf("expected 1 product, got %d", collection.Count())
	}
}

func TestOrderServicePlaceOrder(t *testing.T) {
	ctx := context.Background()
	cartRepo := newInMemoryCartRepository()
	orderRepo := newInMemoryOrderRepository()
	gateway := &mockPaymentGateway{}

	svc := NewOrderService(orderRepo, cartRepo, gateway)

	// Setup: customer with cart containing items
	customerID := domain.NewCustomerID("test_user")
	cartID := domain.NewCartID("cart_1")
	cart := domain.NewCart(cartID, customerID)
	price, _ := domain.NewMoney(1000, "BRL")
	_ = cart.AddItem(domain.NewProductID("p1"), "Item 1", price, 3)
	_ = cartRepo.Save(ctx, cart)

	order, err := svc.PlaceOrder(ctx, customerID, cartID)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}
	if order.Status() != domain.OrderStatusPending {
		t.Errorf("expected pending, got %s", order.Status())
	}
	if order.Total().Amount() != 3000 {
		t.Errorf("expected total 3000, got %d", order.Total().Amount())
	}
}

func TestOrderServiceCancelOrder(t *testing.T) {
	ctx := context.Background()
	cartRepo := newInMemoryCartRepository()
	orderRepo := newInMemoryOrderRepository()
	gateway := &mockPaymentGateway{}

	svc := NewOrderService(orderRepo, cartRepo, gateway)

	// Create and place order
	customerID := domain.NewCustomerID("test_user")
	cartID := domain.NewCartID("cart_1")
	cart := domain.NewCart(cartID, customerID)
	price, _ := domain.NewMoney(1000, "BRL")
	_ = cart.AddItem(domain.NewProductID("p1"), "Item 1", price, 1)
	_ = cartRepo.Save(ctx, cart)

	order, _ := svc.PlaceOrder(ctx, customerID, cartID)

	// Cancel it
	if err := svc.CancelOrder(ctx, order.ID()); err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	cancelled, _ := svc.GetOrder(ctx, order.ID())
	if cancelled.Status() != domain.OrderStatusCancelled {
		t.Errorf("expected cancelled, got %s", cancelled.Status())
	}
}

func TestOrderServiceListCustomerOrders(t *testing.T) {
	ctx := context.Background()
	cartRepo := newInMemoryCartRepository()
	orderRepo := newInMemoryOrderRepository()
	gateway := &mockPaymentGateway{}

	svc := NewOrderService(orderRepo, cartRepo, gateway)

	customerID := domain.NewCustomerID("test_user")
	cartID := domain.NewCartID("cart_1")
	cart := domain.NewCart(cartID, customerID)
	price, _ := domain.NewMoney(1000, "BRL")
	_ = cart.AddItem(domain.NewProductID("p1"), "Item 1", price, 1)
	_ = cartRepo.Save(ctx, cart)

	_, _ = svc.PlaceOrder(ctx, customerID, cartID)

	orders, err := svc.ListCustomerOrders(ctx, customerID)
	if err != nil {
		t.Fatalf("ListCustomerOrders failed: %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orders))
	}
}
