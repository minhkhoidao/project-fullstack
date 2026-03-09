package model

import "time"

type InventoryItem struct {
	ID               string    `json:"id"`
	ProductVariantID string    `json:"product_variant_id"`
	Warehouse        string    `json:"warehouse"`
	Quantity         int       `json:"quantity"`
	Reserved         int       `json:"reserved"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UpdateStockRequest struct {
	ProductVariantID string `json:"product_variant_id"`
	Warehouse        string `json:"warehouse"`
	Quantity         int    `json:"quantity"`
}

type ReserveRequest struct {
	ProductVariantID string `json:"product_variant_id"`
	Quantity         int    `json:"quantity"`
}
