package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/order/model"
	"github.com/kyle/product/internal/order/repository"
	"github.com/kyle/product/internal/order/service"
	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/pkg/httputil"
)

// OrderHandler handles HTTP requests for the order API.
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// RegisterRoutes mounts order routes on the given router. All routes require authentication.
func (h *OrderHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Use(authMW)
		r.Post("/", h.createOrder)
		r.Get("/", h.listOrders)
		r.Get("/{id}", h.getOrder)
		r.Post("/{id}/cancel", h.cancelOrder)
	})
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req model.CreateOrderRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrEmptyCart) {
			httputil.Error(w, http.StatusUnprocessableEntity, "cart is empty")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	httputil.JSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	cursor := r.URL.Query().Get("cursor")

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	orders, nextCursor, err := h.svc.ListOrders(r.Context(), userID, cursor, limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list orders")
		return
	}

	httputil.JSONWithMeta(w, http.StatusOK, orders, &httputil.Meta{
		NextCursor: nextCursor,
	})
}

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	orderID := chi.URLParam(r, "id")

	order, err := h.svc.GetOrder(r.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "order not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get order")
		return
	}

	httputil.JSON(w, http.StatusOK, order)
}

func (h *OrderHandler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	orderID := chi.URLParam(r, "id")

	err := h.svc.CancelOrder(r.Context(), userID, orderID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotOwner):
			httputil.Error(w, http.StatusForbidden, "access denied")
		case errors.Is(err, service.ErrNotCancelable):
			httputil.Error(w, http.StatusConflict, "order cannot be cancelled")
		case errors.Is(err, repository.ErrNotFound):
			httputil.Error(w, http.StatusNotFound, "order not found")
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to cancel order")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
