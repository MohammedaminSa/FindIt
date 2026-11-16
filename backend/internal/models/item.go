package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ItemType defines whether an item is lost or found.
type ItemType string

const (
	// ItemTypeLost represents an item that has been lost by its owner.
	ItemTypeLost ItemType = "LOST"
	// ItemTypeFound represents an item that has been discovered by someone.
	ItemTypeFound ItemType = "FOUND"
)

// IsValid checks if the ItemType is recognized by FindIt.
func (t ItemType) IsValid() bool {
	switch t {
	case ItemTypeLost, ItemTypeFound:
		return true
	default:
		return false
	}
}

// ItemStatus defines the operational lifecycle state of an item.
type ItemStatus string

const (
	// ItemStatusActive represents an open listing waiting for owner/finder matching.
	ItemStatusActive ItemStatus = "ACTIVE"
	// ItemStatusClaimed represents an item with an approved ownership claim.
	ItemStatusClaimed ItemStatus = "CLAIMED"
	// ItemStatusResolved represents an item safely returned or concluded.
	ItemStatusResolved ItemStatus = "RESOLVED"
)

// IsValid checks if the ItemStatus is recognized by FindIt.
func (s ItemStatus) IsValid() bool {
	switch s {
	case ItemStatusActive, ItemStatusClaimed, ItemStatusResolved:
		return true
	default:
		return false
	}
}

// FlexibleTime handles various ISO-8601, RFC3339, and date formats across JSON and form submissions.
type FlexibleTime time.Time

// UnmarshalText implements encoding.TextUnmarshaler for form-data and JSON binding.
func (ft *FlexibleTime) UnmarshalText(text []byte) error {
	raw := strings.Trim(string(text), "\" \t\r\n")
	if raw == "" {
		*ft = FlexibleTime(time.Time{})
		return nil
	}

	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range formats {
		if parsed, err := time.Parse(layout, raw); err == nil {
			*ft = FlexibleTime(parsed)
			return nil
		}
	}

	return fmt.Errorf("unable to parse date '%s' (expected RFC3339 or YYYY-MM-DD)", raw)
}

// UnmarshalJSON implements json.Unmarshaler.
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	return ft.UnmarshalText(b)
}

// MarshalJSON implements json.Marshaler, outputting standard RFC3339 format.
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	t := time.Time(ft)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(time.RFC3339))
}

// Time returns the underlying time.Time value.
func (ft FlexibleTime) Time() time.Time {
	return time.Time(ft)
}

// Item represents a lost or found item record stored in PostgreSQL.
type Item struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	Type            ItemType   `json:"type" db:"type"`
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description" db:"description"`
	CategoryID      int64      `json:"category_id" db:"category_id"`
	LocationName    string     `json:"location_name" db:"location_name"`
	Latitude        *float64   `json:"latitude,omitempty" db:"latitude"`
	Longitude       *float64   `json:"longitude,omitempty" db:"longitude"`
	DateLostOrFound time.Time  `json:"date_lost_or_found" db:"date_lost_or_found"`
	Status          ItemStatus `json:"status" db:"status"`
	ImageURLs       []string   `json:"image_urls" db:"image_urls"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateItemRequest encapsulates user inputs for reporting a new item.
type CreateItemRequest struct {
	Type            ItemType  `json:"type" form:"type" binding:"required"`
	Title           string    `json:"title" form:"title" binding:"required,min=3,max=150"`
	Description     string    `json:"description" form:"description" binding:"required,min=5,max=2000"`
	CategoryID      int64     `json:"category_id" form:"category_id" binding:"required,gt=0"`
	LocationName    string    `json:"location_name" form:"location_name" binding:"required,min=2,max=255"`
	Latitude        *float64  `json:"latitude,omitempty" form:"latitude"`
	Longitude       *float64  `json:"longitude,omitempty" form:"longitude"`
	DateLostOrFound time.Time `json:"date_lost_or_found" form:"date_lost_or_found" binding:"required"`
	ImageURLs       []string  `json:"image_urls,omitempty" form:"image_urls"`
}

// Validate performs deep domain business validation on the create request.
func (r *CreateItemRequest) Validate() error {
	if !r.Type.IsValid() {
		return errors.New("type must be either 'LOST' or 'FOUND'")
	}

	trimmedTitle := strings.TrimSpace(r.Title)
	if len(trimmedTitle) < 3 || len(trimmedTitle) > 150 {
		return errors.New("title must be between 3 and 150 characters")
	}

	trimmedDesc := strings.TrimSpace(r.Description)
	if len(trimmedDesc) < 5 || len(trimmedDesc) > 2000 {
		return errors.New("description must be between 5 and 2000 characters")
	}

	if r.CategoryID <= 0 {
		return errors.New("category_id must be a positive integer")
	}

	trimmedLocation := strings.TrimSpace(r.LocationName)
	if len(trimmedLocation) < 2 || len(trimmedLocation) > 255 {
		return errors.New("location_name must be between 2 and 255 characters")
	}

	if r.Latitude != nil {
		if *r.Latitude < -90.0 || *r.Latitude > 90.0 {
			return errors.New("latitude must be between -90 and 90 degrees")
		}
	}

	if r.Longitude != nil {
		if *r.Longitude < -180.0 || *r.Longitude > 180.0 {
			return errors.New("longitude must be between -180 and 180 degrees")
		}
	}

	if r.DateLostOrFound.IsZero() {
		return errors.New("date_lost_or_found is required")
	}

	// Prevent reporting dates too far in the future (allowing slight clock drift of 24h)
	if r.DateLostOrFound.After(time.Now().Add(24 * time.Hour)) {
		return errors.New("date_lost_or_found cannot be in the future")
	}

	if len(r.ImageURLs) > 5 {
		return errors.New("an item cannot have more than 5 images")
	}

	return nil
}

// UpdateItemStatusRequest encapsulates the payload to change an item's status.
type UpdateItemStatusRequest struct {
	Status ItemStatus `json:"status" form:"status" binding:"required"`
}

// Validate validates the status update request.
func (r *UpdateItemStatusRequest) Validate() error {
	if !r.Status.IsValid() {
		return errors.New("status must be ACTIVE, CLAIMED, or RESOLVED")
	}
	return nil
}

// ItemResponse is the clean, public DTO returned to API consumers.
type ItemResponse struct {
	ID              uuid.UUID         `json:"id"`
	UserID          uuid.UUID         `json:"user_id"`
	Type            ItemType          `json:"type"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	CategoryID      int64             `json:"category_id"`
	Category        *CategoryResponse `json:"category,omitempty"`
	LocationName    string            `json:"location_name"`
	Latitude        *float64          `json:"latitude,omitempty"`
	Longitude       *float64          `json:"longitude,omitempty"`
	DateLostOrFound time.Time         `json:"date_lost_or_found"`
	Status          ItemStatus        `json:"status"`
	ImageURLs       []string          `json:"image_urls"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// ToResponse converts an Item entity into its public ItemResponse DTO.
func (item *Item) ToResponse() *ItemResponse {
	if item == nil {
		return nil
	}

	imageURLs := item.ImageURLs
	if imageURLs == nil {
		imageURLs = []string{}
	}

	return &ItemResponse{
		ID:              item.ID,
		UserID:          item.UserID,
		Type:            item.Type,
		Title:           item.Title,
		Description:     item.Description,
		CategoryID:      item.CategoryID,
		LocationName:    item.LocationName,
		Latitude:        item.Latitude,
		Longitude:       item.Longitude,
		DateLostOrFound: item.DateLostOrFound,
		Status:          item.Status,
		ImageURLs:       imageURLs,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

// ToItemResponseList transforms a slice of Item entities into ItemResponse DTOs safely.
func ToItemResponseList(items []Item) []ItemResponse {
	if len(items) == 0 {
		return []ItemResponse{}
	}

	responses := make([]ItemResponse, len(items))
	for i := range items {
		responses[i] = *items[i].ToResponse()
	}
	return responses
}
