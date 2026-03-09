package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kyle/product/internal/platform/auth"
	"github.com/kyle/product/internal/user/model"
	"github.com/kyle/product/internal/user/repository"
)

const bcryptCost = 12

// UserService contains the business rules for user management.
type UserService struct {
	repo       repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewUserService constructs a UserService with the required dependencies.
func NewUserService(repo repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &model.User{
		ID:           generateID(),
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         "customer",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, repository.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	resp, err := s.issueTokens(user)
	if err != nil {
		return nil, fmt.Errorf("issue tokens after register: %w", err)
	}
	return resp, nil
}

func (s *UserService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("login: %w", ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("login: %w", ErrInvalidCredentials)
	}

	resp, err := s.issueTokens(user)
	if err != nil {
		return nil, fmt.Errorf("issue tokens after login: %w", err)
	}
	return resp, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("update profile fetch: %w", err)
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update profile save: %w", err)
	}
	return user, nil
}

func (s *UserService) CreateAddress(ctx context.Context, userID string, req *model.CreateAddressRequest) (*model.Address, error) {
	addr := &model.Address{
		ID:        generateID(),
		UserID:    userID,
		Label:     req.Label,
		Street:    req.Street,
		City:      req.City,
		State:     req.State,
		Zip:       req.Zip,
		Country:   req.Country,
		IsDefault: false,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateAddress(ctx, addr); err != nil {
		return nil, fmt.Errorf("create address: %w", err)
	}
	return addr, nil
}

func (s *UserService) ListAddresses(ctx context.Context, userID string) ([]model.Address, error) {
	addrs, err := s.repo.ListAddresses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	return addrs, nil
}

func (s *UserService) DeleteAddress(ctx context.Context, addressID, userID string) error {
	if err := s.repo.DeleteAddress(ctx, addressID, userID); err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	return nil
}

func (s *UserService) issueTokens(user *model.User) (*model.LoginResponse, error) {
	access, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &model.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}

// ErrInvalidCredentials is returned when email or password is wrong.
var ErrInvalidCredentials = errors.New("invalid email or password")
