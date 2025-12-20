package handlers

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
)

func CreateProductHandler(s *service.ProductService, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "User not found", "")
			return
		}

		userModel, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}
		userID := userModel.ID
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

		accessToken, err := c.Cookie("access_token")
		if err != nil {
			utils.Unauthorized(c, "Unauthorized", "Please login")
			return
		}

		if !userModel.IsAdminOrSeller() {
			utils.Unauthorized(c, "Unauthorized", "only admins and verified users are authorized ")
			return
		}
		createdProduct, err := s.CreateProduct(c.Request.Context(), &product, accessToken, files, userID)
		if err != nil {
			switch {
			case errors.Is(err, constants.ErrDuplicateSlug):
				utils.BadRequest(c, constants.ErrDuplicateSlug.Error(), "")
			default:
				utils.BadRequest(c, "Failed to create product", "can't use same title twice")
			}
			return
		}

		utils.Created(c, "Product created successfully", createdProduct)
	}
}

func GetProductById(s *service.ProductService) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramId := c.Param(strings.TrimSpace("id"))

		if paramId == "" {
			utils.BadRequest(c, "product id can't be empty", "")
			return
		}

		user, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "User not found", "")
			return
		}

		_, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}

		accessToken, _ := c.Cookie("access_token")
		parsedProductId, err := uuid.Parse(paramId)

		if err != nil {
			switch {
			case errors.Is(err, constants.ErrInvalidID):
				utils.BadRequest(c, constants.ErrInvalidID.Error(), "")
			default:
				utils.NotFound(c, "product not found", "")
			}
			return
		}

		// get product by id
		product, err := s.GetProductById(c.Request.Context(), accessToken, parsedProductId)
		if err != nil {
			if errors.Is(err, constants.ErrNotFound) {
				utils.NotFound(c, "product not found", "")
				return
			}
			utils.BadRequest(c, "Failed to get product", "")
			return
		}

		utils.OK(c, "product retrieved successfully", product)
	}
}

func DeleteProduct(s *service.ProductService) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramId := c.Param(strings.TrimSpace("id"))
		if paramId == "" {
			utils.BadRequest(c, "product id can't be empty", "")
			return
		}

		user, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "User not found", "")
			return
		}

		_, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}

		accessToken, _ := c.Cookie("access_token")
		parsedProductId, err := uuid.Parse(paramId)

		if err != nil {
			switch {
			case errors.Is(err, constants.ErrInvalidID):
				utils.BadRequest(c, constants.ErrInvalidID.Error(), "")
			default:
				utils.NotFound(c, "product not found", "")
			}
			return
		}

		// Optimization: Removing redundant GetProductById.
		// We rely on RLS and the Delete operation result.
		err = s.DeleteProduct(c.Request.Context(), accessToken, parsedProductId)
		if err != nil {
			// Check specific errors if needed, e.g. "0 rows deleted" -> NotFound/Unauthorized
			utils.BadRequest(c, "Failed to delete product or unauthorized", err.Error())
			return
		}

		utils.DELETED(c, "product deleted successfully", nil)
	}
}

func GetProductWithAuctionHandler(s *service.ProductService) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramProductID := c.Param(strings.TrimSpace("id"))
		if paramProductID == "" {
			utils.BadRequest(c, "productID can't be empty", "product")
			return
		}
		productID, err := uuid.Parse(paramProductID)
		if err != nil {
			utils.BadRequest(c, "failed to parse string to uuid", "")
			return
		}
		res, err := s.GetProductWithAuction(c.Request.Context(), productID)
		if err != nil {
			switch {
			case errors.Is(err, constants.ErrNotFound):
				utils.BadRequest(c, "no data found matching ID", "")
			default:
				utils.InternalServerError(c, "internal server error", "")
			}
			return
		}
		utils.OK(c, "product returned successfully", res)
	}
}
func GetUserProductsHandler(s *service.ProductService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "User not found", "")
			return
		}

		userModel, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "User not found", "Please login")
			return
		}

		accessToken, _ := c.Cookie("access_token")

		products, count, err := s.GetProductsByOwner(c.Request.Context(), accessToken, userModel.ID, 10, 0)
		if err != nil {
			utils.InternalServerError(c, "Failed to get user products", err.Error())
			return
		}

		utils.OK(c, "Products retrieved successfully", gin.H{
			"data":  products,
			"count": count,
		})
	}
}
