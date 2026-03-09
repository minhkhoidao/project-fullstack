package service

import (
	"context"
	"log/slog"

	"github.com/kyle/product/pkg/event"
)

// NotificationService dispatches notifications based on domain events.
// In production this would send emails/SMS; for now it performs structured logging.
type NotificationService struct {
	log *slog.Logger
}

// NewNotificationService creates a NotificationService.
func NewNotificationService(log *slog.Logger) *NotificationService {
	return &NotificationService{log: log}
}

// HandleOrderCreated sends an order confirmation to the customer.
func (s *NotificationService) HandleOrderCreated(ctx context.Context, evt event.OrderCreated) error {
	s.log.InfoContext(ctx, "sending order confirmation",
		"user_id", evt.UserID,
		"order_id", evt.OrderID,
		"total_price", evt.TotalPrice,
	)
	return nil
}

// HandleOrderPaid sends a payment receipt to the customer.
func (s *NotificationService) HandleOrderPaid(ctx context.Context, evt event.OrderPaid) error {
	s.log.InfoContext(ctx, "sending payment receipt",
		"order_id", evt.OrderID,
		"payment_id", evt.PaymentID,
		"amount", evt.Amount,
	)
	return nil
}

// HandleOrderCancelled sends a cancellation notice.
func (s *NotificationService) HandleOrderCancelled(ctx context.Context, evt event.OrderCancelled) error {
	s.log.InfoContext(ctx, "sending cancellation notice",
		"order_id", evt.OrderID,
		"reason", evt.Reason,
	)
	return nil
}

// HandlePaymentFailed sends a payment failure notice.
func (s *NotificationService) HandlePaymentFailed(ctx context.Context, evt event.PaymentFailed) error {
	s.log.InfoContext(ctx, "sending payment failure notice",
		"order_id", evt.OrderID,
		"payment_id", evt.PaymentID,
		"reason", evt.Reason,
	)
	return nil
}

// HandleLowStock sends a low stock alert for the given variant.
func (s *NotificationService) HandleLowStock(ctx context.Context, evt event.InventoryLow) error {
	s.log.InfoContext(ctx, "sending low stock alert",
		"product_variant_id", evt.ProductVariantID,
		"current_stock", evt.CurrentStock,
		"warehouse", evt.Warehouse,
	)
	return nil
}
