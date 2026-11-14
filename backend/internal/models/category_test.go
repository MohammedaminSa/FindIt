package models

import (
	"testing"
	"time"
)

func TestCategory_ToResponse(t *testing.T) {
	now := time.Now().UTC()
	cat := &Category{
		ID:        1,
		Name:      "Electronics",
		Icon:      "devices",
		CreatedAt: now,
	}

	res := cat.ToResponse()
	if res == nil {
		t.Fatal("expected non-nil response")
	}

	if res.ID != cat.ID {
		t.Errorf("expected ID %d, got %d", cat.ID, res.ID)
	}
	if res.Name != cat.Name {
		t.Errorf("expected Name %s, got %s", cat.Name, res.Name)
	}
	if res.Icon != cat.Icon {
		t.Errorf("expected Icon %s, got %s", cat.Icon, res.Icon)
	}
	if !res.CreatedAt.Equal(cat.CreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", cat.CreatedAt, res.CreatedAt)
	}
}

func TestCategory_ToResponse_NilReceiver(t *testing.T) {
	var cat *Category
	res := cat.ToResponse()
	if res != nil {
		t.Errorf("expected nil response for nil receiver, got %v", res)
	}
}

func TestToCategoryResponseList(t *testing.T) {
	t.Run("nil slice returns non-nil empty slice", func(t *testing.T) {
		res := ToCategoryResponseList(nil)
		if res == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(res) != 0 {
			t.Fatalf("expected length 0, got %d", len(res))
		}
	})

	t.Run("empty slice returns empty slice", func(t *testing.T) {
		res := ToCategoryResponseList([]Category{})
		if res == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(res) != 0 {
			t.Fatalf("expected length 0, got %d", len(res))
		}
	})

	t.Run("converts multiple categories", func(t *testing.T) {
		now := time.Now().UTC()
		cats := []Category{
			{ID: 1, Name: "Electronics", Icon: "devices", CreatedAt: now},
			{ID: 2, Name: "Keys", Icon: "vpn_key", CreatedAt: now},
		}

		res := ToCategoryResponseList(cats)
		if len(res) != 2 {
			t.Fatalf("expected 2 items, got %d", len(res))
		}

		if res[0].ID != 1 || res[0].Name != "Electronics" || res[0].Icon != "devices" {
			t.Errorf("unexpected category at index 0: %+v", res[0])
		}
		if res[1].ID != 2 || res[1].Name != "Keys" || res[1].Icon != "vpn_key" {
			t.Errorf("unexpected category at index 1: %+v", res[1])
		}
	})
}
