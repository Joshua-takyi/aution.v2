package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
)

// GetUsersHandler handles fetching a paginated list of users (example)
func GetUsersHandler(s *service.UserService, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.Request.Context() // ctx would be used when calling actual service methods

		// Parse pagination parameters from query string
		var paginationParams utils.PaginationParams
		if err := c.ShouldBindQuery(&paginationParams); err != nil {
			// Use default pagination if parsing fails
			paginationParams = utils.DefaultPaginationParams()
		}

		// Validate and sanitize pagination params
		paginationParams.Validate()

		logger.Info("Fetching users list", map[string]interface{}{
			"page":      paginationParams.Page,
			"page_size": paginationParams.PageSize,
		})

		// Example: Fetch users from service with pagination
		// users, totalCount, err := s.GetUsers(ctx, paginationParams.GetOffset(), paginationParams.GetLimit())
		// For now, using mock data to show the pattern
		users := []map[string]interface{}{
			{"id": "1", "email": "user1@example.com"},
			{"id": "2", "email": "user2@example.com"},
		}
		var totalCount int64 = 50 // This would come from your database count query

		// Create pagination metadata
		paginationMeta := utils.NewPaginationMeta(
			paginationParams.Page,
			paginationParams.PageSize,
			totalCount,
		)

		logger.Info("Users fetched successfully", map[string]interface{}{
			"count":        len(users),
			"total_pages":  paginationMeta.TotalPages,
			"current_page": paginationMeta.CurrentPage,
		})

		// Send paginated response
		utils.PaginatedOK(c, constants.MsgOperationSuccess, users, paginationMeta)
	}
}
