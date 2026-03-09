package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyle/product/internal/payment/model"
)

var ErrNotFound = errors.New("payment not found")

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id string) (*model.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Payment, error)
	UpdateStatus(ctx context.Context, id string, status model.PaymentStatus) error
}

type pgRepo struct {
	pool *pgxpool.Pool
}

var _ PaymentRepository = (*pgRepo)(nil)

func NewPGRepository(pool *pgxpool.Pool) PaymentRepository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) Create(ctx context.Context, payment *model.Payment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO orders.payments (id, order_id, amount, method, status, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Method,
		payment.Status,
		payment.ExternalID,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

func (r *pgRepo) GetByID(ctx context.Context, id string) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, amount, method, status, external_id, created_at, updated_at
		FROM orders.payments
		WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.OrderID, &p.Amount, &p.Method,
		&p.Status, &p.ExternalID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("query payment %s: %w", id, err)
	}
	return &p, nil
}

func (r *pgRepo) GetByOrderID(ctx context.Context, orderID string) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, amount, method, status, external_id, created_at, updated_at
		FROM orders.payments
		WHERE order_id = $1`, orderID,
	).Scan(
		&p.ID, &p.OrderID, &p.Amount, &p.Method,
		&p.Status, &p.ExternalID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment for order %s: %w", orderID, ErrNotFound)
		}
		return nil, fmt.Errorf("query payment for order %s: %w", orderID, err)
	}
	return &p, nil
}

func (r *pgRepo) UpdateStatus(ctx context.Context, id string, status model.PaymentStatus) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE orders.payments SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("payment %s: %w", id, ErrNotFound)
	}
	return nil
}
