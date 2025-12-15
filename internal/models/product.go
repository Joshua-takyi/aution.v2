package models

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
)

type Product struct {
	ID          uuid.UUID        `db:"id" json:"id"`
	OwnerID     uuid.UUID        `db:"owner_id" json:"owner_id"`
	Title       string           `db:"title" json:"title"`
	Category    []string         `db:"category" json:"category"`
	Description string           `db:"description" json:"description"`
	Slug        string           `db:"slug" json:"slug"`
	Price       float64          `db:"price" json:"price"`
	Images      []string         `db:"images" json:"images"`
	Details     []map[string]any `db:"details" json:"details"`
	CreatedAt   time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updated_at"`
}

type ProductInterface interface {
	CreateProduct(ctx context.Context, product *Product, accessToken string) (*Product, error)
	UploadImage(bucket, folder, filename string, data []byte) (string, error)
	UploadImages(bucket, folder string, files []FileUpload) ([]string, error)
}

func (su *SupabaseRepo) CreateProduct(ctx context.Context, product *Product, accessToken string) (*Product, error) {

	client, err := su.GetAuthenticatedClient(accessToken)
	if err != nil {
		return nil, errors.New("failed to get authenticated client")
	}

	res, _, err := client.From(string(constants.ProductTable)).Insert(product, false, "", "", "exact").Execute()

	if err != nil {
		return nil, errors.New("failed to insert product into table")
	}

	var p Product
	if err := json.Unmarshal(res, &p); err != nil {
		return nil, errors.New("failed to unmarshal product")
	}

	return &p, nil
}
