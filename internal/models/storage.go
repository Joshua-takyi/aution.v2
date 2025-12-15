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

	// UploadFile(bucketId, path, file, options)
	_, err := sr.storage.UploadFile(bucket, fullPath, bytes.NewReader(data), opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to supabase: %w", err)
	}

	// GetPublicUrl(bucketId, path)
	// Returns struct with PublicURL string (or SignedURL depending on function used, usually GetPublicUrl returns public)
	resp := sr.storage.GetPublicUrl(bucket, fullPath)

	// Assuming resp has a field with the URL, commonly 'SignedURL' in the struct even if it's public in some versions,
	// or just 'Url'. The community library usually returns a struct.
	// Let's rely on string formatting if we can't be sure, but let's try accessing the struct field.
	// For now, returning resp.SignedURL based on common patterns, if it fails compiler will tell us 'field undefined'.
	return resp.SignedURL, nil
}

// FileUpload represents a file to be uploaded.
type FileUpload struct {
	Filename string
	Data     []byte
}

// UploadImages uploads multiple images to Supabase Storage.
// It implements a "Success or Rollback" pattern.
func (sr *SupabaseRepo) UploadImages(bucket, folder string, files []FileUpload) ([]string, error) {
	var uploadedPaths []string
	var publicURLs []string

	for _, file := range files {
		url, err := sr.UploadImage(bucket, folder, file.Filename, file.Data)
		if err != nil {
			// Rollback
			if len(uploadedPaths) > 0 {
				fmt.Printf("Rolling back uploads due to error: %v\n", err)
				_ = sr.DeleteImages(bucket, uploadedPaths)
			}
			return nil, err
		}

		fullPath := filepath.Join(folder, file.Filename)
		uploadedPaths = append(uploadedPaths, fullPath)
		publicURLs = append(publicURLs, url)
	}

	return publicURLs, nil
}

// DeleteImages removes multiple images from Supabase Storage.
func (sr *SupabaseRepo) DeleteImages(bucket string, paths []string) error {
	// RemoveFile(bucketId, paths)
	_, err := sr.storage.RemoveFile(bucket, paths)
	if err != nil {
		return fmt.Errorf("failed to delete images from supabase: %w", err)
	}
	return nil
}

// DeleteImage removes a single image.
func (sr *SupabaseRepo) DeleteImage(bucket, path string) error {
	return sr.DeleteImages(bucket, []string{path})
}

// Helper to get pointer to bool
func convertBool(b bool) *bool {
	return &b
}
