package model

import "time"

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

// Order represents a placed customer order.
type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Status          OrderStatus `json:"status"`
	TotalPrice      float64     `json:"total_price"`
	ShippingAddress Address     `json:"shipping_address"`
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// Address represents a shipping address embedded in an order.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

// OrderItem represents a single line item within an order.
type OrderItem struct {
	ID               string  `json:"id"`
	OrderID          string  `json:"order_id"`
	ProductVariantID string  `json:"product_variant_id"`
	ProductName      string  `json:"product_name"`
	VariantInfo      string  `json:"variant_info"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
}

// CreateOrderRequest is the payload for placing a new order.
// Items are sourced from the user's cart.
type CreateOrderRequest struct {
	ShippingAddress Address `json:"shipping_address"`
}

// Payment represents a payment record associated with an order.
type Payment struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	Amount     float64   `json:"amount"`
	Method     string    `json:"method"`
	Status     string    `json:"status"`
	ExternalID string    `json:"external_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
