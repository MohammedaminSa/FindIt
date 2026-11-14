package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockCategoryRows struct {
	data [][]any
	idx  int
	err  error
}

func (m *mockCategoryRows) Close() {}

func (m *mockCategoryRows) Err() error {
	return m.err
}

func (m *mockCategoryRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("")
}

func (m *mockCategoryRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *mockCategoryRows) Next() bool {
	if m.idx < len(m.data) {
		m.idx++
		return true
	}
	return false
}

func (m *mockCategoryRows) Scan(dest ...any) error {
	row := m.data[m.idx-1]
	for i, val := range row {
		switch d := dest[i].(type) {
		case *int64:
			*d = val.(int64)
		case *string:
			*d = val.(string)
		case *time.Time:
			*d = val.(time.Time)
		default:
			return errors.New("unsupported scan destination type")
		}
	}
	return nil
}

func (m *mockCategoryRows) Values() ([]any, error) {
	return m.data[m.idx-1], nil
}

func (m *mockCategoryRows) RawValues() [][]byte {
	return nil
}

func (m *mockCategoryRows) Conn() *pgx.Conn {
	return nil
}

func (m *mockCategoryRows) TypeMap() *pgtype.Map {
	return pgtype.NewMap()
}

func TestPostgresCategoryRepository_GetAllCategories(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mockRows := &mockCategoryRows{
			data: [][]any{
				{int64(1), "Electronics", "devices", now},
				{int64(2), "Keys", "vpn_key", now},
			},
		}

		mockDB := &mockDBTX{
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return mockRows, nil
			},
		}

		repo := NewCategoryRepository(mockDB)
		cats, err := repo.GetAllCategories(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cats) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(cats))
		}
		if cats[0].Name != "Electronics" || cats[1].Name != "Keys" {
			t.Errorf("unexpected categories: %+v", cats)
		}
	})

	t.Run("nil database returns ErrNilDatabase", func(t *testing.T) {
		repo := NewCategoryRepository(nil)
		_, err := repo.GetAllCategories(ctx)
		if !errors.Is(err, ErrNilDatabase) {
			t.Fatalf("expected ErrNilDatabase, got %v", err)
		}
	})

	t.Run("query execution error", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return nil, errors.New("db connection closed")
			},
		}

		repo := NewCategoryRepository(mockDB)
		_, err := repo.GetAllCategories(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("rows error", func(t *testing.T) {
		mockRows := &mockCategoryRows{
			data: [][]any{},
			err:  errors.New("iteration failed"),
		}

		mockDB := &mockDBTX{
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return mockRows, nil
			},
		}

		repo := NewCategoryRepository(mockDB)
		_, err := repo.GetAllCategories(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPostgresCategoryRepository_GetCategoryByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						*dest[0].(*int64) = int64(1)
						*dest[1].(*string) = "Electronics"
						*dest[2].(*string) = "devices"
						*dest[3].(*time.Time) = now
						return nil
					},
				}
			},
		}

		repo := NewCategoryRepository(mockDB)
		cat, err := repo.GetCategoryByID(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cat.ID != 1 || cat.Name != "Electronics" || cat.Icon != "devices" {
			t.Errorf("unexpected category data: %+v", cat)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return pgx.ErrNoRows
					},
				}
			},
		}

		repo := NewCategoryRepository(mockDB)
		_, err := repo.GetCategoryByID(ctx, 999)
		if !errors.Is(err, ErrCategoryNotFound) {
			t.Fatalf("expected ErrCategoryNotFound, got %v", err)
		}
	})

	t.Run("nil database returns ErrNilDatabase", func(t *testing.T) {
		repo := NewCategoryRepository(nil)
		_, err := repo.GetCategoryByID(ctx, 1)
		if !errors.Is(err, ErrNilDatabase) {
			t.Fatalf("expected ErrNilDatabase, got %v", err)
		}
	})

	t.Run("internal query error", func(t *testing.T) {
		mockDB := &mockDBTX{
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						return errors.New("timeout")
					},
				}
			},
		}

		repo := NewCategoryRepository(mockDB)
		_, err := repo.GetCategoryByID(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
