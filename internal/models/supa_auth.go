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
	if su.url == "" || su.anonKey == "" {
		return su.supabase, nil
	}

	options := &supabase.ClientOptions{
		Headers: map[string]string{
			"Authorization": "Bearer " + accessToken,
		},
	}

	return supabase.NewClient(su.url, su.anonKey, options)
}

func (sr *SupabaseRepo) GetAuthenticatedUser(ctx context.Context, accessToken string) *supabase.Client {
	if accessToken == "" {
		if authClient, err := sr.GetAuthenticatedClient(accessToken); err == nil && authClient != nil {
			return authClient
		}
	}
	if sr.serviceClient != nil {
		return sr.serviceClient
	}
	return nil
}

func (sr *SupabaseRepo) SignUp(ctx context.Context, email, password string) (*types.User, error) {

	_, count, err := sr.supabase.From(string(constants.ProfileTable)).Select("email", "", false).Eq("email", email).Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to check if the user email exist %w", err)
	}
	if count > 0 {
		return nil, constants.ErrUserAlreadyExists
	}

	resp, err := sr.supabase.Auth.Signup(types.SignupRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign up %w", err)
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
	var client *supabase.Client
	var err error

	if accessToken == "" {
		if sr.serviceClient == nil {
			return fmt.Errorf("service client not initialized %w", err)
		}
		client = sr.serviceClient
	} else {
		client, err = sr.GetAuthenticatedClient(accessToken)
		if err != nil {
			return fmt.Errorf("failed to get authenticated client: %w", err)
		}
	}
	return client.Auth.Logout()
}

func (sr *SupabaseRepo) GetUserByID(ctx context.Context, id uuid.UUID, accessToken string) (*User, error) {
	// fmt.Printf("[SupabaseRepo] GetUserByID called for ID: %s, Token Length: %d\n", id.String(), len(accessToken))
	var client *supabase.Client
	var err error

	if accessToken == "" {
		if sr.serviceClient == nil {
			return nil, fmt.Errorf("service client not initialized %w", err)
		}
		client = sr.serviceClient
	} else {
		client, err = sr.GetAuthenticatedClient(accessToken)
		if err != nil {
			return nil, err
		}
	}

	var results []map[string]any
	// Debug log before execution
	// fmt.Println("[SupabaseRepo] Executing query for profile...")
	_, err = client.From(string(constants.ProfileTable)).Select("email, id, username, full_name, role,avatar_url, created_at, updated_at", "exact", false).Eq("id", id.String()).ExecuteTo(&results)
	if err != nil {
		// fmt.Printf("[SupabaseRepo] Error querying profile: %v\n", err)
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	// fmt.Printf("[SupabaseRepo] Query success. Results found: %d\n", len(results))

	if len(results) == 0 {
		// fmt.Println("[SupabaseRepo] No profile found for this ID.")
		return nil, constants.ErrUserNotFound
	}

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

func (sr *SupabaseRepo) GetUserByEmail(ctx context.Context, email string, accessToken string) (*User, error) {
	var client *supabase.Client
	var err error

	if accessToken == "" {
		if sr.serviceClient == nil {
			return nil, fmt.Errorf("service client not initialized %w", err)
		}
		client = sr.serviceClient
	} else {
		client, err = sr.GetAuthenticatedClient(accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get authenticated client: %w", err)
		}
	}

	var results []map[string]any
	_, err = client.From(string(constants.ProfileTable)).Select("email", "exact", false).Eq("email", email).ExecuteTo(&results)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("user not found %w", err)
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

func (sr *SupabaseRepo) RefreshToken(ctx context.Context, refreshToken string) (*types.TokenResponse, error) {
	return sr.supabase.Auth.RefreshToken(refreshToken)
}
