package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	config "github.com/joshua-takyi/auction/internal/configs"
	"github.com/joshua-takyi/auction/internal/container"
	"github.com/joshua-takyi/auction/internal/handlers"
	"github.com/joshua-takyi/auction/internal/middleware"
	"github.com/joshua-takyi/auction/internal/utils"
)

func SetupRoutes(c *container.Container, cfg *config.Config) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create a new Gin router without default middleware
	r := gin.New()

	// Create logger wrapper from container's slog.Logger
	logger := utils.NewLogger(c.Logger)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.RedirectFixedPath = false
	// Apply custom middleware with injected logger
	r.Use(middleware.ErrorHandler(logger))
	r.Use(middleware.RequestLogger(logger))
	// r.Use(middleware.CORS())
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
			users.POST("", handlers.CreateUserHandler(c.UserService, logger))
			users.POST("/login", handlers.AuthenticateUserHandler(c.UserService, logger, c.IsProduction))
			// users.GET("/", handlers.GetUsersHandler(c.UserService, logger)) // Commenting out if not verified present
			// users.GET("/verify", handlers.VerifyUserHandler(c.UserService, logger)) // Removed
		}

		// Product routes
		products := v1.Group("/products")
		{
			products.POST("", handlers.CreateProductHandler(c.ProductService, logger))
		}
	}

	return r
}
