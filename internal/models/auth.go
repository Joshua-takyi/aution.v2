package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

type User struct {
	ID        uuid.UUID `db:"id" json:"id,omitempty"`
	Email     string    `db:"email" json:"email,omitempty"`
	Role      string    `db:"role" json:"role,omitempty"`
	UserName  string    `db:"username" json:"username,omitempty"`
	FullName  string    `db:"full_name" json:"full_name,omitempty"`
	AvatarURL string    `db:"avatar_url" json:"avatar_url,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

func (u *User) IsAdminOrSeller() bool {
	return u.Role == "admin" || u.Role == "seller"
}

func (u *User) IsAdminOrOwner(targetID uuid.UUID) bool {
	if u.Role == "admin" {
		return true
	}
	return u.ID == targetID
}

// UserInterface defines the methods for interacting with users via Supabase
type UserInterface interface {
	// Return Supabase Session or User
	SignUp(ctx context.Context, email, password string) (*types.User, error)
	SignIn(ctx context.Context, email, password string) (*types.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*types.TokenResponse, error)
	SignOut(ctx context.Context, accessToken string) error
	// Helper to get user from DB if we sync them, or just from context
	GetUserByID(ctx context.Context, id uuid.UUID, accessToken string) (*User, error)
	GetUserByEmail(ctx context.Context, email string, accessToken string) (*User, error)
}
