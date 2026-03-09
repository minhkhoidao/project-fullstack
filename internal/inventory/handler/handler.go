package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/inventory/model"
	"github.com/kyle/product/internal/inventory/repository"
	"github.com/kyle/product/internal/inventory/service"
	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/pkg/httputil"
)

type InventoryHandler struct {
	svc *service.InventoryService
}

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1/inventory", func(r chi.Router) {
		r.Use(authMW)
		r.Use(auth.RequireRole("admin"))

		r.Get("/low-stock", h.listLowStock)
		r.Get("/{variantID}", h.getStock)
		r.Put("/", h.updateStock)
	})
}

func (h *InventoryHandler) getStock(w http.ResponseWriter, r *http.Request) {
	variantID := chi.URLParam(r, "variantID")
	warehouse := r.URL.Query().Get("warehouse")
	if warehouse == "" {
		warehouse = "default"
	}

	item, err := h.svc.GetStock(r.Context(), variantID, warehouse)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "inventory item not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get stock")
		return
	}

	httputil.JSON(w, http.StatusOK, item)
}

func (h *InventoryHandler) updateStock(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateStockRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ProductVariantID == "" || req.Warehouse == "" {
		httputil.Error(w, http.StatusBadRequest, "product_variant_id and warehouse are required")
		return
	}

	item, err := h.svc.UpdateStock(r.Context(), req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update stock")
		return
	}

	httputil.JSON(w, http.StatusOK, item)
}

func (h *InventoryHandler) listLowStock(w http.ResponseWriter, r *http.Request) {
	threshold := 10
	if v := r.URL.Query().Get("threshold"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err == nil && parsed > 0 {
			threshold = parsed
		}
	}

	items, err := h.svc.ListLowStock(r.Context(), threshold)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list low stock")
		return
	}

	httputil.JSON(w, http.StatusOK, items)
}
