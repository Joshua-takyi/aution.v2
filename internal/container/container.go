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
	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

// Container holds all application dependencies
type Container struct {
	Logger                *slog.Logger
	Config                *config.Config
	IsProduction          bool
	Cloudinary            *cloudinary.Cloudinary
	ResendClient          *resend.Client
	SupabaseStorageClient *storage_go.Client
	SupabaseClient        *supabase.Client

	UserService    *service.UserService
	ProductService *service.ProductService
	AuctionService *service.AuctionService
	JWTManager     *jwt.JWTManager
}

// NewContainer creates a new dependency injection container
func NewContainer(
	logger *slog.Logger,
	cfg *config.Config,
	cloudinary *cloudinary.Cloudinary,
	resendClient *resend.Client,
	supabaseStorageClient *storage_go.Client,
	supabaseClient *supabase.Client,
	isProduction bool,
) (*Container, error) {
	supaRepo := models.NewSupabaseRepo(supabaseClient, supabaseStorageClient, cfg.SupbaseUrl, cfg.SupabaseAnonKey, cfg.SupabaseServiceKey)

	userService := service.NewUserService(supaRepo, resendClient)
	productService := service.NewProductService(supaRepo, cloudinary)
	auctionService := service.NewAuctionService(supaRepo, supaRepo)
	accessDuration, err := time.ParseDuration(cfg.JWTAccessExpiration)

	//TODO: remove this
	if err != nil {
		logger.Warn("Invalid JWT access expiration, using default 15m", "error", err)
		accessDuration = 15 * time.Minute
	}

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.SupabaseJWTSecret, accessDuration)

	return &Container{
		Logger:                logger,
		Config:                cfg,
		IsProduction:          isProduction,
		Cloudinary:            cloudinary,
		ResendClient:          resendClient,
		SupabaseStorageClient: supabaseStorageClient,
		SupabaseClient:        supabaseClient,
		UserService:           userService,
		ProductService:        productService,
		AuctionService:        auctionService,
		JWTManager:            jwtManager,
	}, nil
}
