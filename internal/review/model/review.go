package model

import "time"

// Review represents a user-submitted product review.
type Review struct {
	ID        string        `json:"id"`
	ProductID string        `json:"product_id"`
	UserID    string        `json:"user_id"`
	Rating    int           `json:"rating"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	Images    []ReviewImage `json:"images"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// ReviewImage is a single image attached to a review.
type ReviewImage struct {
	ID        string `json:"id"`
	ReviewID  string `json:"review_id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

// CreateReviewRequest is the payload for submitting a new review.
type CreateReviewRequest struct {
	ProductID string `json:"product_id"`
	Rating    int    `json:"rating"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

// UpdateReviewRequest supports partial updates via pointer fields.
type UpdateReviewRequest struct {
	Rating *int    `json:"rating"`
	Title  *string `json:"title"`
	Body   *string `json:"body"`
}

// ReviewSummary holds aggregate rating statistics for a product.
type ReviewSummary struct {
	ProductID    string  `json:"product_id"`
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
}
