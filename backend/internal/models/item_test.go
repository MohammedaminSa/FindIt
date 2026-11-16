package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestItemType_IsValid(t *testing.T) {
	tests := []struct {
		itemType ItemType
		expected bool
	}{
		{ItemTypeLost, true},
		{ItemTypeFound, true},
		{ItemType("INVALID"), false},
		{ItemType(""), false},
	}

	for _, tt := range tests {
		if got := tt.itemType.IsValid(); got != tt.expected {
			t.Errorf("ItemType(%q).IsValid() = %v, expected %v", tt.itemType, got, tt.expected)
		}
	}
}

func TestItemStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   ItemStatus
		expected bool
	}{
		{ItemStatusActive, true},
		{ItemStatusClaimed, true},
		{ItemStatusResolved, true},
		{ItemStatus("ARCHIVED"), false},
		{ItemStatus(""), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.expected {
			t.Errorf("ItemStatus(%q).IsValid() = %v, expected %v", tt.status, got, tt.expected)
		}
	}
}

func TestFlexibleTime_Unmarshal(t *testing.T) {
	t.Run("RFC3339 string", func(t *testing.T) {
		jsonStr := `"2026-11-16T15:04:05Z"`
		var ft FlexibleTime
		err := json.Unmarshal([]byte(jsonStr), &ft)
		if err != nil {
			t.Fatalf("unexpected error unmarshaling RFC3339: %v", err)
		}
		if ft.Time().Year() != 2026 || ft.Time().Month() != 11 || ft.Time().Day() != 16 {
			t.Errorf("unexpected date parsed: %v", ft.Time())
		}
	})

	t.Run("Date-only YYYY-MM-DD", func(t *testing.T) {
		jsonStr := `"2026-11-16"`
		var ft FlexibleTime
		err := json.Unmarshal([]byte(jsonStr), &ft)
		if err != nil {
			t.Fatalf("unexpected error unmarshaling YYYY-MM-DD: %v", err)
		}
		if ft.Time().Year() != 2026 || ft.Time().Month() != 11 || ft.Time().Day() != 16 {
			t.Errorf("unexpected date parsed: %v", ft.Time())
		}
	})

	t.Run("Empty string", func(t *testing.T) {
		var ft FlexibleTime
		err := ft.UnmarshalText([]byte(""))
		if err != nil {
			t.Fatalf("unexpected error on empty string: %v", err)
		}
		if !ft.Time().IsZero() {
			t.Errorf("expected zero time, got %v", ft.Time())
		}
	})

	t.Run("Invalid date string", func(t *testing.T) {
		var ft FlexibleTime
		err := json.Unmarshal([]byte(`"not-a-date"`), &ft)
		if err == nil {
			t.Error("expected error parsing invalid date string, got nil")
		}
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		parsed, _ := time.Parse(time.RFC3339, "2026-11-16T10:00:00Z")
		ft := FlexibleTime(parsed)
		data, err := json.Marshal(ft)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}
		if string(data) != `"2026-11-16T10:00:00Z"` {
			t.Errorf("expected marshaled string %q, got %q", `"2026-11-16T10:00:00Z"`, string(data))
		}

		var zeroFt FlexibleTime
		zeroData, err := json.Marshal(zeroFt)
		if err != nil {
			t.Fatalf("unexpected zero marshal error: %v", err)
		}
		if string(zeroData) != "null" {
			t.Errorf("expected null for zero time, got %s", string(zeroData))
		}
	})
}

func TestCreateItemRequest_Validate(t *testing.T) {
	validDate := time.Now().Add(-1 * time.Hour)
	validLat := 37.7749
	validLon := -122.4194

	tests := []struct {
		name    string
		req     CreateItemRequest
		wantErr bool
	}{
		{
			name: "Valid lost item request",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Blue Backpack",
				Description:     "Blue North Face backpack with laptop inside",
				CategoryID:      4,
				LocationName:    "Central Park Gate 3",
				Latitude:        &validLat,
				Longitude:       &validLon,
				DateLostOrFound: validDate,
				ImageURLs:       []string{"/uploads/bag.jpg"},
			},
			wantErr: false,
		},
		{
			name: "Invalid type",
			req: CreateItemRequest{
				Type:            ItemType("STOLEN"),
				Title:           "Lost Backpack",
				Description:     "A blue backpack",
				CategoryID:      4,
				LocationName:    "Park",
				DateLostOrFound: validDate,
			},
			wantErr: true,
		},
		{
			name: "Title too short",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Hi",
				Description:     "Description is long enough here",
				CategoryID:      1,
				LocationName:    "Park",
				DateLostOrFound: validDate,
			},
			wantErr: true,
		},
		{
			name: "Description too short",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Keys",
				Description:     "Key",
				CategoryID:      3,
				LocationName:    "Park",
				DateLostOrFound: validDate,
			},
			wantErr: true,
		},
		{
			name: "Invalid Category ID",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Keys",
				Description:     "Bunch of brass keys on keyring",
				CategoryID:      0,
				LocationName:    "Park",
				DateLostOrFound: validDate,
			},
			wantErr: true,
		},
		{
			name: "Location name too short",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Keys",
				Description:     "Bunch of brass keys on keyring",
				CategoryID:      3,
				LocationName:    "A",
				DateLostOrFound: validDate,
			},
			wantErr: true,
		},
		{
			name: "Latitude out of range",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Keys",
				Description:     "Bunch of brass keys on keyring",
				CategoryID:      3,
				LocationName:    "Library",
				Latitude:        func() *float64 { l := 95.0; return &l }(),
				DateLostOrFound: validDate,
			},
			wantErr: true,
		},
		{
			name: "Longitude out of range",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Keys",
				Description:     "Bunch of brass keys on keyring",
				CategoryID:      3,
				LocationName:    "Library",
				Longitude:       func() *float64 { l := -195.0; return &l }(),
				DateLostOrFound: validDate,
			},
			wantErr: true,
		},
		{
			name: "Date in distant future",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Keys",
				Description:     "Bunch of brass keys on keyring",
				CategoryID:      3,
				LocationName:    "Library",
				DateLostOrFound: time.Now().Add(48 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "Too many images (> 5)",
			req: CreateItemRequest{
				Type:            ItemTypeLost,
				Title:           "Lost Keys",
				Description:     "Bunch of brass keys on keyring",
				CategoryID:      3,
				LocationName:    "Library",
				DateLostOrFound: validDate,
				ImageURLs:       []string{"1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg", "6.jpg"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateItemRequest.Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateItemStatusRequest_Validate(t *testing.T) {
	valid := UpdateItemStatusRequest{Status: ItemStatusClaimed}
	if err := valid.Validate(); err != nil {
		t.Errorf("expected valid status, got error: %v", err)
	}

	invalid := UpdateItemStatusRequest{Status: ItemStatus("UNKNOWN")}
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestItem_ToResponse(t *testing.T) {
	t.Run("Nil Item", func(t *testing.T) {
		var item *Item
		if res := item.ToResponse(); res != nil {
			t.Errorf("expected nil response for nil item, got %v", res)
		}
	})

	t.Run("Valid Item with nil image slice converts to empty slice", func(t *testing.T) {
		id := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		item := &Item{
			ID:              id,
			UserID:          userID,
			Type:            ItemTypeLost,
			Title:           "Lost Phone",
			Description:     "Black iPhone",
			CategoryID:      1,
			LocationName:    "Downtown",
			DateLostOrFound: now,
			Status:          ItemStatusActive,
			ImageURLs:       nil,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		res := item.ToResponse()
		if res == nil {
			t.Fatal("expected non-nil response")
		}
		if res.ID != id || res.UserID != userID || res.Title != "Lost Phone" {
			t.Errorf("field mismatch in ToResponse: %v", res)
		}
		if res.ImageURLs == nil {
			t.Error("expected non-nil empty slice for ImageURLs in response")
		}
	})
}

func TestToItemResponseList(t *testing.T) {
	t.Run("Empty slice returns empty slice", func(t *testing.T) {
		res := ToItemResponseList(nil)
		if res == nil || len(res) != 0 {
			t.Errorf("expected empty non-nil slice, got %v", res)
		}
	})

	t.Run("Multiple items mapped correctly", func(t *testing.T) {
		items := []Item{
			{ID: uuid.New(), Title: "Item 1", Type: ItemTypeLost},
			{ID: uuid.New(), Title: "Item 2", Type: ItemTypeFound},
		}

		res := ToItemResponseList(items)
		if len(res) != 2 {
			t.Fatalf("expected 2 responses, got %d", len(res))
		}
		if res[0].Title != "Item 1" || res[1].Title != "Item 2" {
			t.Errorf("responses mismatch: %v", res)
		}
	})
}
