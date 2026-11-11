package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// AuthorizationHeaderKey is the standard HTTP header key for authentication tokens.
	AuthorizationHeaderKey = "Authorization"

	// AuthorizationTypeBearer is the expected prefix for bearer tokens.
	AuthorizationTypeBearer = "bearer"

	// ContextUserIDKey is the key used to store and retrieve the authenticated user's UUID in the Gin context.
	ContextUserIDKey = "userID"

	// ContextUserEmailKey is the key used to store and retrieve the authenticated user's email in the Gin context.
	ContextUserEmailKey = "userEmail"

	// ContextClaimsKey is the key used to store and retrieve the entire token claims in the Gin context.
	ContextClaimsKey = "claims"
)

// Auth returns a Gin middleware that validates JWT Bearer tokens from the Authorization header.
// If valid, it stores the user claims in the Gin context and calls c.Next().
// If invalid or missing, it aborts the request pipeline with HTTP 401 Unauthorized.
func Auth(tokenManager token.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != AuthorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unsupported authorization type",
			})
			return
		}

		tokenString := fields[1]
		claims, err := tokenManager.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, token.ErrExpiredToken) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "token has expired",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or malformed token",
			})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Set(ContextClaimsKey, claims)

		c.Next()
	}
}

// RequireAuth is an alias for Auth to match common naming conventions.
var RequireAuth = Auth

// GetUserID retrieves the authenticated user's UUID from the Gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetUserEmail retrieves the authenticated user's email address from the Gin context.
func GetUserEmail(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextUserEmailKey)
	if !exists {
		return "", false
	}
	email, ok := val.(string)
	return email, ok
}

// GetClaims retrieves the parsed JWT claims from the Gin context.
func GetClaims(c *gin.Context) (*token.Claims, bool) {
	val, exists := c.Get(ContextClaimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*token.Claims)
	return claims, ok
}
