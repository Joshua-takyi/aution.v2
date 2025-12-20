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
			utils.BadRequest(c, "error", err.Error())
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

func ListAuctionsHandler(s *service.AuctionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params utils.PaginationParams
		if err := c.ShouldBindQuery(&params); err != nil {
			params = utils.DefaultPaginationParams()
		}
		params.Validate()

		// 2. Update service to return auctions AND total count
		auctions, total, err := s.ListAuctions(c.Request.Context(), params.GetLimit(), params.GetOffset())

		if err != nil {
			if errors.Is(err, constants.ErrNoData) {
				// 3. Keep response structure consistent even when empty
				utils.PaginatedOK(c, "no auctions found", []any{}, utils.NewPaginationMeta(params.Page, params.PageSize, 0))
				return
			}
			utils.InternalServerError(c, "failed to list auctions", err.Error())
			return
		}

		// 4. Create the metadata and send paginated response
		meta := utils.NewPaginationMeta(params.Page, params.PageSize, total)
		utils.PaginatedOK(c, "auctions listed successfully", auctions, meta)
	}
}

func SearchAuctionsHandler(s *service.AuctionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			utils.BadRequest(c, "query is required", "q")
			return
		}

		var params utils.PaginationParams
		if err := c.ShouldBindQuery(&params); err != nil {
			params = utils.DefaultPaginationParams()
		}
		params.Validate()

		auctions, total, err := s.SearchAuctions(c.Request.Context(), query, params.GetLimit(), params.GetOffset())

		if err != nil {
			if errors.Is(err, constants.ErrNoData) {
				utils.PaginatedOK(c, "no auctions found", []any{}, utils.NewPaginationMeta(params.Page, params.PageSize, 0))
				return
			}
			utils.InternalServerError(c, "failed to search auctions", err.Error())
			return
		}

		meta := utils.NewPaginationMeta(params.Page, params.PageSize, total)
		utils.PaginatedOK(c, "auctions searched successfully", auctions, meta)
	}
}

func FilterAuctionsHandler(s *service.AuctionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter models.AuctionFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			utils.BadRequest(c, "invalid filter parameters", err.Error())
			return
		}

		var params utils.PaginationParams
		if err := c.ShouldBindQuery(&params); err != nil {
			params = utils.DefaultPaginationParams()
		}
		params.Validate()

		auctions, total, err := s.FilterAuctions(c.Request.Context(), filter, params.GetLimit(), params.GetOffset())

		if err != nil {
			if errors.Is(err, constants.ErrNoData) {
				utils.PaginatedOK(c, "no auctions found", []any{}, utils.NewPaginationMeta(params.Page, params.PageSize, 0))
				return
			}
			utils.InternalServerError(c, "failed to filter auctions", err.Error())
			return
		}

		meta := utils.NewPaginationMeta(params.Page, params.PageSize, total)
		utils.PaginatedOK(c, "auctions filtered successfully", auctions, meta)
	}
}

func RecommendationHandler(s *service.AuctionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryParam := c.Query("category")
		currentID := c.Query("current_id")
		if categoryParam == "" {
			utils.BadRequest(c, "category field can't be empty", "auction")
			return
		}

		var params utils.PaginationParams
		if err := c.ShouldBindQuery(&params); err != nil {
			params = utils.DefaultPaginationParams()
		}

		auctions, total, err := s.Recommendation(c.Request.Context(), categoryParam, currentID, params.GetLimit(), params.GetOffset())
		if err != nil {
			if errors.Is(err, constants.ErrNoData) {
				utils.PaginatedOK(c, "no auctions found", []any{}, utils.NewPaginationMeta(params.Page, params.PageSize, 0))
				return
			}
			utils.InternalServerError(c, "failed to filter auctions", err.Error())
			return
		}

		meta := utils.NewPaginationMeta(params.Page, params.PageSize, total)
		utils.PaginatedOK(c, "auctions filtered successfully", auctions, meta)
	}

}
