package sqlite

import (
	"context"
	"testing"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// CustomerRepository — SQLite adapter integration tests
// ---------------------------------------------------------------------------

func TestCustomerRepositorySaveAndFindByID(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewCustomerRepository(database)
	ctx := context.Background()

	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"John Doe",
	)

	if err := repository.Save(ctx, customer); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	found, err := repository.FindByID(ctx, customer.ID())
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name() != "John Doe" {
		t.Errorf("expected name John Doe, got %s", found.Name())
	}
	if found.Phone().CountryCode() != 55 {
		t.Errorf("expected country code 55, got %d", found.Phone().CountryCode())
	}
	if found.Phone().Number() != "11999999999" {
		t.Errorf("expected phone 11999999999, got %s", found.Phone().Number())
	}
}

func TestCustomerRepositoryFindByIDNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewCustomerRepository(database)
	ctx := context.Background()

	_, err := repository.FindByID(ctx, domain.NewCustomerID("nonexistent"))
	if err != domain.ErrCustomerNotFound {
		t.Fatalf("expected ErrCustomerNotFound, got %v", err)
	}
}

func TestCustomerRepositoryFindByPhone(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewCustomerRepository(database)
	ctx := context.Background()

	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Jane Doe",
	)
	_ = repository.Save(ctx, customer)

	found, err := repository.FindByPhone(ctx, domain.NewPhoneNumber(55, "11999999999"))
	if err != nil {
		t.Fatalf("FindByPhone failed: %v", err)
	}
	if found.Name() != "Jane Doe" {
		t.Errorf("expected Jane Doe, got %s", found.Name())
	}
}

func TestCustomerRepositoryFindByPhoneNotFound(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewCustomerRepository(database)
	ctx := context.Background()

	_, err := repository.FindByPhone(ctx, domain.NewPhoneNumber(55, "00000000000"))
	if err != domain.ErrCustomerNotFound {
		t.Fatalf("expected ErrCustomerNotFound, got %v", err)
	}
}

func TestCustomerRepositoryUpdateExisting(t *testing.T) {
	database := openTestDatabase(t)
	repository := NewCustomerRepository(database)
	ctx := context.Background()

	customer := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"Old Name",
	)
	_ = repository.Save(ctx, customer)

	// Update: save with same ID but different name
	updated := domain.NewCustomer(
		domain.NewCustomerID("5511999999999@s.whatsapp.net"),
		domain.NewPhoneNumber(55, "11999999999"),
		"New Name",
	)
	if err := repository.Save(ctx, updated); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	found, _ := repository.FindByID(ctx, updated.ID())
	if found.Name() != "New Name" {
		t.Errorf("expected New Name after update, got %s", found.Name())
	}
}
