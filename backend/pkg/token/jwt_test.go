package token

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "super-secret-findit-key-for-testing-123456"

func TestNewJWTManager(t *testing.T) {
	t.Run("success with custom duration", func(t *testing.T) {
		mgr, err := NewJWTManager(testSecret, 2*time.Hour)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if mgr == nil {
			t.Fatal("expected manager to not be nil")
		}
		if mgr.duration != 2*time.Hour {
			t.Errorf("expected duration to be 2h, got %v", mgr.duration)
		}
	})

	t.Run("default duration when 0 provided", func(t *testing.T) {
		mgr, err := NewJWTManager(testSecret, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if mgr.duration != DefaultTokenDuration {
			t.Errorf("expected default duration %v, got %v", DefaultTokenDuration, mgr.duration)
		}
	})

	t.Run("error when secret is empty", func(t *testing.T) {
		mgr, err := NewJWTManager("", time.Hour)
		if !errors.Is(err, ErrEmptySecret) {
			t.Fatalf("expected ErrEmptySecret, got %v", err)
		}
		if mgr != nil {
			t.Fatal("expected manager to be nil on error")
		}
	})
}

func TestJWTManager_GenerateAndValidateToken(t *testing.T) {
	mgr, err := NewJWTManager(testSecret, time.Hour)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	userID := uuid.New()
	email := "alice@example.com"

	tokenStr, err := mgr.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}

	claims, err := mgr.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %v, got %v", email, claims.Email)
	}
	if claims.Subject != userID.String() {
		t.Errorf("expected Subject %v, got %v", userID.String(), claims.Subject)
	}
	if claims.Issuer != DefaultIssuer {
		t.Errorf("expected Issuer %v, got %v", DefaultIssuer, claims.Issuer)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		t.Errorf("expected future expiration time, got %v", claims.ExpiresAt)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	// Create manager with negative duration to immediately expire tokens
	mgr, err := NewJWTManager(testSecret, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	userID := uuid.New()
	tokenStr, err := mgr.GenerateToken(userID, "expired@example.com")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := mgr.ValidateToken(tokenStr)
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
	if claims != nil {
		t.Errorf("expected nil claims on expired token, got %v", claims)
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	mgr1, _ := NewJWTManager("secret-key-1-alpha", time.Hour)
	mgr2, _ := NewJWTManager("secret-key-2-beta", time.Hour)

	userID := uuid.New()
	tokenStr, err := mgr1.GenerateToken(userID, "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := mgr2.ValidateToken(tokenStr)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken when signature does not match, got %v", err)
	}
	if claims != nil {
		t.Errorf("expected nil claims on invalid signature, got %v", claims)
	}
}

func TestJWTManager_NilUserID(t *testing.T) {
	mgr, _ := NewJWTManager(testSecret, time.Hour)

	tokenStr, err := mgr.GenerateToken(uuid.Nil, "test@example.com")
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
	if tokenStr != "" {
		t.Errorf("expected empty token string on error, got %q", tokenStr)
	}
}

func TestJWTManager_MalformedAndTamperedTokens(t *testing.T) {
	mgr, _ := NewJWTManager(testSecret, time.Hour)

	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"random string", "not-a-valid-jwt-token"},
		{"invalid base64 parts", "header.payload.signature"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := mgr.ValidateToken(tc.token)
			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
			if claims != nil {
				t.Fatalf("expected nil claims, got %v", claims)
			}
		})
	}

	t.Run("tampered token payload", func(t *testing.T) {
		tokenStr, err := mgr.GenerateToken(uuid.New(), "tamper@example.com")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		parts := strings.Split(tokenStr, ".")
		if len(parts) != 3 {
			t.Fatalf("expected 3 parts in JWT, got %d", len(parts))
		}

		// Tamper with the signature
		tamperedSig := parts[2] + "corrupted"
		tamperedToken := strings.Join([]string{parts[0], parts[1], tamperedSig}, ".")

		claims, err := mgr.ValidateToken(tamperedToken)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for tampered token, got %v", err)
		}
		if claims != nil {
			t.Fatalf("expected nil claims, got %v", claims)
		}
	})
}

func TestJWTManager_AlgorithmNoneAttack(t *testing.T) {
	mgr, _ := NewJWTManager(testSecret, time.Hour)

	// Craft an unsecure token using "none" algorithm
	claims := &Claims{
		UserID: uuid.New(),
		Email:  "attacker@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	unsignedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to create unsigned token: %v", err)
	}

	// Validation must reject alg=none
	parsedClaims, err := mgr.ValidateToken(unsignedToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for unsigned alg=none attack, got %v", err)
	}
	if parsedClaims != nil {
		t.Fatalf("expected nil claims, got %v", parsedClaims)
	}
}

func TestStandaloneFunctions(t *testing.T) {
	userID := uuid.New()
	email := "standalone@example.com"

	t.Run("standalone success", func(t *testing.T) {
		tokenStr, err := GenerateToken(userID, email, testSecret, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error from GenerateToken: %v", err)
		}

		claims, err := ValidateToken(tokenStr, testSecret)
		if err != nil {
			t.Fatalf("unexpected error from ValidateToken: %v", err)
		}

		if claims.UserID != userID || claims.Email != email {
			t.Errorf("claims mismatch: got user %v, email %v", claims.UserID, claims.Email)
		}
	})

	t.Run("empty secret in GenerateToken", func(t *testing.T) {
		_, err := GenerateToken(userID, email, "", time.Hour)
		if !errors.Is(err, ErrEmptySecret) {
			t.Fatalf("expected ErrEmptySecret, got %v", err)
		}
	})

	t.Run("empty secret in ValidateToken", func(t *testing.T) {
		_, err := ValidateToken("some.token.str", "")
		if !errors.Is(err, ErrEmptySecret) {
			t.Fatalf("expected ErrEmptySecret, got %v", err)
		}
	})
}
