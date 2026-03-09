package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/user/model"
	"github.com/kyle/product/internal/user/repository"
	"github.com/kyle/product/internal/user/service"
	"github.com/kyle/product/pkg/httputil"
)

// UserHandler exposes HTTP endpoints for user operations.
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes mounts all user-related routes onto the given router.
func (h *UserHandler) RegisterRoutes(r chi.Router, authMW func(http.Handler) http.Handler) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.register)
		r.Post("/auth/login", h.login)

		r.Group(func(r chi.Router) {
			r.Use(authMW)
			r.Get("/users/me", h.getProfile)
			r.Put("/users/me", h.updateProfile)
			r.Post("/users/me/addresses", h.createAddress)
			r.Get("/users/me/addresses", h.listAddresses)
			r.Delete("/users/me/addresses/{id}", h.deleteAddress)
		})
	})
}

func (h *UserHandler) register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	resp, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			httputil.Error(w, http.StatusConflict, "email already registered")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "registration failed")
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	resp, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httputil.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "login failed")
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func (h *UserHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not fetch profile")
		return
	}

	httputil.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) updateProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	var req model.UpdateProfileRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not update profile")
		return
	}

	httputil.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) createAddress(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	var req model.CreateAddressRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Street == "" || req.City == "" || req.Country == "" {
		httputil.Error(w, http.StatusBadRequest, "street, city, and country are required")
		return
	}

	addr, err := h.svc.CreateAddress(r.Context(), userID, &req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create address")
		return
	}

	httputil.JSON(w, http.StatusCreated, addr)
}

func (h *UserHandler) listAddresses(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	addrs, err := h.svc.ListAddresses(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not list addresses")
		return
	}

	httputil.JSON(w, http.StatusOK, addrs)
}

func (h *UserHandler) deleteAddress(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	addressID := chi.URLParam(r, "id")
	if addressID == "" {
		httputil.Error(w, http.StatusBadRequest, "address id is required")
		return
	}

	if err := h.svc.DeleteAddress(r.Context(), addressID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "address not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not delete address")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
