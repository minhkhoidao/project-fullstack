package model

import "time"

// Notification represents a dispatched notification record.
type Notification struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Type      string     `json:"type"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body"`
	Channel   string     `json:"channel"`
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time  `json:"created_at"`
}

const (
	TypeOrderConfirmation = "order_confirmation"
	TypePaymentReceived   = "payment_received"
	TypeOrderShipped      = "order_shipped"
	TypeOrderCancelled    = "order_cancelled"
	TypeLowStock          = "low_stock"
)
