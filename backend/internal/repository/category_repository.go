package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohammedaminSa/FindIt/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

var (
	// ErrCategoryNotFound is returned when a requested category record cannot be found.
	ErrCategoryNotFound = errors.New("category not found")
)

// CategoryRepository defines the persistence operations for category data.
type CategoryRepository interface {
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id int64) (*models.Category, error)
}

// PostgresCategoryRepository implements CategoryRepository using PostgreSQL.
type PostgresCategoryRepository struct {
	db DBTX
}

// NewCategoryRepository constructs a new PostgresCategoryRepository.
func NewCategoryRepository(db DBTX) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

// GetAllCategories retrieves all categories ordered by ID.
func (r *PostgresCategoryRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	if r.db == nil {
		return nil, ErrNilDatabase
	}

	query := `
		SELECT id, name, icon, created_at
		FROM categories
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return categories, nil
}

// GetCategoryByID retrieves a single category by its unique ID.
func (r *PostgresCategoryRepository) GetCategoryByID(ctx context.Context, id int64) (*models.Category, error) {
	if r.db == nil {
		return nil, ErrNilDatabase
	}

	query := `
		SELECT id, name, icon, created_at
		FROM categories
		WHERE id = $1
	`

	var cat models.Category
	err := r.db.QueryRow(ctx, query, id).Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to query category by id: %w", err)
	}

	return &cat, nil
}
