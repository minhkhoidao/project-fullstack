package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/cart/model"
	"github.com/kyle/product/internal/cart/service"
	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/pkg/httputil"
)

// CartHandler handles HTTP requests for the cart API.
type CartHandler struct {
	svc *service.CartService
}

// NewCartHandler creates a new CartHandler.
func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

// RegisterRoutes mounts cart routes on the given router. All routes require authentication.
func (h *CartHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1/cart", func(r chi.Router) {
		r.Use(authMW)
		r.Get("/", h.getCart)
		r.Post("/items", h.addItem)
		r.Put("/items/{variantID}", h.updateItem)
		r.Delete("/items/{variantID}", h.removeItem)
		r.Delete("/", h.clearCart)
	})
}

func (h *CartHandler) getCart(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	cart, err := h.svc.GetCart(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get cart")
		return
	}

	httputil.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) addItem(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req model.AddItemRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := h.svc.AddItem(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to add item")
		return
	}

	httputil.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) updateItem(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	variantID := chi.URLParam(r, "variantID")

	var req model.UpdateItemRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := h.svc.UpdateItem(r.Context(), userID, variantID, req)
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			httputil.Error(w, http.StatusNotFound, "item not found in cart")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update item")
		return
	}

	httputil.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) removeItem(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	variantID := chi.URLParam(r, "variantID")

	cart, err := h.svc.RemoveItem(r.Context(), userID, variantID)
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			httputil.Error(w, http.StatusNotFound, "item not found in cart")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to remove item")
		return
	}

	httputil.JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) clearCart(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	if err := h.svc.ClearCart(r.Context(), userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to clear cart")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
