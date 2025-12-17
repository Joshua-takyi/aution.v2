package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/shopspring/decimal"
)

type Auction struct {
	ID             uuid.UUID        `db:"id" json:"id"`
	ProductID      uuid.UUID        `db:"product_id" json:"product_id" validate:"required"`
	MinIncrement   decimal.Decimal  `db:"min_increment" json:"min_increment" validate:"required"`
	ReservePrice   *decimal.Decimal `db:"reserve_price" json:"reserve_price" validate:"required"`
	CurrentBid     decimal.Decimal  `db:"current_bid" json:"current_bid"`
	BuyNowPrice    decimal.Decimal  `db:"buy_now_price" json:"buy_now_price"`
	BuyNow         bool             `db:"buy_now" json:"buy_now"`
	EstimatedPrice decimal.Decimal  `db:"estimated_price" json:"estimated_price" validate:"required"`
	StartPrice     decimal.Decimal  `db:"start_price" json:"start_price" validate:"required"`
	StartTime      time.Time        `db:"start_time" json:"start_time" validate:"required"`
	EndTime        time.Time        `db:"end_time" json:"end_time" validate:"required"`
	WinnerID       *uuid.UUID       `db:"winner_id" json:"winner_id"`
	Status         string           `db:"status" json:"status"`
	RoomID         uuid.UUID        `db:"room_id" json:"room_id"`
	CreatedAt      time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at" json:"updated_at"`
}

type AuctionResponse struct {
	Auction
	Product Product `json:"products"`
}

// calculateDuration is to calculate the duration of the auction
func (a *Auction) CalculateDuration() time.Duration {
	return a.EndTime.Sub(a.StartTime)
}

// get live auction
func (a *Auction) GetLiveAuction() bool {
	return a.Status == constants.AuctionLive
}

// calculateRemainingTime is to calculate the remaining time of the auction
func (a *Auction) CalculateRemainingTime() time.Duration {
	return time.Until(a.EndTime)
}

type AuctionInterface interface {
	CreateAuction(ctx context.Context, auction *Auction, accessToken string, productID uuid.UUID) (*Auction, error)
	UpdateAuction(ctx context.Context, auction map[string]any, accessToken string, auctionID uuid.UUID) (*Auction, error)
	DeleteAuction(ctx context.Context, accessToken string, auctionID uuid.UUID) (string, error)
	GetActiveAuctionByProductID(ctx context.Context, accessToken string, productID uuid.UUID) (*Auction, error)
	GetAuctionById(ctx context.Context, auctionID uuid.UUID) (*AuctionResponse, error)
	GetAuctionsByProductID(ctx context.Context, accessToken string, productID uuid.UUID, limit, offset int) ([]*Auction, error)
}

func (sr *SupabaseRepo) CreateAuction(ctx context.Context, auction *Auction, accessToken string, productID uuid.UUID) (*Auction, error) {
	client, err := sr.GetAuthenticatedClient(accessToken)
	if err != nil {
		return nil, constants.ErrNoClient
	}

	fmt.Printf("[SupabaseRepo] Inserting auction: %v\n", auction)
	byteData, count, err := client.From(string(constants.AuctionTable)).Insert(auction, false, "", "", "exact").Execute()
	if err != nil {
		fmt.Printf("[SupabaseRepo] Error executing insert: %v\n", err)
		return nil, fmt.Errorf("failed to insert auction: %w", err)
	}

	fmt.Printf("[SupabaseRepo] Insert returned count: %d, Response: %s\n", count, string(byteData))

	if len(byteData) == 0 || string(byteData) == "null" {
		return nil, fmt.Errorf("failed to insert auction: empty response from supabase")
	}

	var a []Auction
	if err := json.Unmarshal(byteData, &a); err != nil {
		fmt.Printf("[SupabaseRepo] Failed to unmarshal: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal auction: %w", err)
	}
	if len(a) == 0 {
		return nil, fmt.Errorf("failed to insert auction: no data returned")
	}
	fmt.Printf("[SupabaseRepo] Unmarshalled auction: %v\n", a)
	return &a[0], nil
}

func (sr *SupabaseRepo) UpdateAuction(ctx context.Context, auction map[string]any, accessToken string, auctionID uuid.UUID) (*Auction, error) {

	return nil, nil
}

func (sr *SupabaseRepo) DeleteAuction(ctx context.Context, accessToken string, auctionID uuid.UUID) (string, error) {
	client, err := sr.GetAuthenticatedClient(accessToken)
	if err != nil {
		return "", constants.ErrNoClient
	}
	_, count, err := client.From(string(constants.AuctionTable)).Delete("", "exact").Eq("id", auctionID.String()).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete auction: %w", err)
	}
	if count == 0 {
		return "", constants.ErrNoData
	}
	return "", nil
}

func (sr *SupabaseRepo) GetAuctionById(ctx context.Context, auctionID uuid.UUID) (*AuctionResponse, error) {

	byteData, count, err := sr.supabase.From(string(constants.AuctionTable)).Select("*, products(*)", "exact", false).Eq("id", auctionID.String()).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get auction: %w", err)
	}
	if count == 0 {
		return nil, constants.ErrNotFound
	}

	var res []*AuctionResponse
	if err := json.Unmarshal(byteData, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auction: %w", err)
	}
	return res[0], nil
}

func (sr *SupabaseRepo) GetAuctionsByProductID(ctx context.Context, accessToken string, productID uuid.UUID, limit, offset int) ([]*Auction, error) {
	client, err := sr.GetAuthenticatedClient(accessToken)
	if err != nil {
		return nil, constants.ErrNoClient
	}

	res, count, err := client.From(string(constants.AuctionTable)).Select("*, products(*)", "exact", false).Range(offset, offset+limit-1, "").Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get auctions: %w", err)
	}

	if count == 0 {
		return nil, constants.ErrNoData
	}

	var a []*Auction
	if err := json.Unmarshal(res, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auctions: %w", err)
	}
	return a, nil
}

func (sr *SupabaseRepo) GetActiveAuctionByProductID(ctx context.Context, accessToken string, productID uuid.UUID) (*Auction, error) {
	client, err := sr.GetAuthenticatedClient(accessToken)
	if err != nil {
		return nil, constants.ErrNoClient
	}

	res, count, err := client.From(string(constants.AuctionTable)).Select("*, products(*)", "exact", false).Eq("product_id", productID.String()).Eq("status", constants.AuctionLive).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get active auction: %w", err)
	}
	if count == 0 {
		return nil, constants.ErrNoData
	}
	var a []*Auction
	if err := json.Unmarshal(res, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal active auction: %w", err)
	}
	return a[0], nil
}
