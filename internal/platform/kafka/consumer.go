package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/segmentio/kafka-go"
)

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	reader *kafka.Reader
	log    *slog.Logger
}

func NewConsumer(brokers, topic, groupID string, log *slog.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(brokers, ","),
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader: r,
		log:    log,
	}
}

func (c *Consumer) Start(ctx context.Context, handler MessageHandler) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		if err := handler(ctx, msg); err != nil {
			c.log.Error("handle message",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"error", err,
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("commit message",
				"topic", msg.Topic,
				"offset", msg.Offset,
				"error", err,
			)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
