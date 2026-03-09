package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/admin/service"
	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/pkg/dto"
	"github.com/kyle/product/pkg/httputil"
)

// AdminHandler handles HTTP requests for the admin API.
type AdminHandler struct {
	svc *service.AdminService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// RegisterRoutes mounts admin routes on the given router.
// All routes require authentication and admin role.
func (h *AdminHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Use(authMW)
		r.Use(auth.RequireRole("admin"))

		r.Get("/dashboard", h.getDashboard)
		r.Get("/analytics/revenue", h.getRevenue)
		r.Get("/analytics/top-products", h.getTopProducts)
		r.Get("/orders", h.listOrders)
		r.Put("/orders/{id}/status", h.updateOrderStatus)
		r.Get("/users", h.listUsers)
		r.Put("/users/{id}/role", h.updateUserRole)
	})
}

func (h *AdminHandler) getDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetDashboard(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load dashboard")
		return
	}

	httputil.JSON(w, http.StatusOK, stats)
}

func (h *AdminHandler) getRevenue(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			days = parsed
		}
	}

	data, err := h.svc.GetRevenueReport(r.Context(), days)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load revenue report")
		return
	}

	httputil.JSON(w, http.StatusOK, data)
}

func (h *AdminHandler) getTopProducts(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	products, err := h.svc.GetTopProducts(r.Context(), limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load top products")
		return
	}

	httputil.JSON(w, http.StatusOK, products)
}

func (h *AdminHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	p := dto.ParsePagination(r)
	status := r.URL.Query().Get("status")

	orders, nextCursor, err := h.svc.ListAllOrders(r.Context(), status, p.Cursor, p.Limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list orders")
		return
	}

	httputil.JSONWithMeta(w, http.StatusOK, orders, &httputil.Meta{
		NextCursor: nextCursor,
	})
}

func (h *AdminHandler) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status"`
	}
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status == "" {
		httputil.Error(w, http.StatusBadRequest, "status is required")
		return
	}

	if err := h.svc.UpdateOrderStatus(r.Context(), orderID, req.Status); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update order status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	p := dto.ParsePagination(r)

	users, nextCursor, err := h.svc.ListAllUsers(r.Context(), p.Cursor, p.Limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	httputil.JSONWithMeta(w, http.StatusOK, users, &httputil.Meta{
		NextCursor: nextCursor,
	})
}

func (h *AdminHandler) updateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var req struct {
		Role string `json:"role"`
	}
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role == "" {
		httputil.Error(w, http.StatusBadRequest, "role is required")
		return
	}

	if err := h.svc.UpdateUserRole(r.Context(), userID, req.Role); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update user role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
