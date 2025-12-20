package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
	"github.com/shopspring/decimal"
)

type BidRequest struct {
	Amount decimal.Decimal `json:"amount" validate:"required"`
}

func PlaceBidHandler(bidService *service.BidService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auctionIDStr := ctx.Param("id")
		auctionID, err := uuid.Parse(auctionIDStr)
		if err != nil {
			ctx.Error(constants.ErrInvalidID)
			return
		}

		var req BidRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Error(constants.ErrInvalidInput)
			return
		}

		accessToken, err := ctx.Cookie("access_token")
		if err != nil {
			// Fallback to Authorization header
			authHeader := ctx.GetHeader("Authorization")
			if authHeader != "" {
				accessToken = authHeader[7:] // Remove "Bearer "
			}
		}

		if accessToken == "" {
			ctx.Error(constants.ErrUnauthorized)
			return
		}

		result, err := bidService.PlaceBid(ctx, req.Amount, auctionID, accessToken)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func GetBids(bidservice *service.BidService) gin.HandlerFunc {
	return func(c *gin.Context) {
		aucitionParam := c.Param(strings.TrimSpace("id"))
		if aucitionParam == "" {
			utils.BadRequest(c, "auction id can't be empty", "bids")
			return
		}

		parsedAuctionID, err := uuid.Parse(aucitionParam)
		if err != nil {
			utils.BadRequest(c, "failed to parse uuid to string", "bids")
			return
		}
		var params utils.PaginationParams
		if err := c.ShouldBindQuery(&params); err != nil {
			params = utils.DefaultPaginationParams()
		}

		accessToken, err := c.Cookie("access_token")
		if err != nil {
			// Fallback to Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				accessToken = authHeader[7:] // Remove "Bearer "
			}
		}

		auctions, total, err := bidservice.GetBids(c.Request.Context(), parsedAuctionID, accessToken, params.GetLimit(), params.GetOffset())
		if err != nil {
			if errors.Is(err, constants.ErrNoData) {
				utils.PaginatedOK(c, "no bids found", []any{}, utils.NewPaginationMeta(params.Page, params.PageSize, 0))
				return
			}
			utils.InternalServerError(c, "failed to filter auctions", err.Error())
			return
		}

		meta := utils.NewPaginationMeta(params.Page, params.PageSize, total)
		utils.PaginatedOK(c, "auctions filtered successfully", auctions, meta)
	}
}

func GetUserAuctionWithBidHandler(s *service.BidService) gin.HandlerFunc {
	return func(c *gin.Context) {

		user, exist := c.Get("user")
		if !exist {
			utils.Unauthorized(c, "unauthenticated user", "error")
		}
		claims, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "unauthorized access", "error")
		}
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				accessToken = authHeader[7:]
			}
		}
		userID, err := uuid.Parse(claims.ID.String())
		if err != nil {
			utils.BadRequest(c, "failed to parse string to uuid", "bid")
			return
		}
		bids, err := s.GetUserAuctionWithBid(c.Request.Context(), userID, accessToken)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		utils.PaginatedOK(c, "data returned successfully", bids, nil)
	}
}
