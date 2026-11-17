package service

import (
	"context"

	"github.com/auction-studio/backend/internal/models"
	"github.com/auction-studio/backend/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetAll(ctx context.Context) ([]models.Category, error) {
	return s.repo.GetAll(ctx)
}
