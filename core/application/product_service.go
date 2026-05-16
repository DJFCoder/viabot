package application

import (
	"context"
	"fmt"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// Application Service: Product
// Handles product catalog queries.
// Rule 8 compliance: 1 instance variable (productRepo).
// ---------------------------------------------------------------------------

// ProductService handles product catalog use cases.
type ProductService struct {
	productRepo domain.ProductRepository
}

// NewProductService creates a ProductService with injected dependencies (DIP).
func NewProductService(productRepo domain.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

// ListProducts returns the full product catalog.
func (s *ProductService) ListProducts(ctx context.Context) (domain.ProductCollection, error) {
	products, err := s.productRepo.FindAll(ctx)
	if err != nil {
		return domain.ProductCollection{}, fmt.Errorf("list products: %w", err)
	}
	return domain.NewProductCollection(products), nil
}

// SearchProducts searches the catalog by name.
func (s *ProductService) SearchProducts(ctx context.Context, term string) (domain.ProductCollection, error) {
	products, err := s.productRepo.FindAll(ctx)
	if err != nil {
		return domain.ProductCollection{}, fmt.Errorf("search products: %w", err)
	}
	collection := domain.NewProductCollection(products)
	return collection.Search(term), nil
}

// FilterByCategory filters the catalog by category.
func (s *ProductService) FilterByCategory(ctx context.Context, category string) (domain.ProductCollection, error) {
	products, err := s.productRepo.FindAll(ctx)
	if err != nil {
		return domain.ProductCollection{}, fmt.Errorf("filter products: %w", err)
	}
	collection := domain.NewProductCollection(products)
	return collection.FilterByCategory(category), nil
}

// GetProduct retrieves a single product by ID.
func (s *ProductService) GetProduct(ctx context.Context, productID domain.ProductID) (*domain.Product, error) {
	return s.productRepo.FindByID(ctx, productID)
}
