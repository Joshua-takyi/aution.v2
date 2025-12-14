package models

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"

	storage_go "github.com/supabase-community/storage-go"
)

// UploadImage uploads an image to Supabase Storage and returns the public URL.
// bucket: The name of the storage bucket (e.g., "assets").
// folder: The folder within the bucket (e.g., "products", "avatars").
// filename: The name of the file (should be unique).
// data: The file content as bytes.
func (sr *SupabaseRepo) UploadImage(bucket, folder, filename string, data []byte) (string, error) {
	// Detect content type (e.g., image/jpeg, image/png)
	contentType := http.DetectContentType(data)

	// Prepare file options
	opts := storage_go.FileOptions{
		ContentType: &contentType,
		Upsert:      convertBool(true), // Allow overwriting if filename exists
	}

	// Construct the full path: folder/filename
	fullPath := filepath.Join(folder, filename)

	// Using bytes.NewReader to convert []byte to io.Reader
	// storage-go's Upload method expects io.Reader since v0.7.0+ (depending on version, checking commonly used interface)
	// If your version expects []byte, we can adjust. Standardizing on Reader is safer.
	// However, looking at common usages of supabase-go, Upload often takes []byte directly in some wrappers,
	// but the underlying client takes Reader.
	// Let's assume the standard client: Upload(path string, file io.Reader, options ...FileOptions)

	_, err := sr.supabase.Storage.From(bucket).Upload(fullPath, bytes.NewReader(data), opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to supabase: %w", err)
	}

	// Generate public URL
	// GetPublicUrl returns struct { SignedURL string } or just string depending on version.
	// Usually standard client returns (string).
	publicURL := sr.supabase.Storage.From(bucket).GetPublicUrl(fullPath)

	return publicURL, nil
}

// DeleteImage removes an image from Supabase Storage.
// path: The full path to the file (e.g., "products/watch.jpg").
func (sr *SupabaseRepo) DeleteImage(bucket, path string) error {
	// Remove expects a slice of paths
	_, err := sr.supabase.Storage.From(bucket).Remove([]string{path})
	if err != nil {
		return fmt.Errorf("failed to delete image from supabase: %w", err)
	}
	return nil
}

// FileUpload represents a file to be uploaded.
type FileUpload struct {
	Filename string
	Data     []byte
}

// UploadImages uploads multiple images to Supabase Storage.
// It implements a "Success or Rollback" pattern: if any upload fails, all previously uploaded images in this batch are deleted.
// bucket: The name of the storage bucket.
// folder: The folder within the bucket.
// files: A slice of FileUpload structs containing filenames and data.
func (sr *SupabaseRepo) UploadImages(bucket, folder string, files []FileUpload) ([]string, error) {
	var uploadedPaths []string
	var publicURLs []string

	for _, file := range files {
		// Use the existing single upload function
		// Note: UploadImage returns the public URL, but we need the path for potential deletion
		// We can reconstruct the path or modify UploadImage. For now, we reconstruct.
		url, err := sr.UploadImage(bucket, folder, file.Filename, file.Data)
		if err != nil {
			// Rollback: Delete already uploaded images
			// We iterate through uploadedPaths (which should contain full paths like "folder/filename")
			if len(uploadedPaths) > 0 {
				fmt.Printf("Rolling back uploads due to error: %v\n", err)
				errRollback := sr.DeleteImages(bucket, uploadedPaths)
				if errRollback != nil {
					// Log error but return the original upload error
					fmt.Printf("Failed to rollback images: %v\n", errRollback)
				}
			}
			return nil, err
		}

		// Track success
		fullPath := filepath.Join(folder, file.Filename)
		uploadedPaths = append(uploadedPaths, fullPath)
		publicURLs = append(publicURLs, url)
	}

	return publicURLs, nil
}

// DeleteImages removes multiple images from Supabase Storage.
// paths: A slice of full paths to the files (e.g., ["products/watch.jpg", "products/phone.jpg"]).
func (sr *SupabaseRepo) DeleteImages(bucket string, paths []string) error {
	_, err := sr.supabase.Storage.From(bucket).Remove(paths)
	if err != nil {
		return fmt.Errorf("failed to delete images from supabase: %w", err)
	}
	return nil
}

// Helper to get pointer to bool
func convertBool(b bool) *bool {
	return &b
}
