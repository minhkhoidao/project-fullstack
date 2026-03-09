package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/product/model"
	"github.com/kyle/product/internal/product/service"
	"github.com/kyle/product/pkg/dto"
	"github.com/kyle/product/pkg/httputil"
)

// ProductHandler exposes HTTP endpoints for the product domain.
type ProductHandler struct {
	svc *service.ProductService
}

// NewProductHandler returns a handler wired to the given service.
func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// RegisterRoutes mounts public and admin product routes on the router.
func (h *ProductHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1", func(r chi.Router) {
		// Public
		r.Get("/products", h.listProducts)
		r.Get("/products/{idOrSlug}", h.getProduct)
		r.Get("/categories", h.listCategories)

		// Admin
		r.Group(func(r chi.Router) {
			r.Use(authMW)
			r.Use(auth.RequireRole("admin"))

			r.Post("/admin/products", h.createProduct)
			r.Put("/admin/products/{id}", h.updateProduct)
			r.Delete("/admin/products/{id}", h.deleteProduct)
			r.Post("/admin/categories", h.createCategory)
		})
	})
}

// ---------------------------------------------------------------------------
// Public
// ---------------------------------------------------------------------------

func (h *ProductHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	filter := parseFilter(r)
	pg := dto.ParsePagination(r)

	products, nextCursor, err := h.svc.ListProducts(r.Context(), filter, pg.Cursor, pg.Limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list products")
		return
	}

	httputil.JSONWithMeta(w, http.StatusOK, products, &httputil.Meta{
		NextCursor: nextCursor,
		PerPage:    pg.Limit,
	})
}

func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "idOrSlug")

	var (
		product *model.Product
		err     error
	)
	if looksLikeUUID(idOrSlug) {
		product, err = h.svc.GetProduct(r.Context(), idOrSlug)
	} else {
		product, err = h.svc.GetProductBySlug(r.Context(), idOrSlug)
	}

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "product not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get product")
		return
	}

	httputil.JSON(w, http.StatusOK, product)
}

func (h *ProductHandler) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.ListCategories(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	httputil.JSON(w, http.StatusOK, cats)
}

// ---------------------------------------------------------------------------
// Admin
// ---------------------------------------------------------------------------

func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.CreateProduct(r.Context(), req)
	if err != nil {
		httputil.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	httputil.JSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req model.UpdateProductRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.UpdateProduct(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "product not found")
			return
		}
		httputil.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, product)
}

func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.DeleteProduct(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "product not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type createCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *ProductHandler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req createCategoryRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), req.Name, req.Description)
	if err != nil {
		httputil.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	httputil.JSON(w, http.StatusCreated, cat)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseFilter(r *http.Request) model.ProductFilter {
	q := r.URL.Query()
	f := model.ProductFilter{
		CategoryID: q.Get("category"),
		Search:     q.Get("search"),
		Size:       q.Get("size"),
		Color:      q.Get("color"),
	}
	if v := q.Get("min_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			f.MinPrice = &p
		}
	}
	if v := q.Get("max_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			f.MaxPrice = &p
		}
	}
	return f
}

// looksLikeUUID returns true when the string matches a UUID-style format
// (8-4-4-4-12 hex characters) so the handler can distinguish IDs from slugs.
func looksLikeUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
