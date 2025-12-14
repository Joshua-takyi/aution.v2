package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/container"
	"github.com/joshua-takyi/auction/internal/handlers"
	"github.com/joshua-takyi/auction/internal/middleware"
	"github.com/joshua-takyi/auction/internal/utils"
)

func SetupRoutes(c *container.Container) *gin.Engine {
	// Create a new Gin router without default middleware
	r := gin.New()

	// Create logger wrapper from container's slog.Logger
	logger := utils.NewLogger(c.Logger)

	// Apply custom middleware with injected logger
	r.Use(middleware.ErrorHandler(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Health check
		v1.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"message": "Welcome to the Auction API!",
				"version": "1.0.0",
			})
		})

		// User routes
		users := v1.Group("/users")
		{
			users.POST("/", handlers.CreateUserHandler(c.UserService, logger))
			users.POST("/login", handlers.AuthenticateUserHandler(c.UserService, logger, c.JWTManager, c.IsProduction))
			users.GET("/", handlers.GetUsersHandler(c.UserService, logger))
			users.GET("/verify", handlers.VerifyUserHandler(c.UserService, logger))
		}
	}

	return r
}
