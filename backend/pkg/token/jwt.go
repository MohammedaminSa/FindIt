package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrEmptySecret is returned when attempting to create a token manager or generate a token with an empty secret key.
	ErrEmptySecret = errors.New("jwt secret key cannot be empty")
	// ErrInvalidUserID is returned when an empty or nil user ID is provided.
	ErrInvalidUserID = errors.New("invalid or empty user ID")
	// ErrInvalidToken is returned when the token is malformed, has an invalid signature, or cannot be parsed.
	ErrInvalidToken = errors.New("token is invalid")
	// ErrExpiredToken is returned when the token has passed its expiration time.
	ErrExpiredToken = errors.New("token has expired")
)

const (
	// DefaultIssuer is the standard issuer added to FindIt JWT claims.
	DefaultIssuer = "findit-api"
	// DefaultTokenDuration is used when no duration is specified.
	DefaultTokenDuration = 24 * time.Hour
)

// Claims represents the custom claims payload embedded in FindIt JWT tokens.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// TokenManager defines the interface for creating and parsing JWT tokens.
type TokenManager interface {
	GenerateToken(userID uuid.UUID, email string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

// JWTManager implements TokenManager using HMAC-SHA256 (HS256) signing.
type JWTManager struct {
	secretKey string
	duration  time.Duration
	issuer    string
}

// NewJWTManager creates a new JWTManager with the given secret key and token lifetime duration.
func NewJWTManager(secretKey string, duration time.Duration) (*JWTManager, error) {
	if secretKey == "" {
		return nil, ErrEmptySecret
	}
	if duration == 0 {
		duration = DefaultTokenDuration
	}
	return &JWTManager{
		secretKey: secretKey,
		duration:  duration,
		issuer:    DefaultIssuer,
	}, nil
}

// GenerateToken creates and signs a new JWT token containing the user ID and email.
func (m *JWTManager) GenerateToken(userID uuid.UUID, email string) (string, error) {
	if userID == uuid.Nil {
		return "", ErrInvalidUserID
	}

	now := time.Now().UTC()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses and validates a JWT token string, verifying signature, signing method, and expiration.
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Strict check to ensure the signing method is HMAC to prevent algorithm switching attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateToken is a convenience standalone function that creates and signs a JWT token.
func GenerateToken(userID uuid.UUID, email string, secret string, duration time.Duration) (string, error) {
	mgr, err := NewJWTManager(secret, duration)
	if err != nil {
		return "", err
	}
	return mgr.GenerateToken(userID, email)
}

// ValidateToken is a convenience standalone function that parses and validates a JWT token with the given secret.
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	mgr, err := NewJWTManager(secret, DefaultTokenDuration)
	if err != nil {
		return nil, err
	}
	return mgr.ValidateToken(tokenString)
}
