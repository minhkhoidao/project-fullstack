package model

import "time"

// DashboardStats holds high-level platform metrics.
type DashboardStats struct {
	TotalUsers    int     `json:"total_users"`
	TotalProducts int     `json:"total_products"`
	TotalOrders   int     `json:"total_orders"`
	TotalRevenue  float64 `json:"total_revenue"`
	PendingOrders int     `json:"pending_orders"`
	LowStockItems int     `json:"low_stock_items"`
}

// RevenueByDay represents daily revenue aggregation.
type RevenueByDay struct {
	Date       string  `json:"date"`
	Revenue    float64 `json:"revenue"`
	OrderCount int     `json:"order_count"`
}

// TopProduct represents a product ranked by units sold.
type TopProduct struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	Revenue     float64 `json:"revenue"`
}

// OrderSummary is a lightweight order view for admin listings.
type OrderSummary struct {
	ID         string    `json:"id"`
	UserEmail  string    `json:"user_email"`
	Status     string    `json:"status"`
	TotalPrice float64   `json:"total_price"`
	ItemCount  int       `json:"item_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// UserSummary is a lightweight user view for admin listings.
type UserSummary struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Role       string    `json:"role"`
	OrderCount int       `json:"order_count"`
	TotalSpent float64   `json:"total_spent"`
	CreatedAt  time.Time `json:"created_at"`
}
