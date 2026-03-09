package service

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/kyle/product/internal/product/model"
	"github.com/kyle/product/internal/product/repository"
)

// ProductService implements business logic for the product domain.
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService returns a ready-to-use ProductService.
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// CreateProduct validates the request, persists the product and its variants,
// then returns the fully-hydrated product.
func (s *ProductService) CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if req.BasePrice <= 0 {
		return nil, fmt.Errorf("base price must be positive")
	}

	product := &model.Product{
		ID:          generateID(),
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		BasePrice:   req.BasePrice,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	for _, vr := range req.Variants {
		variant := &model.ProductVariant{
			ID:            generateID(),
			ProductID:     product.ID,
			SKU:           vr.SKU,
			Size:          vr.Size,
			Color:         vr.Color,
			PriceOverride: vr.PriceOverride,
		}
		if err := s.repo.CreateVariant(ctx, variant); err != nil {
			return nil, fmt.Errorf("create variant %s: %w", vr.SKU, err)
		}
	}

	return s.repo.GetByID(ctx, product.ID)
}

// GetProduct retrieves a product by ID.
func (s *ProductService) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}

// GetProductBySlug retrieves a product by its URL-friendly slug.
func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (*model.Product, error) {
	p, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get product by slug: %w", err)
	}
	return p, nil
}

// ListProducts returns a paginated, filtered list of products.
func (s *ProductService) ListProducts(ctx context.Context, filter model.ProductFilter, cursor string, limit int) ([]model.Product, string, error) {
	products, next, err := s.repo.List(ctx, filter, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("list products: %w", err)
	}
	return products, next, nil
}

// UpdateProduct applies partial updates and returns the refreshed product.
func (s *ProductService) UpdateProduct(ctx context.Context, id string, req model.UpdateProductRequest) (*model.Product, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find product for update: %w", err)
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.BasePrice > 0 {
		existing.BasePrice = req.BasePrice
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.CategoryID != "" {
		existing.CategoryID = req.CategoryID
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	return s.repo.GetByID(ctx, id)
}

// DeleteProduct removes a product by ID.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

// CreateCategory persists a new category and returns it.
func (s *ProductService) CreateCategory(ctx context.Context, name, description string) (*model.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	cat := &model.Category{
		ID:          generateID(),
		Name:        name,
		Description: description,
	}

	if err := s.repo.CreateCategory(ctx, cat); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return cat, nil
}

// ListCategories returns all categories ordered by name.
func (s *ProductService) ListCategories(ctx context.Context) ([]model.Category, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}

// generateID produces a random UUID v4.
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
