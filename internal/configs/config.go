package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                  string
	MongoDBURI            string
	CloudinaryCloudName   string
	CloudinaryAPIKey      string
	CloudinaryAPISecret   string
	MongoDBPassword       string
	Environment           string
	LogLevel              string
	SupbaseUrl            string
	SupabaseServiceKey    string
	SupabaseAnonKey       string
	AllowedOrigins        []string
	FrontendURL           string
	PaystackPublicKey     string
	PaystackSecretKey     string
	PaystackTestPublicKey string
	PaystackTestSecretKey string
	ResendAPIKey          string
	// JWT Configuration
	JWTSecret            string
	JWTAccessExpiration  string // e.g., "15m", "1h"
	JWTRefreshExpiration string // e.g., "7d", "30d"
}

func LoadConfig() (*Config, error) {
	env := strings.ToLower(strings.TrimSpace(getEnvWithDefault("ENVIRONMENT", "development")))
	cfg := &Config{
		Port:                getEnvWithDefault("PORT", "8080"),
		MongoDBURI:          os.Getenv("MONGODB_URI"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
		MongoDBPassword:     os.Getenv("MONGODB_PASSWORD"),
		Environment:         env,
		SupbaseUrl:          os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey:  os.Getenv("SUPABASE_SERVICE_KEY"),
		SupabaseAnonKey:     os.Getenv("SUPABASE_ANON_KEY"),

		LogLevel:              getEnvWithDefault("LOG_LEVEL", "info"),
		ResendAPIKey:          os.Getenv("RESEND_API_KEY"),
		FrontendURL:           getEnvWithDefault("FRONTEND_URL", "http://localhost:3000"),
		PaystackPublicKey:     os.Getenv("PAYSTACK_PUBLIC_KEY"),
		PaystackSecretKey:     os.Getenv("PAYSTACK_SECRET_KEY"),
		PaystackTestPublicKey: os.Getenv("PAYSTACK_TEST_PUBLIC_KEY"),
		PaystackTestSecretKey: os.Getenv("PAYSTACK_TEST_SECRET_KEY"),
		// JWT Configuration
		JWTSecret:            os.Getenv("JWT_SECRET"),
		JWTAccessExpiration:  getEnvWithDefault("JWT_ACCESS_EXPIRATION", "15m"),
		JWTRefreshExpiration: getEnvWithDefault("JWT_REFRESH_EXPIRATION", "7d"),
	}

	allowedOrigins := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	if allowedOrigins == "" {
		if cfg.IsProduction() {
			return nil, fmt.Errorf("ALLOWED_ORIGINS is required in production")
		}
		allowedOrigins = "http://localhost:3000"
	}
	cfg.AllowedOrigins = splitAndTrim(allowedOrigins)

	if cfg.MongoDBURI == "" {
		return nil, fmt.Errorf("MONGODB_URI is required")
	}

	if cfg.SupabaseServiceKey == "" {
		return nil, fmt.Errorf("failed to load the supabase  service role key ")
	}

	if cfg.SupabaseAnonKey == "" {
		return nil, fmt.Errorf("failed to load the supabase token")
	}

	if cfg.SupbaseUrl == "" {
		return nil, fmt.Errorf("failed to load the supabase url")
	}
	if cfg.MongoDBPassword == "" {
		return nil, fmt.Errorf("MONGODB_PASSWORD is required")
	}
	if cfg.CloudinaryCloudName == "" {
		return nil, fmt.Errorf("CLOUDINARY_CLOUD_NAME is required")
	}
	if cfg.CloudinaryAPIKey == "" {
		return nil, fmt.Errorf("CLOUDINARY_API_KEY is required")
	}
	if cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("CLOUDINARY_API_SECRET is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	if cfg.ResendAPIKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY is required")
	}

	if cfg.IsProduction() {
		if cfg.PaystackPublicKey == "" {
			return nil, fmt.Errorf("PAYSTACK_PUBLIC_KEY is required in production")
		}
		if cfg.PaystackSecretKey == "" {
			return nil, fmt.Errorf("PAYSTACK_SECRET_KEY is required in production")
		}
	}

	if !cfg.IsProduction() {
		if cfg.PaystackTestPublicKey == "" || cfg.PaystackTestSecretKey == "" {
			return nil, fmt.Errorf("PAYSTACK_TEST_PUBLIC_KEY and PAYSTACK_TEST_SECRET_KEY are required outside production")
		}
	}

	// if
	return cfg, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetCloudinaryURL builds the Cloudinary URL from config fields
func (c *Config) GetCloudinaryURL() string {
	if c.CloudinaryCloudName == "" || c.CloudinaryAPIKey == "" || c.CloudinaryAPISecret == "" {
		return ""
	}
	return fmt.Sprintf("cloudinary://%s:%s@%s", c.CloudinaryAPIKey, c.CloudinaryAPISecret, c.CloudinaryCloudName)
}

func splitAndTrim(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
