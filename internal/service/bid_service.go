package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/jwt"
	"github.com/joshua-takyi/auction/internal/models"
	"github.com/shopspring/decimal"
)

type BidService struct {
	bidRepo      models.BidInterface
	auctionRepo  models.AuctionInterface
	jwtManager   *jwt.JWTManager
	notifService *NotificationService
}

func NewBidService(bidRepo models.BidInterface, auctionRepo models.AuctionInterface, jwtManager *jwt.JWTManager, notifService *NotificationService) *BidService {
	return &BidService{
		bidRepo:      bidRepo,
		auctionRepo:  auctionRepo,
		jwtManager:   jwtManager,
		notifService: notifService,
	}
}

func (s *BidService) PlaceBid(ctx context.Context, amount decimal.Decimal, auctionID uuid.UUID, accessToken string) (map[string]any, error) {
	// 1. Verify user from token
	userAuth, err := s.jwtManager.VerifySupabaseToken(accessToken)
	if err != nil {
		return nil, constants.ErrUnauthorized
	}

	bidderID, err := uuid.Parse(userAuth.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id in token: %w", err)
	}

	// 2. Call the RPC for atomic bid placement
	result, err := s.bidRepo.PlaceBid(ctx, auctionID, bidderID, amount, accessToken)
	if err != nil {
		return nil, err
	}

	success, _ := result["success"].(bool)
	if !success {
		errMsg, _ := result["error"].(string)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// 3. Trigger Real-time Notifications
	roomID, _ := result["room_id"].(string)
	newPriceStr := amount.String()

	// Notify the room
	s.notifService.NotifyBidPlaced(roomID, bidderID.String(), newPriceStr)

	// Notify previous winner if they were outbid
	prevWinnerID, ok := result["previous_winner_id"].(string)
	if ok && prevWinnerID != "" && prevWinnerID != bidderID.String() {
		s.notifService.NotifyOutbid(prevWinnerID, roomID, newPriceStr)
	}

	return result, nil
}

func (s *BidService) GetBids(ctx context.Context, auctionID uuid.UUID, accessToken string, limit, offset int) ([]*models.Bid, int64, error) {
	if limit == 0 {
		limit = 10
	}

	if offset == 0 {
		offset = 0
	}

	return s.bidRepo.GetBids(ctx, auctionID, accessToken, limit, offset)
}

func (s *BidService) GetUserAuctionWithBid(ctx context.Context, userID uuid.UUID, accessToken string) ([]any, error) {
	return s.bidRepo.GetUserAuctionWithBid(ctx, userID, accessToken)
}
