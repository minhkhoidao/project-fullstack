package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/kyle/product/internal/review/model"
	"github.com/kyle/product/internal/review/repository"
)

var (
	ErrNotOwner = errors.New("review does not belong to user")
)

// ReviewService implements review business logic.
type ReviewService struct {
	repo repository.ReviewRepository
}

// NewReviewService creates a ReviewService with the given repository.
func NewReviewService(repo repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

// CreateReview persists a new review for the given user.
func (s *ReviewService) CreateReview(ctx context.Context, userID string, req model.CreateReviewRequest) (*model.Review, error) {
	now := time.Now()
	review := &model.Review{
		ID:        generateID(),
		ProductID: req.ProductID,
		UserID:    userID,
		Rating:    req.Rating,
		Title:     req.Title,
		Body:      req.Body,
		Images:    []model.ReviewImage{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	return review, nil
}

// GetReview returns a single review by ID.
func (s *ReviewService) GetReview(ctx context.Context, id string) (*model.Review, error) {
	review, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	return review, nil
}

// ListProductReviews returns a paginated list of reviews for a product.
func (s *ReviewService) ListProductReviews(ctx context.Context, productID, cursor string, limit int) ([]model.Review, string, error) {
	reviews, nextCursor, err := s.repo.ListByProduct(ctx, productID, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("list product reviews: %w", err)
	}
	return reviews, nextCursor, nil
}

// UpdateReview applies a partial update to a review after verifying ownership.
func (s *ReviewService) UpdateReview(ctx context.Context, userID, reviewID string, req model.UpdateReviewRequest) (*model.Review, error) {
	review, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}

	if review.UserID != userID {
		return nil, ErrNotOwner
	}

	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	if req.Title != nil {
		review.Title = *req.Title
	}
	if req.Body != nil {
		review.Body = *req.Body
	}
	review.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, review); err != nil {
		return nil, fmt.Errorf("update review: %w", err)
	}

	return review, nil
}

// DeleteReview removes a review after verifying ownership.
func (s *ReviewService) DeleteReview(ctx context.Context, userID, reviewID string) error {
	review, err := s.repo.GetByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("get review: %w", err)
	}

	if review.UserID != userID {
		return ErrNotOwner
	}

	if err := s.repo.Delete(ctx, reviewID); err != nil {
		return fmt.Errorf("delete review: %w", err)
	}

	return nil
}

// GetProductReviewSummary returns aggregate rating statistics for a product.
func (s *ReviewService) GetProductReviewSummary(ctx context.Context, productID string) (*model.ReviewSummary, error) {
	summary, err := s.repo.GetSummary(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get review summary: %w", err)
	}
	return summary, nil
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("generate id: %v", err))
	}
	return hex.EncodeToString(b)
}
