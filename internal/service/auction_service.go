package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/helpers"
	"github.com/joshua-takyi/auction/internal/models"
)

type AuctionService struct {
	auctionRepo models.AuctionInterface
	productRepo models.ProductInterface
}

func NewAuctionService(auctionRepo models.AuctionInterface, productRepo models.ProductInterface) *AuctionService {
	return &AuctionService{
		auctionRepo: auctionRepo,
		productRepo: productRepo,
	}
}

func (s *AuctionService) CreateAuction(ctx context.Context, auction *models.Auction, accessToken string, productID uuid.UUID) (*models.Auction, error) {

	auction.ProductID = productID

	now := time.Now()
	if auction.StartTime.IsZero() {
		auction.StartTime = now.Add(time.Hour)
	}

	if auction.EndTime.IsZero() {
		auction.EndTime = auction.StartTime.Add(24 * time.Hour * 2)
	}
	if auction.ID == uuid.Nil {
		auction.ID = uuid.New()
	}

	if auction.RoomID == uuid.Nil {
		auction.RoomID = uuid.New()
	}

	auction.CreatedAt = now
	auction.UpdatedAt = now
	auction.Status = constants.AuctionScheduled

	if err := models.Validate.Struct(auction); err != nil {
		fmt.Printf("[AuctionService] Validation failed: %v\n", err)
		return nil, constants.ErrInvalidInput
	}

	product, err := s.productRepo.GetProductById(ctx, accessToken, productID)
	if err != nil {
		return nil, err
	}
	if product.Status != constants.ProductApproved {
		return nil, constants.ErrProductNotApproved
	}

	fmt.Printf("[AuctionService] checking if product %s already have an existing and active auction\n", productID)

	// returnedAuction, err := s.auctionRepo.GetAuctionsByProductID(ctx, accessToken, productID, 1, 0)
	// if err != nil {
	// 	if err == constants.ErrNoData {
	// 		// No active auction found, which is what we want
	// 		fmt.Printf("[AuctionService] No active auction found for product %s, proceeding...\n", productID)
	// 	} else {
	// 		// Real error occurred
	// 		fmt.Printf("[AuctionService] Failed to check for active auctions: %v\n", err)
	// 		return nil, fmt.Errorf("failed to check for active auction: %w", err)
	// 	}
	// }
	// if len(returnedAuction) > 0 {
	// 	return nil, constants.ErrProductHasActiveAuction
	// }

	// check if the product doesn't have an active auction
	activeAuction, err := s.auctionRepo.GetActiveAuctionByProductID(ctx, accessToken, productID)
	if err != nil {
		if err == constants.ErrNoData {
			// No active auction found, which is what we want
			fmt.Printf("[AuctionService] No active auction found for product %s, proceeding...\n", productID)
		} else {
			// Real error occurred
			fmt.Printf("[AuctionService] Failed to check for active auctions: %v\n", err)
			return nil, fmt.Errorf("failed to check for active auction: %w", err)
		}
	} else if activeAuction != nil {
		fmt.Printf("[AuctionService] Active auction found: %s\n", activeAuction.ID)
		return nil, constants.ErrProductHasActiveAuction
	}

	// checking if the start time is less than the end time
	if auction.StartTime.After(auction.EndTime) {
		return nil, constants.ErrInvalidInput
	}

	if auction.MinIncrement.IsZero() {
		return nil, fmt.Errorf("auction min increment cannot be zero: %w", constants.ErrInvalidInput)
	}

	if auction.ReservePrice.IsZero() {
		return nil, fmt.Errorf("auction reserve price cannot be zero: %w", constants.ErrInvalidInput)
	}

	if auction.EstimatedPrice.IsZero() {
		return nil, fmt.Errorf("auction estimated price cannot be zero: %w", constants.ErrInvalidInput)
	}

	if auction.StartPrice.IsZero() {
		return nil, fmt.Errorf("auction start price cannot be zero: %w", constants.ErrInvalidInput)
	}

	fmt.Printf("[AuctionService] creating auction for product %s\n", productID)
	created, err := s.auctionRepo.CreateAuction(ctx, auction, accessToken, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to create auction: %w", err)
	}
	fmt.Printf("[AuctionService] auction created successfully for product %s\n", created.ID)
	return created, nil
}

func (s *AuctionService) DeleteAuction(ctx context.Context, accessToken string, auctionID uuid.UUID) (string, error) {
	if auctionID == uuid.Nil {
		return "", constants.ErrInvalidID
	}
	// check if auction is live or active users can't delete it
	returnedAuction, err := s.auctionRepo.GetAuctionById(ctx, auctionID)
	if err != nil {
		return "", fmt.Errorf("failed to load auction: %w", err)
	}
	if returnedAuction.Auction.Status == constants.AuctionLive {
		return "", constants.ErrLiveAuction
	}

	return s.auctionRepo.DeleteAuction(ctx, accessToken, auctionID)
}

func (s *AuctionService) GetAuctionById(ctx context.Context, auctionID uuid.UUID) (*models.AuctionResponse, error) {
	auction, err := s.auctionRepo.GetAuctionById(ctx, auctionID)
	if err != nil {
		return nil, err
	}
	// auction.Auction.ReservePrice = &decimal.Decimal{}
	return auction, nil
}

func (s *AuctionService) ListAuctions(ctx context.Context, limit, offset int) ([]*models.AuctionResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.auctionRepo.ListAuctions(ctx, limit, offset)
}

func (s *AuctionService) SearchAuctions(ctx context.Context, query string, limit, offset int) ([]*models.AuctionResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.auctionRepo.SearchAuctions(ctx, query, limit, offset)
}

func (s *AuctionService) FilterAuctions(ctx context.Context, filter models.AuctionFilter, limit, offset int) ([]*models.AuctionResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.auctionRepo.FilterAuctions(ctx, filter, limit, offset)
}

func (s *AuctionService) UpdateAuctionStatuses(ctx context.Context) (map[string]any, error) {
	return s.auctionRepo.UpdateAuctionStatuses(ctx)
}

func (s *AuctionService) Recommendation(ctx context.Context, category string, currentID string, limit, offset int) ([]*models.AuctionResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.auctionRepo.Recommendation(ctx, category, currentID, limit, offset)
}

func (s *AuctionService) GetAuctionSummary(ctx context.Context, userID uuid.UUID, limit, offset int, accessToken string) ([]models.AuctionResponse, error) {
	if limit == 0 {
		limit = 10
	}
	if offset == 0 {
		offset = 0
	}

	return s.auctionRepo.GetAuctionSummary(ctx, userID, limit, offset, accessToken)
}

func (s *AuctionService) GetUserAuctions(ctx context.Context, userID uuid.UUID, limit, offset int, accessToken string) ([]models.AuctionResponse, int64, error) {

	lim, off := helpers.DefaultLimitAndOffset(limit, offset)

	return s.auctionRepo.GetUserAuctions(ctx, userID, lim, off, accessToken)
}
