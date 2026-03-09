package main

import (
	"context"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/platform/config"
	"github.com/kyle/product/internal/platform/database"
	"github.com/kyle/product/internal/platform/logger"
	"github.com/kyle/product/internal/platform/server"
	"github.com/kyle/product/internal/user/handler"
	"github.com/kyle/product/internal/user/repository"
	"github.com/kyle/product/internal/user/service"
	"github.com/kyle/product/pkg/httputil"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("user service: %v", err)
	}
}

func run() error {
	cfg, err := config.Load("user-service", ":8081")
	if err != nil {
		return err
	}

	slog := logger.New(cfg.ServiceName, cfg.LogLevel)

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshExpiry)
	repo := repository.NewPostgresRepository(pool)
	svc := service.NewUserService(repo, jwtMgr)
	h := handler.NewUserHandler(svc)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(httputil.RequestLogger(slog))

	h.RegisterRoutes(r, jwtMgr.AuthMiddleware)

	srv := server.New(cfg.HTTPAddr, r, slog, cfg.ShutdownTimeout)
	return srv.Run()
}
