package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	kafkago "github.com/segmentio/kafka-go"

	"github.com/kyle/product/internal/inventory/handler"
	"github.com/kyle/product/internal/inventory/repository"
	"github.com/kyle/product/internal/inventory/service"
	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/platform/config"
	"github.com/kyle/product/internal/platform/database"
	"github.com/kyle/product/internal/platform/kafka"
	"github.com/kyle/product/internal/platform/logger"
	"github.com/kyle/product/internal/platform/server"
	"github.com/kyle/product/pkg/event"
	"github.com/kyle/product/pkg/httputil"
)

func main() {
	cfg, err := config.Load("inventory", ":8086")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	slog := logger.New(cfg.ServiceName, cfg.LogLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	repo := repository.NewPGRepository(pool)
	svc := service.NewInventoryService(repo, producer)
	h := handler.NewInventoryHandler(svc)

	orderCreatedConsumer := kafka.NewConsumer(
		cfg.KafkaBrokers, event.TopicOrderCreated, "inventory-order-created", slog,
	)
	defer orderCreatedConsumer.Close()

	orderCancelledConsumer := kafka.NewConsumer(
		cfg.KafkaBrokers, event.TopicOrderCancelled, "inventory-order-cancelled", slog,
	)
	defer orderCancelledConsumer.Close()

	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()

	go func() {
		if err := orderCreatedConsumer.Start(consumerCtx, func(ctx context.Context, msg kafkago.Message) error {
			var evt event.OrderCreated
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				slog.Error("unmarshal order.created", "error", err)
				return err
			}
			return svc.HandleOrderCreated(ctx, evt)
		}); err != nil {
			slog.Error("order.created consumer stopped", "error", err)
		}
	}()

	go func() {
		if err := orderCancelledConsumer.Start(consumerCtx, func(ctx context.Context, msg kafkago.Message) error {
			var evt event.OrderCancelled
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				slog.Error("unmarshal order.cancelled", "error", err)
				return err
			}
			return svc.HandleOrderCancelled(ctx, evt)
		}); err != nil {
			slog.Error("order.cancelled consumer stopped", "error", err)
		}
	}()

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
