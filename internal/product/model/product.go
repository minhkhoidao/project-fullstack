package model

import "time"

// Category represents a product category in the catalog.
type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ParentID    *string   `json:"parent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Product represents a fashion item in the catalog.
type Product struct {
	ID          string           `json:"id"`
	CategoryID  string           `json:"category_id"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description string           `json:"description"`
	BasePrice   float64          `json:"base_price"`
	IsActive    bool             `json:"is_active"`
	Images      []ProductImage   `json:"images"`
	Variants    []ProductVariant `json:"variants"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ProductImage holds a single image reference for a product.
type ProductImage struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text"`
	SortOrder int    `json:"sort_order"`
}

// ProductVariant represents a size/color combination of a product.
type ProductVariant struct {
	ID            string   `json:"id"`
	ProductID     string   `json:"product_id"`
	SKU           string   `json:"sku"`
	Size          string   `json:"size"`
	Color         string   `json:"color"`
	PriceOverride *float64 `json:"price_override,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateProductRequest is the payload for creating a new product.
type CreateProductRequest struct {
	CategoryID string                 `json:"category_id"`
	Name       string                 `json:"name"`
	Description string                `json:"description"`
	BasePrice  float64                `json:"base_price"`
	Variants   []CreateVariantRequest `json:"variants"`
}

// CreateVariantRequest is a nested payload for creating variants alongside a product.
type CreateVariantRequest struct {
	SKU           string   `json:"sku"`
	Size          string   `json:"size"`
	Color         string   `json:"color"`
	PriceOverride *float64 `json:"price_override,omitempty"`
}

// UpdateProductRequest is the payload for updating an existing product.
type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	BasePrice   float64 `json:"base_price"`
	IsActive    *bool   `json:"is_active"`
	CategoryID  string  `json:"category_id"`
}

// ProductFilter captures query-param based filtering for product listings.
type ProductFilter struct {
	CategoryID string  `json:"category_id"`
	Search     string  `json:"search"`
	MinPrice   *float64 `json:"min_price"`
	MaxPrice   *float64 `json:"max_price"`
	Size       string  `json:"size"`
	Color      string  `json:"color"`
}
