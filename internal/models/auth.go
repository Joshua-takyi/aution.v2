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
	AvatarURL string    `db:"avatar_url" json:"avatar_url,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type Profile struct {
	FirstName     string    `db:"first_name" json:"first_name" validate:"required"`
	LastName      string    `db:"last_name" json:"last_name" validate:"required"`
	UserName      string    `db:"username" json:"username,omitempty"`
	Phone         string    `db:"phone" json:"phone,omitempty"`
	PostalCode    string    `db:"postal_code" json:"postal_code" validate:"required"`
	Region        string    `db:"region" json:"region" validate:"required"`
	City          string    `db:"city" json:"city"  validate:"required"`
	StreetAddress string    `db:"street_address" json:"street_address"  validate:"required"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at,omitempty"`
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

func (u *User) IsOwner(target uuid.UUID) bool {
	return u.ID == target
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
	CreateProfileData(ctx context.Context, profile Profile, userID uuid.UUID, accessToken string) (*Profile, error)
}
