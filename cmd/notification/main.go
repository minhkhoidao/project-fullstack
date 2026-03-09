package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/kyle/product/internal/notification/service"
	"github.com/kyle/product/internal/platform/config"
	"github.com/kyle/product/internal/platform/kafka"
	"github.com/kyle/product/internal/platform/logger"
	"github.com/kyle/product/pkg/event"
)

func main() {
	cfg, err := config.Load("notification", ":8088")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	slog := logger.New(cfg.ServiceName, cfg.LogLevel)
	svc := service.NewNotificationService(slog)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type topicHandler struct {
		topic   string
		handler kafka.MessageHandler
	}

	topics := []topicHandler{
		{
			topic: event.TopicOrderCreated,
			handler: func(ctx context.Context, msg kafkago.Message) error {
				var evt event.OrderCreated
				if err := json.Unmarshal(msg.Value, &evt); err != nil {
					return err
				}
				return svc.HandleOrderCreated(ctx, evt)
			},
		},
		{
			topic: event.TopicOrderPaid,
			handler: func(ctx context.Context, msg kafkago.Message) error {
				var evt event.OrderPaid
				if err := json.Unmarshal(msg.Value, &evt); err != nil {
					return err
				}
				return svc.HandleOrderPaid(ctx, evt)
			},
		},
		{
			topic: event.TopicOrderCancelled,
			handler: func(ctx context.Context, msg kafkago.Message) error {
				var evt event.OrderCancelled
				if err := json.Unmarshal(msg.Value, &evt); err != nil {
					return err
				}
				return svc.HandleOrderCancelled(ctx, evt)
			},
		},
		{
			topic: event.TopicPaymentDone,
			handler: func(ctx context.Context, msg kafkago.Message) error {
				var evt event.PaymentCompleted
				if err := json.Unmarshal(msg.Value, &evt); err != nil {
					return err
				}
				return svc.HandleOrderPaid(ctx, event.OrderPaid{
					OrderID:   evt.OrderID,
					PaymentID: evt.PaymentID,
					Amount:    evt.Amount,
					PaidAt:    evt.PaidAt,
				})
			},
		},
		{
			topic: event.TopicPaymentFailed,
			handler: func(ctx context.Context, msg kafkago.Message) error {
				var evt event.PaymentFailed
				if err := json.Unmarshal(msg.Value, &evt); err != nil {
					return err
				}
				return svc.HandlePaymentFailed(ctx, evt)
			},
		},
		{
			topic: event.TopicInventoryLow,
			handler: func(ctx context.Context, msg kafkago.Message) error {
				var evt event.InventoryLow
				if err := json.Unmarshal(msg.Value, &evt); err != nil {
					return err
				}
				return svc.HandleLowStock(ctx, evt)
			},
		},
	}

	consumers := make([]*kafka.Consumer, 0, len(topics))
	for _, th := range topics {
		c := kafka.NewConsumer(cfg.KafkaBrokers, th.topic, "notification-service", slog)
		consumers = append(consumers, c)

		go func(consumer *kafka.Consumer, handler kafka.MessageHandler, topic string) {
			slog.Info("starting consumer", "topic", topic)
			if err := consumer.Start(ctx, handler); err != nil {
				slog.Error("consumer stopped", "topic", topic, "error", err)
			}
		}(c, th.handler, th.topic)
	}

	slog.Info("notification service started", "addr", cfg.HTTPAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down notification service")
	cancel()

	for _, c := range consumers {
		if err := c.Close(); err != nil {
			slog.Error("close consumer", "error", err)
		}
	}

	slog.Info("notification service stopped")
}
