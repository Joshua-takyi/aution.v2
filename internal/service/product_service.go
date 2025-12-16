package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"sync"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/helpers"
	"github.com/joshua-takyi/auction/internal/models"
)

const (
	productStorage = "assets"
)

type ProductService struct {
	productRepo models.ProductInterface
	cloudinary  *cloudinary.Cloudinary
}

func NewProductService(productRepo models.ProductInterface, cloudinary *cloudinary.Cloudinary) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		cloudinary:  cloudinary,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product, accessToken string, files []*multipart.FileHeader, userID uuid.UUID) (*models.Product, error) {

	if product.Title == "" {
		fmt.Println("[ProductService] ERROR: Empty title")
		return nil, constants.ErrEmptyFields
	}

	// 1. Upload images concurrently
	// fmt.Printf("[ProductService] Starting concurrent upload of %d images...\n", len(files))
	imageUrls := make([]string, len(files))
	publicIdChan := make(chan string, len(files))
	imageChan := make(chan string, len(files))
	uploadedPaths := make([]string, 0, len(files))
	errHolder := make(chan error, len(files))
	errChan := make(chan error, len(files))
	var wg sync.WaitGroup

	for i, fileHeader := range files {
		wg.Add(1)
		go func(idx int, fh *multipart.FileHeader) {
			defer wg.Done()
			// fmt.Printf("[ProductService] Processing file %d: %s (size: %d bytes)\n", idx, fh.Filename, fh.Size)

			file, err := fh.Open()
			if err != nil {
				// fmt.Printf("[ProductService] ERROR: Failed to open file %d (%s): %v\n", idx, fh.Filename, err)
				errChan <- err
				return
			}
			defer file.Close()

			// fmt.Printf("[ProductService] Uploading file %d to Cloudinary...\n", idx)
			url, pId, err := helpers.UploadImage(ctx, s.cloudinary, file, productStorage)
			if err != nil {
				// fmt.Printf("[ProductService] ERROR: Cloudinary upload failed for file %d: %v\n", idx, err)
				errChan <- err
				return
			}
			// fmt.Printf("[ProductService] File %d uploaded successfully. URL: %s, PublicID: %s\n", idx, url, pId)
			publicIdChan <- pId
			imageChan <- url
		}(i, fileHeader)
	}

	// Close channels when done
	go func() {
		wg.Wait()
		// fmt.Println("[ProductService] All goroutines completed, closing channels")
		close(publicIdChan)
		close(imageChan)
		close(errChan)
	}()

	// Collect results
	// fmt.Println("[ProductService] Collecting upload results...")
	for url := range imageChan {
		imageUrls = append(imageUrls, url)
	}
	// fmt.Printf("[ProductService] Collected %d image URLs\n", len(imageUrls))

	for pId := range publicIdChan {
		uploadedPaths = append(uploadedPaths, pId)
	}
	// fmt.Printf("[ProductService] Collected %d public IDs\n", len(uploadedPaths))

	// Check for errors
	for err := range errChan {
		if err != nil {
			errHolder <- err
		}
	}

	if len(errHolder) > 0 {
		// fmt.Printf("[ProductService] ERROR: Upload failed, rolling back %d images\n", len(uploadedPaths))
		// Rollback: delete successfully uploaded images
		if len(uploadedPaths) > 0 {
			// Best effort cleanup
			_ = helpers.DeleteImages(ctx, s.cloudinary, productStorage, uploadedPaths)
		}
		return nil, <-errHolder
	}
	// fmt.Println("[ProductService] All images uploaded successfully")

	// 2. Add image URLs to product
	// fmt.Println("[ProductService] Preparing product for database insertion...")
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
		// fmt.Printf("[ProductService] Generated new product ID: %s\n", product.ID)
	}
	now := time.Now()
	product.Images = imageUrls
	product.CreatedAt = now
	product.UpdatedAt = now
	product.Slug = helpers.GenerateSlug(product.Title, product.Category)
	product.Status = constants.ProductDraft
	product.OwnerID = userID
	// fmt.Printf("[ProductService] Product prepared: ID=%s, Title=%s, Images=%d\n", product.ID, product.Title, len(product.Images))

	// 3. Save product to database
	// fmt.Println("[ProductService] Attempting to save product to database...")
	createdProduct, err := s.productRepo.CreateProduct(ctx, product, accessToken, userID)
	if err != nil {
		// fmt.Printf("[ProductService] ERROR: Database insertion failed: %v\n", err)
		// fmt.Printf("[ProductService] Rolling back %d uploaded images\n", len(uploadedPaths))
		// Rollback: delete successfully uploaded images
		if len(uploadedPaths) > 0 {
			_ = helpers.DeleteImages(ctx, s.cloudinary, productStorage, uploadedPaths)
		}
		return nil, err
	}
	// fmt.Printf("[ProductService] SUCCESS: Product created with ID: %s\n", createdProduct.ID)
	return createdProduct, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, product map[string]any, accessToken string, productID uuid.UUID) (*models.Product, error) {
	/*
		TODO
			check if we don't have an active action running
			if we do, return error

	*/

	return s.productRepo.UpdateProduct(ctx, product, accessToken, productID)
}

func (s *ProductService) DeleteProduct(ctx context.Context, accessToken string, productID uuid.UUID, ownerID uuid.UUID) error {
	return s.productRepo.DeleteProduct(ctx, accessToken, productID, ownerID)
}

func (s *ProductService) GetProductById(ctx context.Context, accessToken string, productID uuid.UUID, ownerID uuid.UUID) (*models.Product, error) {
	if productID == uuid.Nil {
		return nil, constants.ErrInvalidID
	}

	return s.productRepo.GetProductById(ctx, accessToken, productID, ownerID)
}
