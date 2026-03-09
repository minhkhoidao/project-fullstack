package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyle/product/internal/inventory/model"
)

var (
	ErrNotFound          = errors.New("inventory item not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type InventoryRepository interface {
	GetByVariant(ctx context.Context, variantID, warehouse string) (*model.InventoryItem, error)
	Upsert(ctx context.Context, item *model.InventoryItem) error
	Reserve(ctx context.Context, variantID string, qty int) error
	Release(ctx context.Context, variantID string, qty int) error
	ListLowStock(ctx context.Context, threshold int) ([]model.InventoryItem, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

var _ InventoryRepository = (*pgRepo)(nil)

func NewPGRepository(pool *pgxpool.Pool) InventoryRepository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) GetByVariant(ctx context.Context, variantID, warehouse string) (*model.InventoryItem, error) {
	var item model.InventoryItem
	err := r.pool.QueryRow(ctx, `
		SELECT id, product_variant_id, warehouse, quantity, reserved, updated_at
		FROM inventory.inventory
		WHERE product_variant_id = $1 AND warehouse = $2`, variantID, warehouse,
	).Scan(
		&item.ID, &item.ProductVariantID, &item.Warehouse,
		&item.Quantity, &item.Reserved, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("variant %s warehouse %s: %w", variantID, warehouse, ErrNotFound)
		}
		return nil, fmt.Errorf("query inventory: %w", err)
	}
	return &item, nil
}

func (r *pgRepo) Upsert(ctx context.Context, item *model.InventoryItem) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO inventory.inventory (id, product_variant_id, warehouse, quantity, reserved, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (product_variant_id, warehouse)
		DO UPDATE SET quantity = $4, updated_at = $6`,
		item.ID, item.ProductVariantID, item.Warehouse,
		item.Quantity, item.Reserved, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert inventory: %w", err)
	}
	return nil
}

func (r *pgRepo) Reserve(ctx context.Context, variantID string, qty int) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE inventory.inventory
		SET reserved = reserved + $1, updated_at = $3
		WHERE product_variant_id = $2 AND quantity - reserved >= $1`,
		qty, variantID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("reserve stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("variant %s qty %d: %w", variantID, qty, ErrInsufficientStock)
	}
	return nil
}

func (r *pgRepo) Release(ctx context.Context, variantID string, qty int) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE inventory.inventory
		SET reserved = reserved - $1, updated_at = $3
		WHERE product_variant_id = $2 AND reserved >= $1`,
		qty, variantID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("release stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("variant %s qty %d: %w", variantID, qty, ErrInsufficientStock)
	}
	return nil
}

func (r *pgRepo) ListLowStock(ctx context.Context, threshold int) ([]model.InventoryItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, product_variant_id, warehouse, quantity, reserved, updated_at
		FROM inventory.inventory
		WHERE quantity - reserved < $1
		ORDER BY quantity - reserved ASC`, threshold,
	)
	if err != nil {
		return nil, fmt.Errorf("list low stock: %w", err)
	}
	defer rows.Close()

	var items []model.InventoryItem
	for rows.Next() {
		var item model.InventoryItem
		if err := rows.Scan(
			&item.ID, &item.ProductVariantID, &item.Warehouse,
			&item.Quantity, &item.Reserved, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory: %w", err)
	}

	return items, nil
}
