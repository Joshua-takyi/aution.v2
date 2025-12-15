package service

import (
	"context"
	"sync"

	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/models"
)

type ProductService struct {
	productRepo models.ProductInterface
}

func NewProductService(productRepo models.ProductInterface) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product, accessToken string, files []models.FileUpload) (*models.Product, error) {
	if product.Title == "" || product.Price == 0 {
		return nil, constants.ErrEmptyFields
	}

	// 1. Upload images concurrently
	imageUrls := make([]string, len(files))
	errChan := make(chan error, len(files))
	var wg sync.WaitGroup

	for i, file := range files {
		wg.Add(1)
		go func(idx int, f models.FileUpload) {
			defer wg.Done()
			// Bucket "assets", folder "products"
			url, err := s.productRepo.UploadImage("assets", "products", f.Filename, f.Data)
			if err != nil {
				errChan <- err
				return
			}
			imageUrls[idx] = url
		}(i, file)
	}

	// Wait for all uploads to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	// 2. Add image URLs to product
	product.Images = imageUrls

	// 3. Save product to database
	return s.productRepo.CreateProduct(ctx, product, accessToken)
}
