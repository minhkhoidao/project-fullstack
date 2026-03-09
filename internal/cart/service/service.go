package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyle/product/internal/cart/model"
	"github.com/kyle/product/internal/cart/repository"
)

var ErrItemNotFound = errors.New("item not found in cart")

// CartService implements cart business logic.
type CartService struct {
	repo repository.CartRepository
}

// NewCartService creates a CartService with the given repository.
func NewCartService(repo repository.CartRepository) *CartService {
	return &CartService{repo: repo}
}

// GetCart returns the user's current cart.
func (s *CartService) GetCart(ctx context.Context, userID string) (*model.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}
	return cart, nil
}

// AddItem adds a new item to the cart or increments its quantity if already present.
func (s *CartService) AddItem(ctx context.Context, userID string, req model.AddItemRequest) (*model.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	for i, item := range cart.Items {
		if item.ProductVariantID == req.ProductVariantID {
			cart.Items[i].Quantity += req.Quantity
			if err := s.repo.Save(ctx, cart); err != nil {
				return nil, fmt.Errorf("save cart: %w", err)
			}
			return cart, nil
		}
	}

	cart.Items = append(cart.Items, model.CartItem{
		ProductVariantID: req.ProductVariantID,
		ProductName:      req.ProductName,
		VariantInfo:      req.VariantInfo,
		Quantity:         req.Quantity,
		UnitPrice:        req.UnitPrice,
	})

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, fmt.Errorf("save cart: %w", err)
	}

	return cart, nil
}

// UpdateItem updates the quantity of an existing cart item.
// If the new quantity is 0, the item is removed.
func (s *CartService) UpdateItem(ctx context.Context, userID, variantID string, req model.UpdateItemRequest) (*model.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	idx := -1
	for i, item := range cart.Items {
		if item.ProductVariantID == variantID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return nil, ErrItemNotFound
	}

	if req.Quantity <= 0 {
		cart.Items = append(cart.Items[:idx], cart.Items[idx+1:]...)
	} else {
		cart.Items[idx].Quantity = req.Quantity
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, fmt.Errorf("save cart: %w", err)
	}

	return cart, nil
}

// RemoveItem removes an item from the cart by variant ID.
func (s *CartService) RemoveItem(ctx context.Context, userID, variantID string) (*model.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	filtered := make([]model.CartItem, 0, len(cart.Items))
	found := false
	for _, item := range cart.Items {
		if item.ProductVariantID == variantID {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}

	if !found {
		return nil, ErrItemNotFound
	}

	cart.Items = filtered

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, fmt.Errorf("save cart: %w", err)
	}

	return cart, nil
}

// ClearCart removes all items from the user's cart.
func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("clear cart: %w", err)
	}
	return nil
}
