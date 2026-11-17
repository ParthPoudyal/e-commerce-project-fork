package service

import (
	"context"

	"github.com/auction-studio/backend/internal/repository"
)

type WishlistService struct {
	wishlistRepo *repository.WishlistRepository
}

func NewWishlistService(wishlistRepo *repository.WishlistRepository) *WishlistService {
	return &WishlistService{wishlistRepo: wishlistRepo}
}

func (s *WishlistService) GetWishlist(ctx context.Context, userID string) ([]string, error) {
	return s.wishlistRepo.GetItemIDsByUserID(ctx, userID)
}

func (s *WishlistService) Toggle(ctx context.Context, userID, itemID string) (bool, error) {
	exists, err := s.wishlistRepo.Exists(ctx, userID, itemID)
	if err != nil {
		return false, err
	}

	if exists {
		err := s.wishlistRepo.Remove(ctx, userID, itemID)
		return false, err
	}

	err = s.wishlistRepo.Add(ctx, userID, itemID)
	return true, err
}

func (s *WishlistService) Remove(ctx context.Context, userID, itemID string) error {
	return s.wishlistRepo.Remove(ctx, userID, itemID)
}
