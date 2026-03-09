package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyle/product/internal/order/model"
)

var ErrNotFound = errors.New("order not found")

// OrderRepository defines the persistence interface for orders.
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	ListByUser(ctx context.Context, userID, cursor string, limit int) ([]model.Order, string, error)
	UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus) error
}

type pgRepo struct {
	pool *pgxpool.Pool
}

// NewPGRepository returns an OrderRepository backed by PostgreSQL.
func NewPGRepository(pool *pgxpool.Pool) OrderRepository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO orders.orders (id, user_id, status, total_price, street, city, state, zip, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.ID,
		order.UserID,
		order.Status,
		order.TotalPrice,
		order.ShippingAddress.Street,
		order.ShippingAddress.City,
		order.ShippingAddress.State,
		order.ShippingAddress.Zip,
		order.ShippingAddress.Country,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO orders.order_items (id, order_id, product_variant_id, product_name, variant_info, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			item.ID,
			order.ID,
			item.ProductVariantID,
			item.ProductName,
			item.VariantInfo,
			item.Quantity,
			item.UnitPrice,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *pgRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var o model.Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, total_price, street, city, state, zip, country, created_at, updated_at
		FROM orders.orders
		WHERE id = $1`, id,
	).Scan(
		&o.ID, &o.UserID, &o.Status, &o.TotalPrice,
		&o.ShippingAddress.Street, &o.ShippingAddress.City,
		&o.ShippingAddress.State, &o.ShippingAddress.Zip,
		&o.ShippingAddress.Country,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("query order %s: %w", id, err)
	}

	items, err := r.queryItems(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *pgRepo) ListByUser(ctx context.Context, userID, cursor string, limit int) ([]model.Order, string, error) {
	if limit <= 0 {
		limit = 20
	}

	var (
		rows pgx.Rows
		err  error
	)

	if cursor != "" {
		cursorTime, decErr := decodeCursor(cursor)
		if decErr != nil {
			return nil, "", fmt.Errorf("decode cursor: %w", decErr)
		}
		rows, err = r.pool.Query(ctx, `
			SELECT id, user_id, status, total_price, street, city, state, zip, country, created_at, updated_at
			FROM orders.orders
			WHERE user_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3`, userID, cursorTime, limit+1,
		)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, user_id, status, total_price, street, city, state, zip, country, created_at, updated_at
			FROM orders.orders
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2`, userID, limit+1,
		)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0, limit)
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.TotalPrice,
			&o.ShippingAddress.Street, &o.ShippingAddress.City,
			&o.ShippingAddress.State, &o.ShippingAddress.Zip,
			&o.ShippingAddress.Country,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate orders: %w", err)
	}

	var nextCursor string
	if len(orders) > limit {
		orders = orders[:limit]
		nextCursor = encodeCursor(orders[len(orders)-1].CreatedAt)
	}

	for i := range orders {
		items, err := r.queryItems(ctx, orders[i].ID)
		if err != nil {
			return nil, "", err
		}
		orders[i].Items = items
	}

	return orders, nextCursor, nil
}

func (r *pgRepo) UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE orders.orders SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), orderID,
	)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("order %s: %w", orderID, ErrNotFound)
	}
	return nil
}

func (r *pgRepo) queryItems(ctx context.Context, orderID string) ([]model.OrderItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, product_variant_id, product_name, variant_info, quantity, unit_price
		FROM orders.order_items
		WHERE order_id = $1`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductVariantID,
			&item.ProductName, &item.VariantInfo,
			&item.Quantity, &item.UnitPrice,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func encodeCursor(t time.Time) string {
	return base64.StdEncoding.EncodeToString([]byte(t.Format(time.RFC3339Nano)))
}

func decodeCursor(cursor string) (time.Time, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, string(b))
}
