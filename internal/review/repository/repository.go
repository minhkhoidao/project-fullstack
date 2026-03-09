package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyle/product/internal/review/model"
)

var ErrNotFound = errors.New("review not found")

// ReviewRepository defines the persistence interface for reviews.
type ReviewRepository interface {
	Create(ctx context.Context, review *model.Review) error
	GetByID(ctx context.Context, id string) (*model.Review, error)
	ListByProduct(ctx context.Context, productID, cursor string, limit int) ([]model.Review, string, error)
	Update(ctx context.Context, review *model.Review) error
	Delete(ctx context.Context, id string) error
	GetSummary(ctx context.Context, productID string) (*model.ReviewSummary, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

var _ ReviewRepository = (*pgRepo)(nil)

// NewPGRepository returns a ReviewRepository backed by PostgreSQL.
func NewPGRepository(pool *pgxpool.Pool) ReviewRepository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) Create(ctx context.Context, review *model.Review) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO reviews.reviews (id, product_id, user_id, rating, title, body, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		review.ID,
		review.ProductID,
		review.UserID,
		review.Rating,
		review.Title,
		review.Body,
		review.CreatedAt,
		review.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert review: %w", err)
	}

	for _, img := range review.Images {
		_, err = tx.Exec(ctx, `
			INSERT INTO reviews.review_images (id, review_id, url, sort_order)
			VALUES ($1, $2, $3, $4)`,
			img.ID, review.ID, img.URL, img.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("insert review image: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *pgRepo) GetByID(ctx context.Context, id string) (*model.Review, error) {
	var rv model.Review
	err := r.pool.QueryRow(ctx, `
		SELECT id, product_id, user_id, rating, title, body, created_at, updated_at
		FROM reviews.reviews
		WHERE id = $1`, id,
	).Scan(
		&rv.ID, &rv.ProductID, &rv.UserID, &rv.Rating,
		&rv.Title, &rv.Body, &rv.CreatedAt, &rv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("review %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("query review %s: %w", id, err)
	}

	images, err := r.queryImages(ctx, id)
	if err != nil {
		return nil, err
	}
	rv.Images = images

	return &rv, nil
}

func (r *pgRepo) ListByProduct(ctx context.Context, productID, cursor string, limit int) ([]model.Review, string, error) {
	if limit <= 0 {
		limit = 20
	}

	var (
		rows pgx.Rows
		err  error
	)

	if cursor != "" {
		cursorTime, cursorID, decErr := decodeCursor(cursor)
		if decErr != nil {
			return nil, "", fmt.Errorf("decode cursor: %w", decErr)
		}
		rows, err = r.pool.Query(ctx, `
			SELECT id, product_id, user_id, rating, title, body, created_at, updated_at
			FROM reviews.reviews
			WHERE product_id = $1
			  AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4`, productID, cursorTime, cursorID, limit+1,
		)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, product_id, user_id, rating, title, body, created_at, updated_at
			FROM reviews.reviews
			WHERE product_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2`, productID, limit+1,
		)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]model.Review, 0, limit)
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(
			&rv.ID, &rv.ProductID, &rv.UserID, &rv.Rating,
			&rv.Title, &rv.Body, &rv.CreatedAt, &rv.UpdatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate reviews: %w", err)
	}

	var nextCursor string
	if len(reviews) > limit {
		reviews = reviews[:limit]
		last := reviews[len(reviews)-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}

	for i := range reviews {
		images, err := r.queryImages(ctx, reviews[i].ID)
		if err != nil {
			return nil, "", err
		}
		reviews[i].Images = images
	}

	return reviews, nextCursor, nil
}

func (r *pgRepo) Update(ctx context.Context, review *model.Review) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE reviews.reviews
		SET rating = $1, title = $2, body = $3, updated_at = $4
		WHERE id = $5`,
		review.Rating, review.Title, review.Body, review.UpdatedAt, review.ID,
	)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("review %s: %w", review.ID, ErrNotFound)
	}
	return nil
}

func (r *pgRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM reviews.reviews WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("review %s: %w", id, ErrNotFound)
	}
	return nil
}

func (r *pgRepo) GetSummary(ctx context.Context, productID string) (*model.ReviewSummary, error) {
	var s model.ReviewSummary
	err := r.pool.QueryRow(ctx, `
		SELECT product_id, AVG(rating), COUNT(*)
		FROM reviews.reviews
		WHERE product_id = $1
		GROUP BY product_id`, productID,
	).Scan(&s.ProductID, &s.AverageRating, &s.TotalReviews)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.ReviewSummary{ProductID: productID}, nil
		}
		return nil, fmt.Errorf("query review summary: %w", err)
	}
	return &s, nil
}

func (r *pgRepo) queryImages(ctx context.Context, reviewID string) ([]model.ReviewImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, review_id, url, sort_order
		FROM reviews.review_images
		WHERE review_id = $1
		ORDER BY sort_order`, reviewID,
	)
	if err != nil {
		return nil, fmt.Errorf("query review images: %w", err)
	}
	defer rows.Close()

	var images []model.ReviewImage
	for rows.Next() {
		var img model.ReviewImage
		if err := rows.Scan(&img.ID, &img.ReviewID, &img.URL, &img.SortOrder); err != nil {
			return nil, fmt.Errorf("scan review image: %w", err)
		}
		images = append(images, img)
	}

	return images, rows.Err()
}

// Cursor encodes (created_at, id) for deterministic keyset pagination.
func encodeCursor(t time.Time, id string) string {
	raw := t.Format(time.RFC3339Nano) + "|" + id
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (time.Time, string, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}

	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("malformed cursor")
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}

	return t, parts[1], nil
}
