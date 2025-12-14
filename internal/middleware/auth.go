package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/jwt"
	"github.com/joshua-takyi/auction/internal/utils"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// AuthorizationPayloadKey is the key to store user auth in context
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for authentication
func AuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			utils.Unauthorized(c, "Authorization header is required", "Please provide a valid token")
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			utils.Unauthorized(c, "Invalid authorization header format", "Format should be: Bearer <token>")
			c.Abort()
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != "bearer" {
			utils.Unauthorized(c, "Unsupported authorization type", "Only Bearer tokens are supported")
			c.Abort()
			return
		}

		// Get the token
		token := fields[1]

		// Verify token
		userAuth, err := jwtManager.VerifyToken(token)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				utils.Unauthorized(c, "Token has expired", "Please login again")
			} else if err == jwt.ErrInvalidSignature {
				utils.Unauthorized(c, "Invalid token signature", "Token has been tampered with")
			} else {
				utils.Unauthorized(c, "Invalid token", err.Error())
			}
			c.Abort()
			return
		}

		// Store user info in context
		c.Set(AuthorizationPayloadKey, userAuth)
		c.Next()
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't abort if no token
func OptionalAuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.Next()
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) >= 2 && strings.ToLower(fields[0]) == "bearer" {
			token := fields[1]
			if userAuth, err := jwtManager.VerifyToken(token); err == nil {
				c.Set(AuthorizationPayloadKey, userAuth)
			}
		}

		c.Next()
	}
}

// GetCurrentUser retrieves the authenticated user from the context
func GetCurrentUser(c *gin.Context) (*jwt.UserAuth, bool) {
	value, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return nil, false
	}

	userAuth, ok := value.(*jwt.UserAuth)
	return userAuth, ok
}

// RequireRole creates middleware that checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAuth, exists := GetCurrentUser(c)
		if !exists {
			utils.Unauthorized(c, "Authentication required", "Please login")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if userAuth.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.Forbidden(c, "Insufficient permissions", "You don't have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}
