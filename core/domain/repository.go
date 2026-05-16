package domain

import "context"

// ---------------------------------------------------------------------------
// Ports: Repositories
// Abstraction for data persistence — swap SQLite (dev) ↔ pgx (prod).
// ---------------------------------------------------------------------------

// OrderRepository persists Order aggregates.
type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id OrderID) (*Order, error)
	FindByCustomer(ctx context.Context, customerID CustomerID) ([]*Order, error)
	FindPendingByCustomer(ctx context.Context, customerID CustomerID) (*Order, error)
	UpdateStatus(ctx context.Context, id OrderID, status OrderStatus) error
}

// ProductRepository persists Product entities.
type ProductRepository interface {
	Save(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id ProductID) (*Product, error)
	FindAll(ctx context.Context) ([]Product, error)
	FindByCategory(ctx context.Context, category string) ([]Product, error)
	UpdateAvailability(ctx context.Context, id ProductID, available bool) error
}

// CartRepository persists Cart aggregates.
type CartRepository interface {
	Save(ctx context.Context, cart *Cart) error
	FindByCustomer(ctx context.Context, customerID CustomerID) (*Cart, error)
	Delete(ctx context.Context, id CartID) error
}

// CustomerRepository persists Customer entities.
type CustomerRepository interface {
	Save(ctx context.Context, customer *Customer) error
	FindByID(ctx context.Context, id CustomerID) (*Customer, error)
	FindByPhone(ctx context.Context, phone PhoneNumber) (*Customer, error)
}
