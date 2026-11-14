package models

import "time"

// Category represents a category of lost or found items in the database.
type Category struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Icon      string    `json:"icon" db:"icon"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CategoryResponse represents the public category payload returned in API responses.
type CategoryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a Category entity into its public API response DTO.
func (c *Category) ToResponse() *CategoryResponse {
	if c == nil {
		return nil
	}
	return &CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Icon:      c.Icon,
		CreatedAt: c.CreatedAt,
	}
}

// ToCategoryResponseList transforms a slice of Category entities into a slice of CategoryResponse DTOs.
func ToCategoryResponseList(categories []Category) []CategoryResponse {
	if len(categories) == 0 {
		return []CategoryResponse{}
	}
	responses := make([]CategoryResponse, len(categories))
	for i := range categories {
		responses[i] = *categories[i].ToResponse()
	}
	return responses
}
