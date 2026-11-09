package hasher

import (
	"strings"
	"testing"
)

func TestHashPassword_Success(t *testing.T) {
	password := "SecretPass123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if hash == password {
		t.Fatal("hash should not equal plain text password")
	}

	if !CheckPasswordHash(password, hash) {
		t.Error("expected CheckPasswordHash to return true for matching password and hash")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Error("expected CheckPasswordHash to return false for incorrect password")
	}
}

func TestHashPassword_Empty(t *testing.T) {
	_, err := HashPassword("")
	if err != ErrEmptyPassword {
		t.Errorf("expected ErrEmptyPassword, got: %v", err)
	}
}

func TestHashPassword_TooLong(t *testing.T) {
	tooLong := strings.Repeat("a", 73)
	_, err := HashPassword(tooLong)
	if err != ErrPasswordTooLong {
		t.Errorf("expected ErrPasswordTooLong, got: %v", err)
	}
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	if CheckPasswordHash("password", "invalid_not_a_bcrypt_hash") {
		t.Error("expected false for invalid bcrypt hash string")
	}
}
