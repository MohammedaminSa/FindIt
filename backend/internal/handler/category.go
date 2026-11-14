package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// CategoryHandler handles HTTP requests for item categories.
type CategoryHandler struct {
	categoryService service.CategoryService
}

// NewCategoryHandler constructs a new CategoryHandler.
func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// RegisterRoutes registers category endpoints under the provided router group.
func (h *CategoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	categories := rg.Group("/categories")
	{
		categories.GET("", h.GetAll)
		categories.GET("/:id", h.GetByID)
	}
}

// GetAll retrieves all active item categories.
// GET /api/v1/categories
func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve categories",
		})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetByID retrieves a single item category by its primary key ID.
// GET /api/v1/categories/:id
func (h *CategoryHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	cat, err := h.categoryService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve category",
		})
		return
	}

	c.JSON(http.StatusOK, cat)
}
