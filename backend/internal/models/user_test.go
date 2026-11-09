package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserToResponse(t *testing.T) {
	phone := "+1234567890"
	avatar := "https://example.com/avatar.jpg"
	id := uuid.New()
	now := time.Now()

	u := &User{
		ID:           id,
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "secret_hash",
		Phone:        &phone,
		AvatarURL:    &avatar,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	res := u.ToResponse()
	if res == nil {
		t.Fatal("expected non-nil response")
	}

	if res.ID != id {
		t.Errorf("expected ID %v, got %v", id, res.ID)
	}
	if res.Name != "Alice" {
		t.Errorf("expected Name Alice, got %s", res.Name)
	}
	if res.Email != "alice@example.com" {
		t.Errorf("expected Email alice@example.com, got %s", res.Email)
	}
	if *res.Phone != phone {
		t.Errorf("expected Phone %s, got %s", phone, *res.Phone)
	}

	// Test nil receiver
	var nilUser *User
	if nilUser.ToResponse() != nil {
		t.Error("expected nil response for nil user")
	}

	// Test JSON serialization does not include password hash
	bytes, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}
	jsonStr := string(bytes)
	if strings.Contains(jsonStr, "secret_hash") || strings.Contains(jsonStr, "password_hash") {
		t.Error("user JSON serialization leaked password hash")
	}
}
