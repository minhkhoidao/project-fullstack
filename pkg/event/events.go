package event

import "time"

const (
	TopicOrderCreated   = "order.created"
	TopicOrderPaid      = "order.paid"
	TopicOrderCancelled = "order.cancelled"
	TopicInventoryLow   = "inventory.low-stock"
	TopicPaymentDone    = "payment.completed"
	TopicPaymentFailed  = "payment.failed"
)

type OrderCreated struct {
	OrderID    string      `json:"order_id"`
	UserID     string      `json:"user_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	CreatedAt  time.Time   `json:"created_at"`
}

type OrderItem struct {
	ProductVariantID string  `json:"product_variant_id"`
	Quantity         int     `json:"quantity"`
	Price            float64 `json:"price"`
}

type OrderPaid struct {
	OrderID   string    `json:"order_id"`
	PaymentID string    `json:"payment_id"`
	Amount    float64   `json:"amount"`
	PaidAt    time.Time `json:"paid_at"`
}

type OrderCancelled struct {
	OrderID     string      `json:"order_id"`
	Items       []OrderItem `json:"items"`
	Reason      string      `json:"reason"`
	CancelledAt time.Time   `json:"cancelled_at"`
}

type InventoryLow struct {
	ProductVariantID string `json:"product_variant_id"`
	CurrentStock     int    `json:"current_stock"`
	Warehouse        string `json:"warehouse"`
}

type PaymentCompleted struct {
	OrderID   string    `json:"order_id"`
	PaymentID string    `json:"payment_id"`
	Amount    float64   `json:"amount"`
	Method    string    `json:"method"`
	PaidAt    time.Time `json:"paid_at"`
}

type PaymentFailed struct {
	OrderID   string    `json:"order_id"`
	PaymentID string    `json:"payment_id"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}
