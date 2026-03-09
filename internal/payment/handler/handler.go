package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/payment/model"
	"github.com/kyle/product/internal/payment/repository"
	"github.com/kyle/product/internal/payment/service"
	"github.com/kyle/product/pkg/httputil"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1/payments", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMW)
			r.Post("/process", h.processPayment)
			r.Get("/{id}", h.getPayment)
		})

		r.Post("/webhook", h.handleWebhook)
	})
}

func (h *PaymentHandler) processPayment(w http.ResponseWriter, r *http.Request) {
	var req model.ProcessPaymentRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OrderID == "" || req.Amount <= 0 || req.Method == "" {
		httputil.Error(w, http.StatusBadRequest, "order_id, amount, and method are required")
		return
	}

	payment, err := h.svc.ProcessPayment(r.Context(), req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to process payment")
		return
	}

	httputil.JSON(w, http.StatusCreated, payment)
}

func (h *PaymentHandler) getPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	payment, err := h.svc.GetPayment(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "payment not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get payment")
		return
	}

	httputil.JSON(w, http.StatusOK, payment)
}

func (h *PaymentHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload model.WebhookPayload
	if err := httputil.Decode(r, &payload); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	if payload.PaymentID == "" || payload.Status == "" {
		httputil.Error(w, http.StatusBadRequest, "payment_id and status are required")
		return
	}

	if err := h.svc.HandleWebhook(r.Context(), payload); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "payment not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to handle webhook")
		return
	}

	w.WriteHeader(http.StatusOK)
}
