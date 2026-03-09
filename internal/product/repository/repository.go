package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyle/product/internal/product/model"
)

// ProductRepository defines persistence operations for the product domain.
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id string) (*model.Product, error)
	GetBySlug(ctx context.Context, slug string) (*model.Product, error)
	List(ctx context.Context, filter model.ProductFilter, cursor string, limit int) ([]model.Product, string, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id string) error
	CreateCategory(ctx context.Context, cat *model.Category) error
	ListCategories(ctx context.Context) ([]model.Category, error)
	CreateVariant(ctx context.Context, variant *model.ProductVariant) error
	ListVariants(ctx context.Context, productID string) ([]model.ProductVariant, error)
}

// Compile-time interface compliance check.
var _ ProductRepository = (*pgRepo)(nil)

type pgRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a ProductRepository backed by PostgreSQL.
func NewPostgresRepository(pool *pgxpool.Pool) ProductRepository {
	return &pgRepo{pool: pool}
}

// ---------------------------------------------------------------------------
// Products
// ---------------------------------------------------------------------------

func (r *pgRepo) Create(ctx context.Context, product *model.Product) error {
	product.Slug = slugify(product.Name)
	product.CreatedAt = time.Now().UTC()
	product.UpdatedAt = product.CreatedAt

	_, err := r.pool.Exec(ctx, `
		INSERT INTO products.products (id, category_id, name, slug, description, base_price, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		product.ID, product.CategoryID, product.Name, product.Slug,
		product.Description, product.BasePrice, product.IsActive,
		product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *pgRepo) GetByID(ctx context.Context, id string) (*model.Product, error) {
	return r.getProduct(ctx, "id", id)
}

func (r *pgRepo) GetBySlug(ctx context.Context, slug string) (*model.Product, error) {
	return r.getProduct(ctx, "slug", slug)
}

func (r *pgRepo) getProduct(ctx context.Context, column, value string) (*model.Product, error) {
	query := fmt.Sprintf(`
		SELECT id, category_id, name, slug, description, base_price, is_active, created_at, updated_at
		FROM products.products
		WHERE %s = $1`, column)

	var p model.Product
	err := r.pool.QueryRow(ctx, query, value).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
		&p.BasePrice, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("query product by %s: %w", column, err)
	}

	images, err := r.listImages(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Images = images

	variants, err := r.ListVariants(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Variants = variants

	return &p, nil
}

func (r *pgRepo) List(ctx context.Context, filter model.ProductFilter, cursor string, limit int) ([]model.Product, string, error) {
	args := []any{}
	clauses := []string{"1=1"}
	argIdx := 1

	if filter.CategoryID != "" {
		clauses = append(clauses, fmt.Sprintf("p.category_id = $%d", argIdx))
		args = append(args, filter.CategoryID)
		argIdx++
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("p.name ILIKE '%%' || $%d || '%%'", argIdx))
		args = append(args, filter.Search)
		argIdx++
	}
	if filter.MinPrice != nil {
		clauses = append(clauses, fmt.Sprintf("p.base_price >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}
	if filter.MaxPrice != nil {
		clauses = append(clauses, fmt.Sprintf("p.base_price <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}
	if filter.Size != "" {
		clauses = append(clauses, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM products.product_variants v WHERE v.product_id = p.id AND v.size = $%d)", argIdx))
		args = append(args, filter.Size)
		argIdx++
	}
	if filter.Color != "" {
		clauses = append(clauses, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM products.product_variants v WHERE v.product_id = p.id AND v.color = $%d)", argIdx))
		args = append(args, filter.Color)
		argIdx++
	}

	if cursor != "" {
		cursorTime, cursorID, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("decode cursor: %w", err)
		}
		clauses = append(clauses, fmt.Sprintf(
			"(p.created_at, p.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursorTime, cursorID)
		argIdx += 2
	}

	where := strings.Join(clauses, " AND ")
	query := fmt.Sprintf(`
		SELECT p.id, p.category_id, p.name, p.slug, p.description,
		       p.base_price, p.is_active, p.created_at, p.updated_at
		FROM products.products p
		WHERE %s AND p.is_active = true
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $%d`, where, argIdx)
	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(
			&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
			&p.BasePrice, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scan product row: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate product rows: %w", err)
	}

	var nextCursor string
	if len(products) > limit {
		products = products[:limit]
		last := products[limit-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}

	return products, nextCursor, nil
}

func (r *pgRepo) Update(ctx context.Context, product *model.Product) error {
	product.Slug = slugify(product.Name)
	product.UpdatedAt = time.Now().UTC()

	tag, err := r.pool.Exec(ctx, `
		UPDATE products.products
		SET category_id = $1, name = $2, slug = $3, description = $4,
		    base_price = $5, is_active = $6, updated_at = $7
		WHERE id = $8`,
		product.CategoryID, product.Name, product.Slug, product.Description,
		product.BasePrice, product.IsActive, product.UpdatedAt, product.ID,
	)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func (r *pgRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM products.products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Categories
// ---------------------------------------------------------------------------

func (r *pgRepo) CreateCategory(ctx context.Context, cat *model.Category) error {
	cat.Slug = slugify(cat.Name)
	cat.CreatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO products.categories (id, name, slug, description, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		cat.ID, cat.Name, cat.Slug, cat.Description, cat.ParentID, cat.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func (r *pgRepo) ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, description, parent_id, created_at
		FROM products.categories
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.ParentID, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category rows: %w", err)
	}
	return cats, nil
}

// ---------------------------------------------------------------------------
// Variants
// ---------------------------------------------------------------------------

func (r *pgRepo) CreateVariant(ctx context.Context, variant *model.ProductVariant) error {
	variant.CreatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO products.product_variants (id, product_id, sku, size, color, price_override, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		variant.ID, variant.ProductID, variant.SKU, variant.Size,
		variant.Color, variant.PriceOverride, variant.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert variant: %w", err)
	}
	return nil
}

func (r *pgRepo) ListVariants(ctx context.Context, productID string) ([]model.ProductVariant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, product_id, sku, size, color, price_override, created_at
		FROM products.product_variants
		WHERE product_id = $1
		ORDER BY created_at ASC`, productID)
	if err != nil {
		return nil, fmt.Errorf("list variants: %w", err)
	}
	defer rows.Close()

	var variants []model.ProductVariant
	for rows.Next() {
		var v model.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Size, &v.Color, &v.PriceOverride, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan variant row: %w", err)
		}
		variants = append(variants, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate variant rows: %w", err)
	}
	return variants, nil
}

// ---------------------------------------------------------------------------
// Images (internal helper)
// ---------------------------------------------------------------------------

func (r *pgRepo) listImages(ctx context.Context, productID string) ([]model.ProductImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, product_id, url, alt_text, sort_order
		FROM products.product_images
		WHERE product_id = $1
		ORDER BY sort_order ASC`, productID)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	defer rows.Close()

	var images []model.ProductImage
	for rows.Next() {
		var img model.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.URL, &img.AltText, &img.SortOrder); err != nil {
			return nil, fmt.Errorf("scan image row: %w", err)
		}
		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate image rows: %w", err)
	}
	return images, nil
}

// ---------------------------------------------------------------------------
// Cursor helpers (created_at + id keyset pagination)
// ---------------------------------------------------------------------------

type cursorPayload struct {
	CreatedAt time.Time `json:"c"`
	ID        string    `json:"i"`
}

func encodeCursor(createdAt time.Time, id string) string {
	payload := cursorPayload{CreatedAt: createdAt, ID: id}
	b, _ := json.Marshal(payload)
	return base64.URLEncoding.EncodeToString(b)
}

func decodeCursor(cursor string) (time.Time, string, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("base64 decode: %w", err)
	}
	var payload cursorPayload
	if err := json.Unmarshal(b, &payload); err != nil {
		return time.Time{}, "", fmt.Errorf("json unmarshal cursor: %w", err)
	}
	return payload.CreatedAt, payload.ID, nil
}

// slugify lowercases and replaces spaces with hyphens.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	return strings.ReplaceAll(s, " ", "-")
}
