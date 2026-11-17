package service

import (
	"context"
	"errors"
	"time"

	"github.com/auction-studio/backend/internal/auth"
	"github.com/auction-studio/backend/internal/config"
	"github.com/auction-studio/backend/internal/dto"
	"github.com/auction-studio/backend/internal/models"
	"github.com/auction-studio/backend/internal/repository"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, errors.New("name, email and password are required")
	}

	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           "usr-" + uuid.New().String()[:8],
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         models.RoleBuyer,
		Phone:        req.Phone,
		Location:     req.Location,
		Rating:       5.0,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	rt, err := s.userRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil || rt == nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Delete old token
	_ = s.userRepo.DeleteRefreshToken(ctx, refreshToken)

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Phone:     user.Phone,
		Location:  user.Location,
		Rating:    user.Rating,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *models.User) (*dto.AuthResponse, error) {
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenModel := &models.RefreshToken{
		ID:        "rf-" + uuid.New().String()[:8],
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWTRefreshExpiry),
		CreatedAt: time.Now(),
	}
	if err := s.userRepo.SaveRefreshToken(ctx, tokenModel); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			Phone:     user.Phone,
			Location:  user.Location,
			Rating:    user.Rating,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
