package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kyle/product/internal/cart/model"
)

const cartTTL = 7 * 24 * time.Hour // 7 days

// CartRepository defines the persistence interface for shopping carts.
type CartRepository interface {
	Get(ctx context.Context, userID string) (*model.Cart, error)
	Save(ctx context.Context, cart *model.Cart) error
	Delete(ctx context.Context, userID string) error
}

type redisRepo struct {
	client *redis.Client
}

// NewRedisRepository returns a CartRepository backed by Redis.
func NewRedisRepository(client *redis.Client) CartRepository {
	return &redisRepo{client: client}
}

func cartKey(userID string) string {
	return "cart:" + userID
}

func (r *redisRepo) Get(ctx context.Context, userID string) (*model.Cart, error) {
	data, err := r.client.Get(ctx, cartKey(userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &model.Cart{
				UserID:    userID,
				Items:     []model.CartItem{},
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("get cart %s: %w", userID, err)
	}

	var cart model.Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, fmt.Errorf("unmarshal cart %s: %w", userID, err)
	}

	return &cart, nil
}

func (r *redisRepo) Save(ctx context.Context, cart *model.Cart) error {
	cart.UpdatedAt = time.Now()

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("marshal cart %s: %w", cart.UserID, err)
	}

	if err := r.client.Set(ctx, cartKey(cart.UserID), data, cartTTL).Err(); err != nil {
		return fmt.Errorf("save cart %s: %w", cart.UserID, err)
	}

	return nil
}

func (r *redisRepo) Delete(ctx context.Context, userID string) error {
	if err := r.client.Del(ctx, cartKey(userID)).Err(); err != nil {
		return fmt.Errorf("delete cart %s: %w", userID, err)
	}
	return nil
}
