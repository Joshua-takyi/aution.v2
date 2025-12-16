package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadImage(ctx context.Context, cld *cloudinary.Cloudinary, input any, folder string) (string, string, error) {
	uploadResult, err := cld.Upload.Upload(ctx, input, uploader.UploadParams{
		Folder: folder,
		Tags:   []string{"ww-app"},
	})

	if err != nil {
		return "", "", fmt.Errorf("upload failed: %v", err)
	}

	return uploadResult.SecureURL, uploadResult.PublicID, nil
}

func DeleteImages(ctx context.Context, cld *cloudinary.Cloudinary, folderName string, publicIDs []string) error {
	for _, rawID := range publicIDs {
		publicID := strings.TrimSpace(rawID)
		if publicID == "" {
			continue
		}

		// Ensure folder prefix if provided
		if folderName != "" && !strings.HasPrefix(publicID, folderName+"/") {
			publicID = fmt.Sprintf("%s/%s", folderName, publicID)
		}

		// Attempt deletion
		resp, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
			PublicID: publicID,
		})
		if err != nil {
			fmt.Printf("[Cloudinary] Error deleting '%s': %v\n", publicID, err)
			continue
		}

		switch resp.Result {
		case "ok":
			fmt.Printf("[Cloudinary] Deleted: %s\n", publicID)
		case "not found":
			fmt.Printf("[Cloudinary] Not found: %s\n", publicID)
		default:
			fmt.Printf("[Cloudinary] Unexpected result for '%s': %s\n", publicID, resp.Result)
		}
	}

	return nil
}
