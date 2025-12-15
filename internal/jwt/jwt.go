package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrTokenNotFound    = errors.New("token not found")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// UserAuth represents the authenticated user information
type UserAuth struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Claims represents JWT claims with user information
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey         string
	supabaseSecretKey string
	tokenExpiration   time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey, supabaseSecretKey string, tokenExpiration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:         secretKey,
		supabaseSecretKey: supabaseSecretKey,
		tokenExpiration:   tokenExpiration,
	}
}

func (m *JWTManager) VerifySupabaseToken(tokenString string) (*UserAuth, error) {
	// Supabase tokens are standard JWTs signed with the project secret
	// We use the supabaseSecretKey to verify signature
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(m.supabaseSecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Supabase user claims:
	// "sub": user_id
	// "email": user_email
	// "role": authenticated (usually)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract standard claims
	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string) // "authenticated" usually

	// You can also check "aud" == "authenticated" if needed.

	return &UserAuth{
		ID:    userID,
		Email: email,
		Role:  role,
	}, nil
}
