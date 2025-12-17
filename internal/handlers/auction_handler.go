package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
)

func CreateAuctionHandler(s *service.AuctionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param(strings.TrimSpace("id"))
		if productID == "" {
			utils.BadRequest(c, "product id is required", "product_id")
			return
		}

		parsedProductID, err := uuid.Parse(productID)
		if err != nil {
			utils.BadRequest(c, "invalid product id", "product_id")
			return
		}

		user, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "user not found", "user")
			return
		}

		claims, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "user not found", "user")
			return
		}

		if !claims.IsAdminOrSeller() {
			utils.Unauthorized(c, "user not authorized", "user")
			return
		}

		var auction models.Auction
		if err := c.ShouldBindJSON(&auction); err != nil {
			utils.BadRequest(c, "invalid auction data", "auction")
			return
		}

		accessToken, err := c.Cookie("access_token")
		if err != nil {
			utils.Unauthorized(c, "user not found", "user")
			return
		}
		created, err := s.CreateAuction(c.Request.Context(), &auction, accessToken, parsedProductID)
		if err != nil {
			switch {
			case errors.Is(err, constants.ErrProductHasActiveAuction):
				utils.BadRequest(c, "product has already been added auction", productID)
			case errors.Is(err, constants.ErrProductNotApproved):
				utils.BadRequest(c, "product is not approved for auction", productID)
			case errors.Is(err, constants.ErrInvalidInput):
				utils.BadRequest(c, "invalid input", "auction")
			default:
				utils.InternalServerError(c, "failed to create auction", "auction")
			}
			return
		}

		utils.Created(c, "auction created successfully", created)
	}
}

func DeleteAuctionHandler(s *service.AuctionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		auctionParamId := c.Param(strings.TrimSpace("id"))
		if auctionParamId == "" {
			utils.BadRequest(c, "auction id can't be empty", auctionParamId)
			return
		}

		auctionID, err := uuid.Parse(auctionParamId)
		if err != nil {
			utils.BadRequest(c, "invalid auction id", "auction_id")
			return
		}
		accessToken, _ := c.Cookie("access_token")

		user, exists := c.Get("user")
		if !exists {
			utils.Unauthorized(c, "user not authenticated", "user")
			return
		}

		claims, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "invalid user token", "user")
			return
		}

		returnedAuction, err := s.GetAuctionById(c.Request.Context(), auctionID)

		if err != nil {
			utils.NotFound(c, "no data found matching the id", err.Error())
			return
		}

		if !claims.IsAdminOrOwner(returnedAuction.Product.OwnerID) {
			utils.Unauthorized(c, "unauthorized access", "user")
		}

		_, err = s.DeleteAuction(c.Request.Context(), accessToken, auctionID)
		if err != nil {
			utils.InternalServerError(c, "failed to delete auction", err.Error())
			return
		}
		utils.OK(c, "auction deleted successfully", "")
	}
}

func GetAuctionByIdHandler(s *service.AuctionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		auctionParamId := c.Param(strings.TrimSpace("id"))
		if auctionParamId == "" {
			utils.BadRequest(c, "auction id can't be empty", auctionParamId)
			return
		}

		parsedAuctionId, err := uuid.Parse(auctionParamId)
		if err != nil {
			utils.BadRequest(c, "invalid auction id", "auction_id")
			return
		}
		returnedAuction, err := s.GetAuctionById(c.Request.Context(), parsedAuctionId)

		if err != nil {
			switch {
			case errors.Is(err, constants.ErrNotFound):
				utils.NotFound(c, "no data found matching the id", "auction")
			case errors.Is(err, constants.ErrInvalidID):
				utils.BadRequest(c, "invalid auction id", "auction_id")
			case errors.Is(err, constants.ErrNoClient):
				utils.InternalServerError(c, "internal server error", "server error")
			default:
				utils.InternalServerError(c, "internal server error", "server error")

			}
			return
		}

		utils.OK(c, "auction retrieved successfully", returnedAuction)

	}
}
