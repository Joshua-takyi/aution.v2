package middleware

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/jwt"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// AuthorizationPayloadKey is the key to store user auth in context
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for authentication
func AuthMiddleware(jwtManager *jwt.JWTManager, userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			utils.Unauthorized(c, "Access token not found", "Please login")
			c.Abort()
			return
		}
		// Verify CSRF token for mutating requests only
		methods := []string{"POST", "PATCH", "DELETE", "PUT"}
		for _, v := range methods {
			if c.Request.Method == v {
				csrfTokenFromHeader := c.GetHeader("X-CSRF-Token")
				csrfTokenFromCookie, err := c.Cookie("csrf_token")
				if err != nil {
					utils.Unauthorized(c, "CSRF token not found", "Please login")
					c.Abort()
					return
				}

				if csrfTokenFromCookie == "" || csrfTokenFromHeader == "" || csrfTokenFromCookie != csrfTokenFromHeader {
					utils.Unauthorized(c, "Invalid CSRF token", "Please try again")
					c.Abort()
					return
				}
				break // Exit loop after validating for the matching method
			}
		}

		// Verify token with Supabase secret
		userAuth, err := jwtManager.VerifySupabaseToken(token)
		if err != nil {
			// validate refresh token
			switch {
			case errors.Is(err, jwt.ErrExpiredToken):
				utils.Unauthorized(c, "Token has expired", "Please login again")
			default:
				utils.Unauthorized(c, "Invalid token", "Please try again")
			}
			c.Abort()
			return
		}

		parsedUserID, err := uuid.Parse(userAuth.ID)
		if err != nil {
			utils.Unauthorized(c, "Invalid user ID", "Please try again")
			c.Abort()
			return
		}
		user, err := userService.GetUserById(c.Request.Context(), parsedUserID, token)
		if err != nil {
			fmt.Printf("[AuthMiddleware] GetUserById failed for userID %s: %v\n", parsedUserID.String(), err)
			utils.Unauthorized(c, "User not found", "Please try again")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set(AuthorizationPayloadKey, userAuth)
		c.Next()
	}
}
