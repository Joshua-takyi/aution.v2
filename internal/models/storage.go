package models

import (
	"fmt"
)

// UploadImage and UploadImages have been removed in favor of Cloudinary upload helper.
// See server/internal/helpers/upload.go

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
