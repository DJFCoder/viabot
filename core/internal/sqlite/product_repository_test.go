package sqlite

import (
	"context"
	"testing"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// ProductRepository — SQLite adapter integration tests
// ---------------------------------------------------------------------------

func makeTestProduct(t *testing.T, id, name, category string, price int64, available bool) *domain.Product {
	t.Helper()
	money, err := domain.NewMoney(price, "BRL")
	if err != nil {
		t.Fatalf("NewMoney failed: %v", err)
	}
	return domain.NewProduct(
		domain.NewProductID(id),
		name,
		"Description for "+name,
		money,
		category,
		available,
		"https://example.com/img/"+id+".jpg",
	)
}

func TestProductRepositorySaveAndFindByID(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	product := makeTestProduct(t, "prod_1", "T-Shirt", "clothing", 2990, true)
	if err := repository.Save(ctx, product); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	found, err := repository.FindByID(ctx, product.ID())
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name() != "T-Shirt" {
		t.Errorf("expected T-Shirt, got %s", found.Name())
	}
	if !found.IsAvailable() {
		t.Error("expected product to be available")
	}
	if found.Price().Amount() != 2990 {
		t.Errorf("expected price 2990, got %d", found.Price().Amount())
	}
}

func TestProductRepositoryFindByIDNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	_, err := repository.FindByID(ctx, domain.NewProductID("nonexistent"))
	if err != domain.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestProductRepositoryFindAll(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	_ = repository.Save(ctx, makeTestProduct(t, "p1", "Shirt", "clothing", 2990, true))
	_ = repository.Save(ctx, makeTestProduct(t, "p2", "Mug", "accessories", 1990, true))
	_ = repository.Save(ctx, makeTestProduct(t, "p3", "Hat", "clothing", 990, false))

	products, err := repository.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(products) != 3 {
		t.Errorf("expected 3 products, got %d", len(products))
	}
}

func TestProductRepositoryFindByCategory(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	_ = repository.Save(ctx, makeTestProduct(t, "p1", "Shirt", "clothing", 2990, true))
	_ = repository.Save(ctx, makeTestProduct(t, "p2", "Pants", "clothing", 5990, true))
	_ = repository.Save(ctx, makeTestProduct(t, "p3", "Mug", "accessories", 1990, true))

	clothing, err := repository.FindByCategory(ctx, "clothing")
	if err != nil {
		t.Fatalf("FindByCategory failed: %v", err)
	}
	if len(clothing) != 2 {
		t.Errorf("expected 2 clothing products, got %d", len(clothing))
	}
}

func TestProductRepositoryFindByCategoryEmpty(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	products, err := repository.FindByCategory(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("FindByCategory failed: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestProductRepositoryUpdateAvailability(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	product := makeTestProduct(t, "prod_1", "T-Shirt", "clothing", 2990, true)
	_ = repository.Save(ctx, product)

	if err := repository.UpdateAvailability(ctx, product.ID(), false); err != nil {
		t.Fatalf("UpdateAvailability failed: %v", err)
	}

	found, _ := repository.FindByID(ctx, product.ID())
	if found.IsAvailable() {
		t.Error("expected product to be unavailable after update")
	}
}

func TestProductRepositoryUpdateAvailabilityNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	err := repository.UpdateAvailability(ctx, domain.NewProductID("nonexistent"), false)
	if err != domain.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestProductRepositoryUpdateExisting(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewProductRepository(database)
	ctx := context.Background()

	product := makeTestProduct(t, "prod_1", "Old Name", "clothing", 2990, true)
	_ = repository.Save(ctx, product)

	updated := makeTestProduct(t, "prod_1", "New Name", "electronics", 4990, false)
	if err := repository.Save(ctx, updated); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	found, _ := repository.FindByID(ctx, updated.ID())
	if found.Name() != "New Name" {
		t.Errorf("expected New Name, got %s", found.Name())
	}
	if found.Category() != "electronics" {
		t.Errorf("expected electronics category, got %s", found.Category())
	}
	if found.IsAvailable() {
		t.Error("expected product to be unavailable")
	}
}
