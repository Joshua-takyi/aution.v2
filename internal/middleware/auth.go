package middleware

import (
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
		// Verify CSRF token
		csrfTokenFromHeader := c.GetHeader("X-CSRF-Token")
		csrfTokenFromCookie, err := c.Cookie("csrf_token")
		if err != nil {
			utils.Unauthorized(c, "CSRF token not found", "Please login")
			c.Abort()
			return
		}

		// Get authorization header
		methods := []string{"POST", "PATCH", "DELETE", "PUT"}
		for _, v := range methods {
			if c.Request.Method == v {
				if csrfTokenFromCookie == "" || csrfTokenFromHeader == "" || csrfTokenFromCookie != csrfTokenFromHeader {
					utils.Unauthorized(c, "Invalid CSRF token", "Please try again")
					c.Abort()
					return
				}

			}
		}

		// Verify token with Supabase secret
		userAuth, err := jwtManager.VerifySupabaseToken(token)
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

		parsedUserID, err := uuid.Parse(userAuth.ID)
		if err != nil {
			utils.Unauthorized(c, "Invalid user ID", "Please login")
			c.Abort()
			return
		}
		// get user from the ID
		user, er := userService.GetUserById(c, parsedUserID)
		if er != nil {
			utils.Unauthorized(c, "User not found", "Please login")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set(AuthorizationPayloadKey, userAuth)
		c.Next()
	}
}
