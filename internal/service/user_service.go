package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/resend/resend-go/v3"
	"github.com/supabase-community/gotrue-go/types"
)

type UserService struct {
	userRepo models.UserInterface
	resend   *resend.Client
	// profileRepo models.ProfileInterface
	// We remove verificationClient as Supabase handles verification
}

// NewUserService - removing verificationClient from constructor
func NewUserService(userRepo models.UserInterface, resend *resend.Client) *UserService {
	return &UserService{
		userRepo: userRepo,
		resend:   resend,
		// profileRepo: profileRepo,
	}
}

// CreateUser registers a user using Supabase Auth
func (u *UserService) CreateUser(ctx context.Context, email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, constants.ErrEmptyFields
	}

	// we pass an empty string for the accessToken , so that we can use the service role
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email, "")
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, constants.ErrUserAlreadyExists
	}
	// Call Supabase SignUp
	sbUser, err := u.userRepo.SignUp(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// Return our local User model (shell)
	return &models.User{
		ID:        sbUser.ID,
		Email:     sbUser.Email,
		CreatedAt: sbUser.CreatedAt,
	}, nil
}

// VerifyUser - DEPRECATED/REMOVED for Supabase flow (handled by Supabase link)
func (u *UserService) VerifyUser(ctx context.Context, token, email string) error {
	return errors.New("verification is handled by supabase")
}

// AuthenticateUser signs in via Supabase and returns the Token Response
func (u *UserService) AuthenticateUser(ctx context.Context, email, password string) (*types.TokenResponse, error) {
	if email == "" || password == "" {
		return nil, constants.ErrEmptyFields
	}

	return u.userRepo.SignIn(ctx, email, password)
}

func (u *UserService) GetUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return u.userRepo.GetUserByID(ctx, id)
}
