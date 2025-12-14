package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/joho/godotenv"
	config "github.com/joshua-takyi/auction/internal/configs"
	"github.com/joshua-takyi/auction/internal/connection"
	"github.com/joshua-takyi/auction/internal/container"
	"github.com/joshua-takyi/auction/internal/helpers"
	"github.com/joshua-takyi/auction/internal/routes"
)

func main() {
	loadLocalEnv()
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		log.Fatal(err)
	}
	supa, err := connection.ConnectSupabase(cfg.SupbaseUrl, cfg.SupabaseAnonKey)
	if err != nil {
		logger.Error("failed to connect to supabase servers", "error", err)
		os.Exit(1)
	}

	// Connect to MongoDB
	mongoClient, err := connection.Connect(cfg.MongoDBURI, cfg.MongoDBPassword)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		log.Fatal(err)
	}

	defer func() {
		if err := connection.Disconnect(); err != nil {
			logger.Error("Failed to disconnect from MongoDB", "error", err)
		}
	}()

	// Initialize Cloudinary (optional - can be nil if not configured)
	var cloudinaryClient *cloudinary.Cloudinary
	cloudinaryURL := cfg.GetCloudinaryURL()
	if cloudinaryURL != "" {
		cloudinaryClient, err = cloudinary.NewFromURL(cloudinaryURL)
		if err != nil {
			logger.Warn("Failed to initialize Cloudinary", "error", err)
		}
	}

	// Initialize Resend
	resendClient, err := connection.ResendConnect(cfg.ResendAPIKey)
	if err != nil {
		logger.Warn("Failed to initialize Resend", "error", err)
	}

	// Initialize dependency injection container
	appContainer, err := container.NewContainer(logger, cfg, cloudinaryClient, mongoClient, resendClient, supa, cfg.IsProduction())
	if err != nil {
		logger.Error("Failed to create container", "error", err)
		log.Fatal(err)
	}

	// Start background workers
	helpers.StartCleanupWorker(context.Background(), logger, appContainer.VerificationRepo.DeleteExpiredVerificationTokens)

	// Setup routes with injected dependencies
	router := routes.SetupRoutes(appContainer)

	// Configure HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("Server starting", "port", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server failed to start", "error", err)
		log.Fatal(err)
	}
}

func loadLocalEnv() {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
	if env == "production" {
		return
	}

	if _, err := os.Stat(".env.local"); err == nil {
		if err := godotenv.Load(".env.local"); err != nil {
			slog.Warn("Failed to load .env.local", "error", err)
		}
	}
}
