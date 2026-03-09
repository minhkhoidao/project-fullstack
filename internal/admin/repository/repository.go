package repository

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyle/product/internal/admin/model"
)

// AdminRepository defines the persistence interface for admin operations.
type AdminRepository interface {
	GetDashboardStats(ctx context.Context) (*model.DashboardStats, error)
	GetRevenueByDay(ctx context.Context, days int) ([]model.RevenueByDay, error)
	GetTopProducts(ctx context.Context, limit int) ([]model.TopProduct, error)
	ListOrders(ctx context.Context, status, cursor string, limit int) ([]model.OrderSummary, string, error)
	ListUsers(ctx context.Context, cursor string, limit int) ([]model.UserSummary, string, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	UpdateUserRole(ctx context.Context, userID, role string) error
}

type pgRepo struct {
	pool *pgxpool.Pool
}

// NewPGRepository returns an AdminRepository backed by PostgreSQL.
func NewPGRepository(pool *pgxpool.Pool) AdminRepository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	var s model.DashboardStats

	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users.users),
			(SELECT COUNT(*) FROM products.products),
			(SELECT COUNT(*) FROM orders.orders),
			(SELECT COALESCE(SUM(total_price), 0) FROM orders.orders WHERE status != 'cancelled'),
			(SELECT COUNT(*) FROM orders.orders WHERE status = 'pending'),
			(SELECT COUNT(*) FROM inventory.inventory WHERE quantity - reserved < 10)
	`).Scan(
		&s.TotalUsers,
		&s.TotalProducts,
		&s.TotalOrders,
		&s.TotalRevenue,
		&s.PendingOrders,
		&s.LowStockItems,
	)
	if err != nil {
		return nil, fmt.Errorf("query dashboard stats: %w", err)
	}

	return &s, nil
}

func (r *pgRepo) GetRevenueByDay(ctx context.Context, days int) ([]model.RevenueByDay, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			DATE(created_at)::text AS date,
			SUM(total_price)       AS revenue,
			COUNT(*)               AS order_count
		FROM orders.orders
		WHERE status != 'cancelled'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT $1`, days,
	)
	if err != nil {
		return nil, fmt.Errorf("query revenue by day: %w", err)
	}
	defer rows.Close()

	var results []model.RevenueByDay
	for rows.Next() {
		var r model.RevenueByDay
		if err := rows.Scan(&r.Date, &r.Revenue, &r.OrderCount); err != nil {
			return nil, fmt.Errorf("scan revenue row: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate revenue rows: %w", err)
	}

	return results, nil
}

func (r *pgRepo) GetTopProducts(ctx context.Context, limit int) ([]model.TopProduct, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			p.id                        AS product_id,
			p.name                      AS product_name,
			SUM(oi.quantity)::int       AS total_sold,
			SUM(oi.quantity * oi.unit_price) AS revenue
		FROM orders.order_items oi
		JOIN products.products p ON p.id = oi.product_variant_id
		GROUP BY p.id, p.name
		ORDER BY total_sold DESC
		LIMIT $1`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query top products: %w", err)
	}
	defer rows.Close()

	var results []model.TopProduct
	for rows.Next() {
		var tp model.TopProduct
		if err := rows.Scan(&tp.ProductID, &tp.ProductName, &tp.TotalSold, &tp.Revenue); err != nil {
			return nil, fmt.Errorf("scan top product: %w", err)
		}
		results = append(results, tp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top products: %w", err)
	}

	return results, nil
}

func (r *pgRepo) ListOrders(ctx context.Context, status, cursor string, limit int) ([]model.OrderSummary, string, error) {
	if limit <= 0 {
		limit = 20
	}

	var (
		rows pgx.Rows
		err  error
	)

	query := `
		SELECT
			o.id,
			u.email,
			o.status,
			o.total_price,
			(SELECT COUNT(*) FROM orders.order_items oi WHERE oi.order_id = o.id)::int AS item_count,
			o.created_at
		FROM orders.orders o
		JOIN users.users u ON u.id = o.user_id`

	args := make([]any, 0, 3)
	argIdx := 1
	where := ""

	if status != "" {
		where += fmt.Sprintf(" WHERE o.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if cursor != "" {
		cursorTime, decErr := decodeCursor(cursor)
		if decErr != nil {
			return nil, "", fmt.Errorf("decode cursor: %w", decErr)
		}
		if where == "" {
			where += " WHERE"
		} else {
			where += " AND"
		}
		where += fmt.Sprintf(" o.created_at < $%d", argIdx)
		args = append(args, cursorTime)
		argIdx++
	}

	query += where + fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit+1)

	rows, err = r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.OrderSummary, 0, limit)
	for rows.Next() {
		var o model.OrderSummary
		if err := rows.Scan(&o.ID, &o.UserEmail, &o.Status, &o.TotalPrice, &o.ItemCount, &o.CreatedAt); err != nil {
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

	return orders, nextCursor, nil
}

func (r *pgRepo) ListUsers(ctx context.Context, cursor string, limit int) ([]model.UserSummary, string, error) {
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
			SELECT
				u.id,
				u.email,
				u.first_name,
				u.last_name,
				u.role,
				COUNT(o.id)::int              AS order_count,
				COALESCE(SUM(o.total_price), 0) AS total_spent,
				u.created_at
			FROM users.users u
			LEFT JOIN orders.orders o ON o.user_id = u.id
			WHERE u.created_at < $1
			GROUP BY u.id
			ORDER BY u.created_at DESC
			LIMIT $2`, cursorTime, limit+1,
		)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT
				u.id,
				u.email,
				u.first_name,
				u.last_name,
				u.role,
				COUNT(o.id)::int              AS order_count,
				COALESCE(SUM(o.total_price), 0) AS total_spent,
				u.created_at
			FROM users.users u
			LEFT JOIN orders.orders o ON o.user_id = u.id
			GROUP BY u.id
			ORDER BY u.created_at DESC
			LIMIT $1`, limit+1,
		)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]model.UserSummary, 0, limit)
	for rows.Next() {
		var u model.UserSummary
		if err := rows.Scan(
			&u.ID, &u.Email, &u.FirstName, &u.LastName,
			&u.Role, &u.OrderCount, &u.TotalSpent, &u.CreatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate users: %w", err)
	}

	var nextCursor string
	if len(users) > limit {
		users = users[:limit]
		nextCursor = encodeCursor(users[len(users)-1].CreatedAt)
	}

	return users, nextCursor, nil
}

func (r *pgRepo) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE orders.orders SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, orderID,
	)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("order %s not found", orderID)
	}
	return nil
}

func (r *pgRepo) UpdateUserRole(ctx context.Context, userID, role string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE users.users SET role = $1, updated_at = NOW() WHERE id = $2`,
		role, userID,
	)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user %s not found", userID)
	}
	return nil
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
