package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/supabase-community/supabase-go"
)

type Product struct {
	ID          uuid.UUID        `db:"id" json:"id"`
	OwnerID     uuid.UUID        `db:"owner_id" json:"owner_id"`
	Title       string           `db:"title" json:"title"`
	Category    string           `db:"category" json:"category"`
	Description string           `db:"description" json:"description"`
	Slug        string           `db:"slug" json:"slug"`
	Brand       string           `db:"brand" json:"brand"`
	Specs       []map[string]any `db:"specs" json:"specs"`
	Images      []string         `db:"images" json:"images"`
	Status      string           `db:"status" json:"status"`
	CreatedAt   time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updated_at"`
}

type ProductInterface interface {
	CreateProduct(ctx context.Context, product *Product, accessToken string, userID uuid.UUID) (*Product, error)
	GetProductById(ctx context.Context, accessToken string, productID, ownerID uuid.UUID) (*Product, error)
	UpdateProduct(ctx context.Context, product map[string]any, accessToken string, productID uuid.UUID) (*Product, error)
	DeleteProduct(ctx context.Context, accessToken string, productID, ownerID uuid.UUID) error
}

func (sr *SupabaseRepo) CreateProduct(ctx context.Context, product *Product, accessToken string, userID uuid.UUID) (*Product, error) {

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
	// fmt.Printf("[DB] Inserting product into table: %s\n", constants.ProductTable)
	// fmt.Printf("[DB] Product data: ID=%s, Title=%s, Images=%d, OwnerID=%s\n",
	// 	product.ID, product.Title, len(product.Images), product.OwnerID)

	res, _, err := client.From(string(constants.ProductTable)).Insert(product, false, "", "", "exact").Execute()

	if err != nil {
		// fmt.Printf("[DB] ERROR: Failed to insert product: %v\n", err)
		return nil, fmt.Errorf("failed to insert product into table: %w", err)
	}
	// fmt.Printf("[DB] Product inserted successfully. Response length: %d bytes\n", len(res))

	var p []Product
	if err := json.Unmarshal(res, &p); err != nil {
		// fmt.Printf("[DB] ERROR: Failed to unmarshal response: %v\n", err)
		// fmt.Printf("[DB] Response data: %s\n", string(res))
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}
	// fmt.Printf("[DB] Product unmarshaled successfully: ID=%s\n", p[0].ID)

	return &p[0], nil
}

func (sr *SupabaseRepo) UpdateProduct(ctx context.Context, product map[string]any, accessToken string, productID uuid.UUID) (*Product, error) {
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

	allowedFields := []string{"title", "category", "description", "brand", "specs", "images"}

	for key := range product {
		if !contains(allowedFields, key) {
			return nil, fmt.Errorf("field %s is not allowed", key)
		}
	}

	byteData, _, err := client.From(string(constants.ProductTable)).Update(product, "", "exact").Eq("id", productID.String()).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	var p []Product
	if err := json.Unmarshal(byteData, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return &p[0], nil

}

func (sr *SupabaseRepo) DeleteProduct(ctx context.Context, accessToken string, productID, ownerID uuid.UUID) error {
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
			return err
		}
	}

	_, count, err := client.From(string(constants.ProductTable)).Delete("", "exact").Eq("id", productID.String()).Eq("owner_id", ownerID.String()).Execute()
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("couldn't delete product with id:%s", productID)
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func (sr *SupabaseRepo) GetProductById(ctx context.Context, accessToken string, productId uuid.UUID, ownerID uuid.UUID) (*Product, error) {
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

	byteData, count, err := client.From(string(constants.ProductTable)).Select("*", "exact", false).Eq("id", productId.String()).Eq("owner_id", ownerID.String()).Execute()

	if err != nil {
		return nil, constants.ErrNoClient
	}
	if count == 0 {
		return nil, constants.ErrNotFound
	}

	var p []Product
	if err := json.Unmarshal(byteData, &p); err != nil {
		return nil, constants.ErrInternalServer
	}
	return &p[0], nil
}
