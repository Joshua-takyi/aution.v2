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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.RedirectFixedPath = false
	// Apply custom middleware with injected logger
	r.Use(middleware.ErrorHandler(logger))
	r.Use(middleware.RequestLogger(logger))
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

		// --- Public Routes ---

		// Auth & User Registration (Public)
		v1.POST("/users", handlers.CreateUserHandler(c.UserService, logger))
		v1.POST("/users/login", handlers.AuthenticateUserHandler(c.UserService, logger, c.IsProduction))
		v1.POST("/auth/refresh", handlers.RefreshToken(c.UserService, c.IsProduction))
		v1.GET("/auctions/:id", handlers.GetAuctionByIdHandler(c.AuctionService))

		// --- Protected Routes (Require Auth) ---

		// Protected grouping
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(c.JWTManager, c.UserService))

		// User Protected Routes
		userRoutes := protected.Group("/users")
		{
			userRoutes.GET("/", handlers.GetUserHandler())
			userRoutes.GET("/profile", handlers.GetProfileHandler(c.UserService))
			userRoutes.POST("/auth/signout", handlers.SignOut(c.UserService, c.IsProduction))
		}

		// Product Protected Routes
		productRoutes := protected.Group("/products")
		{
			productRoutes.POST("", handlers.CreateProductHandler(c.ProductService, logger))
			productRoutes.GET("/:id", handlers.GetProductById(c.ProductService))
			productRoutes.DELETE("/:id", handlers.DeleteProduct(c.ProductService))
		}

		// Auction Protected Routes
		auctionRoutes := protected.Group("/auctions")
		{
			auctionRoutes.POST("/:id", handlers.CreateAuctionHandler(c.AuctionService))
			auctionRoutes.DELETE("/:id", handlers.DeleteAuctionHandler(c.AuctionService))
		}
	}

	return r
}
