package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/handler"
	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/repository"
	"github.com/MohammedaminSa/FindIt/backend/internal/service"
)

// inMemoryCategoryRepo provides an in-memory implementation for integration testing.
type inMemoryCategoryRepo struct {
	mu         sync.RWMutex
	categories []models.Category
}

func newInMemoryCategoryRepo() *inMemoryCategoryRepo {
	now := time.Now().UTC()
	return &inMemoryCategoryRepo{
		categories: []models.Category{
			{ID: 1, Name: "Electronics", Icon: "devices", CreatedAt: now},
			{ID: 2, Name: "Wallets & Cards", Icon: "account_balance_wallet", CreatedAt: now},
			{ID: 3, Name: "Keys", Icon: "vpn_key", CreatedAt: now},
			{ID: 4, Name: "Bags & Luggage", Icon: "backpack", CreatedAt: now},
			{ID: 5, Name: "Documents & IDs", Icon: "badge", CreatedAt: now},
			{ID: 6, Name: "Jewelry & Watches", Icon: "watch", CreatedAt: now},
			{ID: 7, Name: "Clothing", Icon: "checkroom", CreatedAt: now},
			{ID: 8, Name: "Pets", Icon: "pets", CreatedAt: now},
			{ID: 9, Name: "Other", Icon: "category", CreatedAt: now},
		},
	}
}

func (r *inMemoryCategoryRepo) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]models.Category, len(r.categories))
	copy(result, r.categories)
	return result, nil
}

func (r *inMemoryCategoryRepo) GetCategoryByID(ctx context.Context, id int64) (*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.categories {
		if c.ID == id {
			copied := c
			return &copied, nil
		}
	}
	return nil, repository.ErrCategoryNotFound
}

func TestCategoryAPI_Integration(t *testing.T) {
	repo := newInMemoryCategoryRepo()
	svc := service.NewCategoryService(repo)
	catHandler := handler.NewCategoryHandler(svc)
	router := setupRouter(nil, nil, catHandler)

	t.Run("GET /api/v1/categories - list all default categories", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/categories", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d: %s", w.Code, w.Body.String())
		}

		var categories []models.CategoryResponse
		if err := json.Unmarshal(w.Body.Bytes(), &categories); err != nil {
			t.Fatalf("failed to parse categories JSON: %v", err)
		}

		if len(categories) != 9 {
			t.Fatalf("expected 9 seeded categories, got %d", len(categories))
		}

		expectedNames := []string{
			"Electronics",
			"Wallets & Cards",
			"Keys",
			"Bags & Luggage",
			"Documents & IDs",
			"Jewelry & Watches",
			"Clothing",
			"Pets",
			"Other",
		}

		for i, name := range expectedNames {
			if categories[i].ID != int64(i+1) {
				t.Errorf("expected category index %d to have ID %d, got %d", i, i+1, categories[i].ID)
			}
			if categories[i].Name != name {
				t.Errorf("expected category index %d name %q, got %q", i, name, categories[i].Name)
			}
			if categories[i].Icon == "" {
				t.Errorf("expected non-empty icon for category %s", name)
			}
		}
	})

	t.Run("GET /api/v1/categories/:id - retrieve existing category", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d: %s", w.Code, w.Body.String())
		}

		var cat models.CategoryResponse
		if err := json.Unmarshal(w.Body.Bytes(), &cat); err != nil {
			t.Fatalf("failed to parse category JSON: %v", err)
		}

		if cat.ID != 1 || cat.Name != "Electronics" || cat.Icon != "devices" {
			t.Errorf("unexpected category data: %+v", cat)
		}
	})

	t.Run("GET /api/v1/categories/:id - not found", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected HTTP 404, got %d: %s", w.Code, w.Body.String())
		}

		var errRes map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &errRes); err != nil {
			t.Fatalf("failed to parse error JSON: %v", err)
		}

		if errRes["error"] != "Category not found" {
			t.Errorf("expected 'Category not found', got %q", errRes["error"])
		}
	})

	t.Run("GET /api/v1/categories/:id - invalid ID format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected HTTP 400, got %d: %s", w.Code, w.Body.String())
		}

		var errRes map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &errRes); err != nil {
			t.Fatalf("failed to parse error JSON: %v", err)
		}

		if errRes["error"] != "Invalid category ID" {
			t.Errorf("expected 'Invalid category ID', got %q", errRes["error"])
		}
	})

	t.Run("GET /api/v1/categories/:id - zero or negative ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/0", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected HTTP 400, got %d", w.Code)
		}
	})
}
