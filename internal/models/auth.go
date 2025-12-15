package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"password"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// UserInterface defines the methods for interacting with users via Supabase
type UserInterface interface {
	// Return Supabase Session or User
	SignUp(ctx context.Context, email, password string) (*types.User, error)
	SignIn(ctx context.Context, email, password string) (*types.TokenResponse, error)
	SignOut(ctx context.Context, accessToken string) error
	// Helper to get user from DB if we sync them, or just from context
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string, accessToken string) (*User, error)
}
