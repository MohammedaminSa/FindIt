package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/repository"
)

type mockCategoryRepo struct {
	getAllFunc func(ctx context.Context) ([]models.Category, error)
	getByIDFunc func(ctx context.Context, id int64) (*models.Category, error)
}

func (m *mockCategoryRepo) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockCategoryRepo) GetCategoryByID(ctx context.Context, id int64) (*models.Category, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func TestDefaultCategoryService_GetAllCategories(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("success with categories", func(t *testing.T) {
		repo := &mockCategoryRepo{
			getAllFunc: func(ctx context.Context) ([]models.Category, error) {
				return []models.Category{
					{ID: 1, Name: "Electronics", Icon: "devices", CreatedAt: now},
					{ID: 2, Name: "Keys", Icon: "vpn_key", CreatedAt: now},
				}, nil
			},
		}

		svc := NewCategoryService(repo)
		res, err := svc.GetAllCategories(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(res) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(res))
		}
		if res[0].ID != 1 || res[0].Name != "Electronics" {
			t.Errorf("unexpected category at 0: %+v", res[0])
		}
		if res[1].ID != 2 || res[1].Name != "Keys" {
			t.Errorf("unexpected category at 1: %+v", res[1])
		}
	})

	t.Run("success with empty list", func(t *testing.T) {
		repo := &mockCategoryRepo{
			getAllFunc: func(ctx context.Context) ([]models.Category, error) {
				return []models.Category{}, nil
			},
		}

		svc := NewCategoryService(repo)
		res, err := svc.GetAllCategories(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(res) != 0 {
			t.Fatalf("expected 0 categories, got %d", len(res))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &mockCategoryRepo{
			getAllFunc: func(ctx context.Context) ([]models.Category, error) {
				return nil, errors.New("db query failed")
			},
		}

		svc := NewCategoryService(repo)
		_, err := svc.GetAllCategories(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDefaultCategoryService_GetCategoryByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		repo := &mockCategoryRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*models.Category, error) {
				return &models.Category{
					ID:        id,
					Name:      "Pets",
					Icon:      "pets",
					CreatedAt: now,
				}, nil
			},
		}

		svc := NewCategoryService(repo)
		res, err := svc.GetCategoryByID(ctx, 8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.ID != 8 || res.Name != "Pets" || res.Icon != "pets" {
			t.Errorf("unexpected category result: %+v", res)
		}
	})

	t.Run("invalid ID (<= 0)", func(t *testing.T) {
		svc := NewCategoryService(&mockCategoryRepo{})

		_, err := svc.GetCategoryByID(ctx, 0)
		if !errors.Is(err, ErrInvalidCategoryID) {
			t.Fatalf("expected ErrInvalidCategoryID for 0, got %v", err)
		}

		_, err = svc.GetCategoryByID(ctx, -5)
		if !errors.Is(err, ErrInvalidCategoryID) {
			t.Fatalf("expected ErrInvalidCategoryID for -5, got %v", err)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		repo := &mockCategoryRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*models.Category, error) {
				return nil, repository.ErrCategoryNotFound
			},
		}

		svc := NewCategoryService(repo)
		_, err := svc.GetCategoryByID(ctx, 999)
		if !errors.Is(err, repository.ErrCategoryNotFound) {
			t.Fatalf("expected ErrCategoryNotFound, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &mockCategoryRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*models.Category, error) {
				return nil, errors.New("db error")
			},
		}

		svc := NewCategoryService(repo)
		_, err := svc.GetCategoryByID(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
