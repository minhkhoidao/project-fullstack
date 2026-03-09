package service

import (
	"context"
	"fmt"

	"github.com/kyle/product/internal/admin/model"
	"github.com/kyle/product/internal/admin/repository"
)

// AdminService implements admin business logic.
type AdminService struct {
	repo repository.AdminRepository
}

// NewAdminService creates an AdminService with the given repository.
func NewAdminService(repo repository.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

// GetDashboard returns high-level platform metrics.
func (s *AdminService) GetDashboard(ctx context.Context) (*model.DashboardStats, error) {
	stats, err := s.repo.GetDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get dashboard stats: %w", err)
	}
	return stats, nil
}

// GetRevenueReport returns daily revenue for the last N days.
func (s *AdminService) GetRevenueReport(ctx context.Context, days int) ([]model.RevenueByDay, error) {
	if days <= 0 {
		days = 30
	}

	data, err := s.repo.GetRevenueByDay(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("get revenue report: %w", err)
	}
	return data, nil
}

// GetTopProducts returns the top-selling products by quantity.
func (s *AdminService) GetTopProducts(ctx context.Context, limit int) ([]model.TopProduct, error) {
	if limit <= 0 {
		limit = 10
	}

	products, err := s.repo.GetTopProducts(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("get top products: %w", err)
	}
	return products, nil
}

// ListAllOrders returns a paginated list of all orders, optionally filtered by status.
func (s *AdminService) ListAllOrders(ctx context.Context, status, cursor string, limit int) ([]model.OrderSummary, string, error) {
	orders, nextCursor, err := s.repo.ListOrders(ctx, status, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("list orders: %w", err)
	}
	return orders, nextCursor, nil
}

// ListAllUsers returns a paginated list of all users.
func (s *AdminService) ListAllUsers(ctx context.Context, cursor string, limit int) ([]model.UserSummary, string, error) {
	users, nextCursor, err := s.repo.ListUsers(ctx, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("list users: %w", err)
	}
	return users, nextCursor, nil
}

// UpdateOrderStatus changes the status of an order.
func (s *AdminService) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	if err := s.repo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

// UpdateUserRole changes the role of a user.
func (s *AdminService) UpdateUserRole(ctx context.Context, userID, role string) error {
	if err := s.repo.UpdateUserRole(ctx, userID, role); err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	return nil
}
