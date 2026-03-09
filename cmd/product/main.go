package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/platform/config"
	"github.com/kyle/product/internal/platform/database"
	"github.com/kyle/product/internal/platform/logger"
	"github.com/kyle/product/internal/platform/server"
	"github.com/kyle/product/internal/product/handler"
	"github.com/kyle/product/internal/product/repository"
	"github.com/kyle/product/internal/product/service"
	"github.com/kyle/product/pkg/httputil"
)

func main() {
	cfg, err := config.Load("product-service", ":8082")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		return
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)
	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshExpiry)

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(httputil.RequestLogger(log))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	h.RegisterRoutes(r, jwtMgr.AuthMiddleware)

	srv := server.New(cfg.HTTPAddr, r, log, cfg.ShutdownTimeout)
	if err := srv.Run(); err != nil {
		log.Error("server error", "error", err)
	}
}
