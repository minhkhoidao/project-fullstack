package main

import (
	"context"
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/kyle/product/internal/cart/handler"
	"github.com/kyle/product/internal/cart/repository"
	"github.com/kyle/product/internal/cart/service"
	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/platform/config"
	"github.com/kyle/product/internal/platform/logger"
	"github.com/kyle/product/internal/platform/redisclient"
	"github.com/kyle/product/internal/platform/server"
	"github.com/kyle/product/pkg/httputil"
)

func main() {
	cfg, err := config.Load("cart", ":8083")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	slog := logger.New(cfg.ServiceName, cfg.LogLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rdb, err := redisclient.New(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer rdb.Close()

	repo := repository.NewRedisRepository(rdb)
	svc := service.NewCartService(repo)
	h := handler.NewCartHandler(svc)

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshExpiry)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(httputil.RequestLogger(slog))

	h.RegisterRoutes(r, jwtMgr.AuthMiddleware)

	srv := server.New(cfg.HTTPAddr, r, slog, cfg.ShutdownTimeout)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
