package service

import (
	"context"
	"errors"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/repository"
)

var (
	// ErrCategoryNotFound is returned when a requested category does not exist.
	ErrCategoryNotFound = repository.ErrCategoryNotFound
	// ErrInvalidCategoryID is returned when an invalid category ID (<= 0) is supplied.
	ErrInvalidCategoryID = errors.New("invalid category ID")
)

// CategoryService defines business operations for managing categories.
type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error)
	GetCategoryByID(ctx context.Context, id int64) (*models.CategoryResponse, error)
}

// DefaultCategoryService implements CategoryService.
type DefaultCategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService constructs a new DefaultCategoryService.
func NewCategoryService(categoryRepo repository.CategoryRepository) *DefaultCategoryService {
	return &DefaultCategoryService{
		categoryRepo: categoryRepo,
	}
}

// GetAllCategories retrieves all available item categories.
func (s *DefaultCategoryService) GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error) {
	categories, err := s.categoryRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	return models.ToCategoryResponseList(categories), nil
}

// GetCategoryByID retrieves a single category by its ID.
func (s *DefaultCategoryService) GetCategoryByID(ctx context.Context, id int64) (*models.CategoryResponse, error) {
	if id <= 0 {
		return nil, ErrInvalidCategoryID
	}

	cat, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return cat.ToResponse(), nil
}
