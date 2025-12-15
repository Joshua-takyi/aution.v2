package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
)

func (su *SupabaseRepo) GetAuthenticatedClient(accessToken string) (*supabase.Client, error) {
	if su.supabase == nil {
		return nil, errors.New("supabase is nil")
	}

	if su.supabase.Auth == nil {
		return nil, errors.New("supabase auth is nil")
	}

	options := &supabase.ClientOptions{
		Headers: map[string]string{
			"Authorization": "Bearer " + accessToken,
		},
	}

	return supabase.NewClient(su.url, su.anonKey, options)
}

func (sr *SupabaseRepo) GetAuthenticatedUser(ctx context.Context, accessToken string) (*types.User, error) {
	if accessToken == "" {
		return nil, errors.New("access token is empty")
	}

	// Use the auth client to get the user using the access token
	resp, err := sr.supabase.Auth.WithToken(accessToken).GetUser()
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}

func (sr *SupabaseRepo) SignUp(ctx context.Context, email, password string) (*types.User, error) {
	_, count, err := sr.supabase.From(string(constants.ProductTable)).Select("email", "", false).Eq("email", email).Execute()
	if err != nil {
		return nil, errors.New("failed to check if the user email exist")
	}
	if count > 0 {
		return nil, errors.New("user already exists")
	}

	resp, err := sr.supabase.Auth.Signup(types.SignupRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, errors.New("failed to sign up")
	}
	return &resp.User, nil
}

func (sr *SupabaseRepo) SignIn(ctx context.Context, email, password string) (*types.TokenResponse, error) {
	resp, err := sr.supabase.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, errors.New("failed to sign in")
	}
	return resp, nil
}

func (sr *SupabaseRepo) SignOut(ctx context.Context, accessToken string) error {
	return sr.supabase.Auth.Logout()
}

func (sr *SupabaseRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if sr.serviceClient == nil {
		return nil, errors.New("service client not initialized")
	}

	// Use AdminGetUser with the service role client (Admin Request)
	u, err := sr.serviceClient.Auth.AdminGetUser(types.AdminGetUserRequest{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (sr *SupabaseRepo) GetUserByEmail(ctx context.Context, email string, accessToken string) (*User, error) {
	var client *supabase.Client
	var err error

	if accessToken == "" {
		if sr.serviceClient == nil {
			return nil, errors.New("service client not initialized")
		}
		client = sr.serviceClient
	} else {
		client, err = sr.GetAuthenticatedClient(accessToken)
		if err != nil {
			return nil, err
		}
	}

	var results []map[string]interface{}
	_, err = client.From(string(constants.ProfileTable)).Select("email", "exact", false).Eq("email", email).ExecuteTo(&results)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("user not found %w", constants.ErrUserNotFound)
	}

	// Marshal the first result to User struct
	// Note: The User struct in models should match the Profile table structure or we need manual mapping
	jsonBody, err := json.Marshal(results[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user data: %w", err)
	}

	var u User
	if err := json.Unmarshal(jsonBody, &u); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data to struct: %w", err)
	}
	return &u, nil
}
