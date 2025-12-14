package container

import (
	"log/slog"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	config "github.com/joshua-takyi/auction/internal/configs"
	"github.com/joshua-takyi/auction/internal/jwt"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/resend/resend-go/v3"
	"github.com/supabase-community/supabase-go"
	"go.mongodb.org/mongo-driver/mongo"
)

// Container holds all application dependencies
type Container struct {
	Logger       *slog.Logger
	Config       *config.Config
	IsProduction bool
	Cloudinary   *cloudinary.Cloudinary
	// Database clients
	ResendClient   *resend.Client
	MongoDBClient  *mongo.Client
	SupabaseClient *supabase.Client
	// Services
	VerificationRepo models.VerificationInterface
	UserService      *service.UserService
	JWTManager       *jwt.JWTManager
}

// NewContainer creates a new dependency injection container
func NewContainer(
	logger *slog.Logger,
	cfg *config.Config,
	cloudinary *cloudinary.Cloudinary,
	mongoDBClient *mongo.Client,
	resendClient *resend.Client,
	supabaseClient *supabase.Client,
	isProduction bool,
) (*Container, error) {
	// Initialize repositories
	mongoRepo, err := models.MongodbNewRepo(mongoDBClient)
	if err != nil {
		return nil, err
	}

	// Initialize services
	// mongoRepo implements both UserInterface and VerificationInterface
	userService := service.NewUserService(mongoRepo, resendClient, mongoRepo, mongoRepo)

	// Parse JWT durations
	accessDuration, err := time.ParseDuration(cfg.JWTAccessExpiration)
	if err != nil {
		logger.Warn("Invalid JWT access expiration, using default 15m", "error", err)
		accessDuration = 15 * time.Minute
	}

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, accessDuration)

	return &Container{
		Logger:           logger,
		Config:           cfg,
		IsProduction:     isProduction,
		Cloudinary:       cloudinary,
		ResendClient:     resendClient,
		MongoDBClient:    mongoDBClient,
		SupabaseClient:   supabaseClient,
		VerificationRepo: mongoRepo,
		UserService:      userService,
		JWTManager:       jwtManager,
	}, nil
}
