package model

import "time"

// Cart represents a user's shopping cart stored in Redis.
type Cart struct {
	UserID    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CartItem represents a single product variant in the cart.
type CartItem struct {
	ProductVariantID string  `json:"product_variant_id"`
	ProductName      string  `json:"product_name"`
	VariantInfo      string  `json:"variant_info"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
}

// AddItemRequest is the payload for adding an item to the cart.
type AddItemRequest struct {
	ProductVariantID string  `json:"product_variant_id"`
	ProductName      string  `json:"product_name"`
	VariantInfo      string  `json:"variant_info"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
}

// UpdateItemRequest is the payload for updating a cart item's quantity.
type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}
