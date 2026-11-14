package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type mockCategoryService struct {
	getAllFunc  func(ctx context.Context) ([]models.CategoryResponse, error)
	getByIDFunc func(ctx context.Context, id int64) (*models.CategoryResponse, error)
}

func (m *mockCategoryService) GetAllCategories(ctx context.Context) ([]models.CategoryResponse, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockCategoryService) GetCategoryByID(ctx context.Context, id int64) (*models.CategoryResponse, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func setupCategoryTestRouter(svc service.CategoryService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewCategoryHandler(svc)
	v1 := router.Group("/api/v1")
	h.RegisterRoutes(v1)
	return router
}

func TestCategoryHandler_GetAll(t *testing.T) {
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		svc := &mockCategoryService{
			getAllFunc: func(ctx context.Context) ([]models.CategoryResponse, error) {
				return []models.CategoryResponse{
					{ID: 1, Name: "Electronics", Icon: "devices", CreatedAt: now},
					{ID: 2, Name: "Keys", Icon: "vpn_key", CreatedAt: now},
				}, nil
			},
		}

		router := setupCategoryTestRouter(svc)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var res []models.CategoryResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if len(res) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(res))
		}
		if res[0].Name != "Electronics" || res[1].Name != "Keys" {
			t.Errorf("unexpected categories: %+v", res)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		svc := &mockCategoryService{
			getAllFunc: func(ctx context.Context) ([]models.CategoryResponse, error) {
				return nil, errors.New("db error")
			},
		}

		router := setupCategoryTestRouter(svc)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}

func TestCategoryHandler_GetByID(t *testing.T) {
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		svc := &mockCategoryService{
			getByIDFunc: func(ctx context.Context, id int64) (*models.CategoryResponse, error) {
				return &models.CategoryResponse{
					ID:        id,
					Name:      "Wallets & Cards",
					Icon:      "account_balance_wallet",
					CreatedAt: now,
				}, nil
			},
		}

		router := setupCategoryTestRouter(svc)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var res models.CategoryResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if res.ID != 2 || res.Name != "Wallets & Cards" {
			t.Errorf("unexpected category: %+v", res)
		}
	})

	t.Run("invalid ID string", func(t *testing.T) {
		svc := &mockCategoryService{}
		router := setupCategoryTestRouter(svc)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/not-a-number", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid ID non-positive", func(t *testing.T) {
		svc := &mockCategoryService{}
		router := setupCategoryTestRouter(svc)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/0", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockCategoryService{
			getByIDFunc: func(ctx context.Context, id int64) (*models.CategoryResponse, error) {
				return nil, service.ErrCategoryNotFound
			},
		}

		router := setupCategoryTestRouter(svc)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		svc := &mockCategoryService{
			getByIDFunc: func(ctx context.Context, id int64) (*models.CategoryResponse, error) {
				return nil, errors.New("unexpected error")
			},
		}

		router := setupCategoryTestRouter(svc)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}
