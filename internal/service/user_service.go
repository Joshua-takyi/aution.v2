package service

import (
	"context"

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
}

// NewUserService - removing verificationClient from constructor
func NewUserService(userRepo models.UserInterface, resend *resend.Client) *UserService {
	return &UserService{
		userRepo: userRepo,
		resend:   resend,
	}
}

func (u *UserService) CreateUser(ctx context.Context, email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, constants.ErrEmptyFields
	}

	existingUser, err := u.userRepo.GetUserByEmail(ctx, email, "")
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, constants.ErrUserAlreadyExists
	}
	sbUser, err := u.userRepo.SignUp(ctx, email, password)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        sbUser.ID,
		Email:     sbUser.Email,
		CreatedAt: sbUser.CreatedAt,
	}, nil
}

func (u *UserService) AuthenticateUser(ctx context.Context, email, password string) (*types.TokenResponse, error) {
	if email == "" || password == "" {
		return nil, constants.ErrEmptyFields
	}

	return u.userRepo.SignIn(ctx, email, password)
}

func (u *UserService) SignOut(ctx context.Context, accessToken string) error {
	return u.userRepo.SignOut(ctx, accessToken)
}
func (u *UserService) GetUserById(ctx context.Context, id uuid.UUID, accessToken string) (*models.User, error) {
	return u.userRepo.GetUserByID(ctx, id, accessToken)
}

func (u *UserService) GetUserByEmail(ctx context.Context, email string, accessToken string) (*models.User, error) {
	return u.userRepo.GetUserByEmail(ctx, email, accessToken)
}

func (u *UserService) RefreshToken(ctx context.Context, refreshToken string) (*types.TokenResponse, error) {
	return u.userRepo.RefreshToken(ctx, refreshToken)
}
