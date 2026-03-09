package model

import "time"

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID         string        `json:"id"`
	OrderID    string        `json:"order_id"`
	Amount     float64       `json:"amount"`
	Method     string        `json:"method"`
	Status     PaymentStatus `json:"status"`
	ExternalID string        `json:"external_id"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type ProcessPaymentRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Method  string  `json:"method"`
}

// WebhookPayload simulates an external payment provider callback.
type WebhookPayload struct {
	PaymentID   string `json:"payment_id"`
	Status      string `json:"status"`
	ExternalRef string `json:"external_ref"`
}
