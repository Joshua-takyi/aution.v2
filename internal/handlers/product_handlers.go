package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
)

func CreateProductHandler(s *service.ProductService, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse Multipart Form
		// Increase memory limit for parsing (default is 32MB)
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			utils.BadRequest(c, "Failed to parse multipart form", err.Error())
			return
		}

		// 2. Extract Product JSON
		jsonData := c.PostForm("data")
		if jsonData == "" {
			utils.BadRequest(c, "Missing product data", "")
			return
		}

		var product models.Product
		if err := json.Unmarshal([]byte(jsonData), &product); err != nil {
			logger.Warn("Invalid product JSON", map[string]interface{}{"error": err.Error()})
			utils.BadRequest(c, constants.ErrInvalidInput.Error(), "")
			return
		}

		// 3. Extract Files
		form, err := c.MultipartForm()
		if err != nil {
			utils.BadRequest(c, "Failed to get multipart form", err.Error())
			return
		}
		files := form.File["images"]

		var uploadFiles []models.FileUpload
		for _, fileHeader := range files {
			f, err := fileHeader.Open()
			if err != nil {
				logger.Warn("Failed to open file", map[string]interface{}{"filename": fileHeader.Filename, "error": err.Error()})
				continue
			}
			defer f.Close()

			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, f); err != nil {
				logger.Warn("Failed to read file", map[string]interface{}{"filename": fileHeader.Filename, "error": err.Error()})
				continue
			}

			uploadFiles = append(uploadFiles, models.FileUpload{
				Filename: fileHeader.Filename,
				Data:     buf.Bytes(),
			})
		}

		// Extract access token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "Authorization header required", "")
			return
		}
		accessToken := strings.TrimPrefix(authHeader, "Bearer ")

		createdProduct, err := s.CreateProduct(c.Request.Context(), &product, accessToken, uploadFiles)
		if err != nil {
			logger.Error("Failed to create product", err, nil)
			utils.InternalServerError(c, "Failed to create product", "")
			return
		}

		utils.Created(c, "Product created successfully", createdProduct)
	}
}
