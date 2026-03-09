package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kyle/product/internal/inventory/model"
	"github.com/kyle/product/internal/inventory/repository"
	"github.com/kyle/product/internal/platform/kafka"
	"github.com/kyle/product/pkg/event"
)

const lowStockThreshold = 10

type InventoryService struct {
	repo     repository.InventoryRepository
	producer *kafka.Producer
}

func NewInventoryService(repo repository.InventoryRepository, producer *kafka.Producer) *InventoryService {
	return &InventoryService{
		repo:     repo,
		producer: producer,
	}
}

func (s *InventoryService) GetStock(ctx context.Context, variantID, warehouse string) (*model.InventoryItem, error) {
	item, err := s.repo.GetByVariant(ctx, variantID, warehouse)
	if err != nil {
		return nil, fmt.Errorf("get stock: %w", err)
	}
	return item, nil
}

func (s *InventoryService) UpdateStock(ctx context.Context, req model.UpdateStockRequest) (*model.InventoryItem, error) {
	item := &model.InventoryItem{
		ID:               generateID(),
		ProductVariantID: req.ProductVariantID,
		Warehouse:        req.Warehouse,
		Quantity:         req.Quantity,
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.Upsert(ctx, item); err != nil {
		return nil, fmt.Errorf("upsert stock: %w", err)
	}

	updated, err := s.repo.GetByVariant(ctx, req.ProductVariantID, req.Warehouse)
	if err != nil {
		return nil, fmt.Errorf("get updated stock: %w", err)
	}

	available := updated.Quantity - updated.Reserved
	if available < lowStockThreshold {
		evt := event.InventoryLow{
			ProductVariantID: updated.ProductVariantID,
			CurrentStock:     available,
			Warehouse:        updated.Warehouse,
		}
		if err := s.producer.Publish(ctx, event.TopicInventoryLow, updated.ProductVariantID, evt); err != nil {
			return nil, fmt.Errorf("publish inventory.low-stock: %w", err)
		}
	}

	return updated, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, variantID string, qty int) error {
	if err := s.repo.Reserve(ctx, variantID, qty); err != nil {
		return fmt.Errorf("reserve stock: %w", err)
	}
	return nil
}

func (s *InventoryService) ReleaseStock(ctx context.Context, variantID string, qty int) error {
	if err := s.repo.Release(ctx, variantID, qty); err != nil {
		return fmt.Errorf("release stock: %w", err)
	}
	return nil
}

func (s *InventoryService) ListLowStock(ctx context.Context, threshold int) ([]model.InventoryItem, error) {
	items, err := s.repo.ListLowStock(ctx, threshold)
	if err != nil {
		return nil, fmt.Errorf("list low stock: %w", err)
	}
	return items, nil
}

func (s *InventoryService) HandleOrderCreated(ctx context.Context, evt event.OrderCreated) error {
	for _, item := range evt.Items {
		if err := s.repo.Reserve(ctx, item.ProductVariantID, item.Quantity); err != nil {
			return fmt.Errorf("reserve variant %s: %w", item.ProductVariantID, err)
		}
	}
	return nil
}

func (s *InventoryService) HandleOrderCancelled(ctx context.Context, evt event.OrderCancelled) error {
	for _, item := range evt.Items {
		if err := s.repo.Release(ctx, item.ProductVariantID, item.Quantity); err != nil {
			return fmt.Errorf("release variant %s: %w", item.ProductVariantID, err)
		}
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
