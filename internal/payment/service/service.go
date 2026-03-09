package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kyle/product/internal/payment/model"
	"github.com/kyle/product/internal/payment/repository"
	"github.com/kyle/product/internal/platform/kafka"
	"github.com/kyle/product/pkg/event"
)

type PaymentService struct {
	repo     repository.PaymentRepository
	producer *kafka.Producer
}

func NewPaymentService(repo repository.PaymentRepository, producer *kafka.Producer) *PaymentService {
	return &PaymentService{
		repo:     repo,
		producer: producer,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req model.ProcessPaymentRequest) (*model.Payment, error) {
	now := time.Now()
	payment := &model.Payment{
		ID:         generateID(),
		OrderID:    req.OrderID,
		Amount:     req.Amount,
		Method:     req.Method,
		Status:     model.StatusPending,
		ExternalID: "ext_" + generateID(),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	// In dev, payment always succeeds immediately.
	payment.Status = model.StatusCompleted
	payment.UpdatedAt = time.Now()

	if err := s.repo.UpdateStatus(ctx, payment.ID, model.StatusCompleted); err != nil {
		return nil, fmt.Errorf("update payment status: %w", err)
	}

	evt := event.PaymentCompleted{
		OrderID:   payment.OrderID,
		PaymentID: payment.ID,
		Amount:    payment.Amount,
		Method:    payment.Method,
		PaidAt:    payment.UpdatedAt,
	}
	if err := s.producer.Publish(ctx, event.TopicPaymentDone, payment.OrderID, evt); err != nil {
		return nil, fmt.Errorf("publish payment.completed: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload model.WebhookPayload) error {
	status := model.PaymentStatus(payload.Status)

	if err := s.repo.UpdateStatus(ctx, payload.PaymentID, status); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	payment, err := s.repo.GetByID(ctx, payload.PaymentID)
	if err != nil {
		return fmt.Errorf("get payment: %w", err)
	}

	switch status {
	case model.StatusCompleted:
		evt := event.PaymentCompleted{
			OrderID:   payment.OrderID,
			PaymentID: payment.ID,
			Amount:    payment.Amount,
			Method:    payment.Method,
			PaidAt:    payment.UpdatedAt,
		}
		if err := s.producer.Publish(ctx, event.TopicPaymentDone, payment.OrderID, evt); err != nil {
			return fmt.Errorf("publish payment.completed: %w", err)
		}
	case model.StatusFailed:
		evt := event.PaymentFailed{
			OrderID:   payment.OrderID,
			PaymentID: payment.ID,
			Reason:    "payment declined",
			FailedAt:  payment.UpdatedAt,
		}
		if err := s.producer.Publish(ctx, event.TopicPaymentFailed, payment.OrderID, evt); err != nil {
			return fmt.Errorf("publish payment.failed: %w", err)
		}
	}

	return nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (*model.Payment, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return p, nil
}

func (s *PaymentService) GetPaymentByOrder(ctx context.Context, orderID string) (*model.Payment, error) {
	p, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get payment by order: %w", err)
	}
	return p, nil
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("generate id: %v", err))
	}
	return hex.EncodeToString(b)
}
