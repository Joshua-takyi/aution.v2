package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/joshua-takyi/auction/internal/utils"
	"github.com/joshua-takyi/auction/internal/websockets"
)

func CreateWSTicketHandler(m *websockets.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "User not found in context", "Please login")
			return
		}

		user, ok := userVal.(*models.User)
		if !ok {
			utils.InternalServerError(c, "User context invalid type", "Please try again later")
			return
		}

		ticket := m.CreateTicket(user.ID.String())

		utils.StatusOK(c, "Ticket generated successfully", gin.H{
			"ticket": ticket,
		})
	}
}
