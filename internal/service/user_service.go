package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/helpers"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/resend/resend-go/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	userRepo           models.UserInterface
	resend             *resend.Client
	profileRepo        models.ProfileInterface
	verificationClient models.VerificationInterface
}

func NewUserService(userRepo models.UserInterface, resend *resend.Client, profileRepo models.ProfileInterface, verificationClient models.VerificationInterface) *UserService {
	return &UserService{
		userRepo:           userRepo,
		resend:             resend,
		profileRepo:        profileRepo,
		verificationClient: verificationClient,
	}
}

func (u *UserService) CreateUser(ctx context.Context, email, password string) (*models.User, error) {
	// validate form
	if email == "" || password == "" {
		return nil, constants.ErrEmptyFields
	}

	ok := helpers.ValidatePassword(password)
	if !ok {
		return nil, constants.ErrWeakPassword
	}

	// Step 1: Create the user first
	user, err := u.userRepo.CreateUser(ctx, email, password)
	if err != nil {
		return nil, err
	}

	verificationToken := helpers.GenerateVerificationToken()
	hashedToken, err := helpers.HashPassword(verificationToken)
	if err != nil {
		// User is created but we couldn't generate token
		log.Printf("Failed to generate verification token: %v", err)
		return user, nil // Still return user as they're created
	}
	verification := &models.Verification{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		Token:     hashedToken,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	_, err = u.verificationClient.CreateVerification(ctx, verification)
	if err != nil {
		log.Printf("Failed to save verification token: %v", err)
		return user, nil // Still return user
	}

	// Step 4: Send verification email
	verificationLink := fmt.Sprintf("http://localhost:8080/verify?token=%s&email=%s", verificationToken, email)
	err = helpers.SendVerificationEmail(u.resend, email, verificationLink)
	if err != nil {
		log.Printf("Failed to send verification email: %v", err)
	}

	return user, nil
}

func (u *UserService) VerifyUser(ctx context.Context, token, email string) error {
	if token == "" || email == "" {
		return constants.ErrEmptyFields
	}

	// Find user by email
	user, err := u.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Find verification token by UserID
	verification, err := u.verificationClient.FindVerificationByUserID(ctx, user.ID)
	if err != nil {
		return err
	}

	if verification == nil {
		return constants.ErrUserNotFound
	}

	// Verify token hash
	ok := helpers.CheckPasswordHash(token, verification.Token)
	if !ok {
		return constants.ErrInvalidToken
	}

	if time.Now().After(verification.ExpiresAt) {
		return constants.ErrTokenExpired // Or a more appropriate error
	}

	// Step 5: Update user verification status
	if err := u.userRepo.UpdateUserVerificationStatus(ctx, verification.UserID, true); err != nil {
		log.Printf("Failed to update user verification status: %v", err)
		return err
	}

	// Step 6: Create profile
	if err := u.profileRepo.CreateProfile(ctx, email, &models.Profile{}); err != nil {
		log.Printf("Failed to create profile: %v", err)
		return err
	}
	// Step 6: Delete verification record
	if err := u.verificationClient.DeleteVerificationToken(ctx, token); err != nil {
		log.Printf("Failed to delete verification token: %v", err)
		return err
	}

	return nil
}

func (u *UserService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	//if the user isn't verified, return error
	ok, err := u.userRepo.IsVerified(ctx, email)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, constants.ErrUserNotVerified
	}

	user, err := u.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if ok := helpers.CheckPasswordHash(password, user.Password); !ok {
		return nil, constants.ErrInvalidCredentials
	}
	return user, nil
}

func (s *UserService) GetProfileByUserId(ctx context.Context, id primitive.ObjectID) (*models.Profile, error) {
	return s.profileRepo.GetUserProfileByID(ctx, id)
}
