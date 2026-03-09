package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/review/model"
	"github.com/kyle/product/internal/review/repository"
	"github.com/kyle/product/internal/review/service"
	"github.com/kyle/product/pkg/dto"
	"github.com/kyle/product/pkg/httputil"
)

// ReviewHandler handles HTTP requests for the review API.
type ReviewHandler struct {
	svc *service.ReviewService
}

// NewReviewHandler creates a new ReviewHandler.
func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// RegisterRoutes mounts review routes on the given router.
func (h *ReviewHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1/reviews", func(r chi.Router) {
		r.Get("/product/{productID}", h.listProductReviews)
		r.Get("/product/{productID}/summary", h.getProductSummary)

		r.Group(func(r chi.Router) {
			r.Use(authMW)
			r.Post("/", h.createReview)
			r.Put("/{id}", h.updateReview)
			r.Delete("/{id}", h.deleteReview)
		})
	})
}

func (h *ReviewHandler) listProductReviews(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productID")
	p := dto.ParsePagination(r)

	reviews, nextCursor, err := h.svc.ListProductReviews(r.Context(), productID, p.Cursor, p.Limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}

	httputil.JSONWithMeta(w, http.StatusOK, reviews, &httputil.Meta{
		NextCursor: nextCursor,
	})
}

func (h *ReviewHandler) getProductSummary(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productID")

	summary, err := h.svc.GetProductReviewSummary(r.Context(), productID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get review summary")
		return
	}

	httputil.JSON(w, http.StatusOK, summary)
}

func (h *ReviewHandler) createReview(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req model.CreateReviewRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	review, err := h.svc.CreateReview(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create review")
		return
	}

	httputil.JSON(w, http.StatusCreated, review)
}

func (h *ReviewHandler) updateReview(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	reviewID := chi.URLParam(r, "id")

	var req model.UpdateReviewRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	review, err := h.svc.UpdateReview(r.Context(), userID, reviewID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotOwner):
			httputil.Error(w, http.StatusForbidden, "access denied")
		case errors.Is(err, repository.ErrNotFound):
			httputil.Error(w, http.StatusNotFound, "review not found")
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to update review")
		}
		return
	}

	httputil.JSON(w, http.StatusOK, review)
}

func (h *ReviewHandler) deleteReview(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	reviewID := chi.URLParam(r, "id")

	err := h.svc.DeleteReview(r.Context(), userID, reviewID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotOwner):
			httputil.Error(w, http.StatusForbidden, "access denied")
		case errors.Is(err, repository.ErrNotFound):
			httputil.Error(w, http.StatusNotFound, "review not found")
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to delete review")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
