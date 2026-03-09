package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	cartrepo "github.com/kyle/product/internal/cart/repository"
	"github.com/kyle/product/internal/order/model"
	"github.com/kyle/product/internal/order/repository"
	"github.com/kyle/product/internal/platform/kafka"
	"github.com/kyle/product/pkg/event"
)

var (
	ErrEmptyCart     = errors.New("cart is empty")
	ErrNotOwner      = errors.New("order does not belong to user")
	ErrNotCancelable = errors.New("only pending orders can be cancelled")
)

// OrderService implements order business logic.
type OrderService struct {
	orderRepo repository.OrderRepository
	cartRepo  cartrepo.CartRepository
	producer  *kafka.Producer
}

// NewOrderService creates an OrderService with the given dependencies.
func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo cartrepo.CartRepository,
	producer *kafka.Producer,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
		producer:  producer,
	}
}

// CreateOrder builds an order from the user's cart, persists it,
// publishes an order.created event, and clears the cart.
func (s *OrderService) CreateOrder(ctx context.Context, userID string, req model.CreateOrderRequest) (*model.Order, error) {
	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}

	now := time.Now()
	orderID := generateID()

	items := make([]model.OrderItem, 0, len(cart.Items))
	var total float64
	for _, ci := range cart.Items {
		items = append(items, model.OrderItem{
			ID:               generateID(),
			OrderID:          orderID,
			ProductVariantID: ci.ProductVariantID,
			ProductName:      ci.ProductName,
			VariantInfo:      ci.VariantInfo,
			Quantity:         ci.Quantity,
			UnitPrice:        ci.UnitPrice,
		})
		total += ci.UnitPrice * float64(ci.Quantity)
	}

	order := &model.Order{
		ID:              orderID,
		UserID:          userID,
		Status:          model.StatusPending,
		TotalPrice:      total,
		ShippingAddress: req.ShippingAddress,
		Items:           items,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	eventItems := make([]event.OrderItem, 0, len(items))
	for _, item := range items {
		eventItems = append(eventItems, event.OrderItem{
			ProductVariantID: item.ProductVariantID,
			Quantity:         item.Quantity,
			Price:            item.UnitPrice,
		})
	}

	evt := event.OrderCreated{
		OrderID:    order.ID,
		UserID:     userID,
		Items:      eventItems,
		TotalPrice: total,
		CreatedAt:  now,
	}

	if err := s.producer.Publish(ctx, event.TopicOrderCreated, order.ID, evt); err != nil {
		return nil, fmt.Errorf("publish order.created: %w", err)
	}

	if err := s.cartRepo.Delete(ctx, userID); err != nil {
		return nil, fmt.Errorf("clear cart: %w", err)
	}

	return order, nil
}

// GetOrder returns an order by ID, verifying the caller owns it.
func (s *OrderService) GetOrder(ctx context.Context, userID, orderID string) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	if order.UserID != userID {
		return nil, ErrNotOwner
	}

	return order, nil
}

// ListOrders returns a paginated list of orders for the given user.
func (s *OrderService) ListOrders(ctx context.Context, userID, cursor string, limit int) ([]model.Order, string, error) {
	orders, nextCursor, err := s.orderRepo.ListByUser(ctx, userID, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("list orders: %w", err)
	}
	return orders, nextCursor, nil
}

// CancelOrder cancels a pending order and publishes an order.cancelled event.
func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	if order.UserID != userID {
		return ErrNotOwner
	}

	if order.Status != model.StatusPending {
		return ErrNotCancelable
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, model.StatusCancelled); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	eventItems := make([]event.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		eventItems = append(eventItems, event.OrderItem{
			ProductVariantID: item.ProductVariantID,
			Quantity:         item.Quantity,
			Price:            item.UnitPrice,
		})
	}

	evt := event.OrderCancelled{
		OrderID:     orderID,
		Items:       eventItems,
		Reason:      "cancelled by user",
		CancelledAt: time.Now(),
	}

	if err := s.producer.Publish(ctx, event.TopicOrderCancelled, orderID, evt); err != nil {
		return fmt.Errorf("publish order.cancelled: %w", err)
	}

	return nil
}

// UpdateOrderStatus updates an order's status (for internal use by payment service, etc.).
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("generate id: %v", err))
	}
	return hex.EncodeToString(b)
}
