package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/jwt"
	"github.com/joshua-takyi/auction/internal/middleware"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
)

// GetUsersHandler handles fetching a paginated list of users (example)
func GetUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}
		utils.StatusOK(c, "User found", user)
	}
}

func GetProfileHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exist := c.Get(middleware.AuthorizationPayloadKey)
		accessToken, _ := c.Cookie("access_token")
		if !exist {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}

		claims, ok := user.(*jwt.UserAuth)
		if !ok {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}

		parsedUserId, err := uuid.Parse(claims.ID)
		if err != nil {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}

		u, err := s.GetUserById(c.Request.Context(), parsedUserId, accessToken)
		if err != nil {
			switch {
			case errors.Is(err, constants.ErrUserNotFound):
				utils.NotFound(c, "User not found", "User not found")
			default:
				utils.Error(c, "Failed to get user", err)
			}
			return
		}
		if u.ID != parsedUserId {
			utils.Forbidden(c, "unauthorized access", "You are not authorized to access this resource")
			return
		}
		utils.StatusOK(c, "User found", u)
	}
}
