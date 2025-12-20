package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/shopspring/decimal"
	"github.com/supabase-community/postgrest-go"
)

type Bid struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	AuctionID uuid.UUID       `db:"auction_id" json:"auction_id"`
	BidBy     uuid.UUID       `db:"bid_by" json:"bid_by"`
	BidAmount decimal.Decimal `db:"bid_amount" json:"bid_amount"`
	BidAt     time.Time       `db:"bid_at" json:"bid_at"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
	IsWinning bool            `db:"is_winning" json:"is_winning"`
	Profile   *User           `json:"profiles"`
}

type BidInterface interface {
	PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount decimal.Decimal, accessToken string) (map[string]any, error)
	GetBids(ctx context.Context, auctionID uuid.UUID, accessToken string, limit, offset int) ([]*Bid, int64, error)
	GetUserAuctionWithBid(ctx context.Context, userID uuid.UUID, accessToken string) ([]any, error)
}

func (sr *SupabaseRepo) PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount decimal.Decimal, accessToken string) (map[string]any, error) {
	client, err := sr.GetAuthenticatedClient(accessToken)
	if err != nil {
		return nil, constants.ErrNoClient
	}

	params := map[string]any{
		"p_auction_id": auctionID.String(),
		"p_bidder_id":  bidderID.String(),
		"p_amount":     amount,
	}

	// Workaround: supabase-go's Rpc returns a string, so we use From with the rpc/ prefix
	// which returns a QueryBuilder that we can Execute.
	res, _, err := client.From("rpc/place_bid").Insert(params, false, "", "", "exact").Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to call place_bid rpc: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(res), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rpc result: %w", err)
	}

	return result, nil
}

func (sr *SupabaseRepo) GetBids(ctx context.Context, auctionID uuid.UUID, accessToken string, limit, offset int) ([]*Bid, int64, error) {
	client := sr.supabase
	if sr.serviceClient != nil {
		client = sr.serviceClient
	}
	byteData, count, err := client.From(string(constants.BidTable)).
		Select("*, auctions!inner(*), profiles!inner(*)", "exact", false).
		Eq("auction_id", auctionID.String()).
		Range(offset, offset+limit-1, "").

		// TODO: learn about this package
		Order("created_at", &postgrest.OrderOpts{Ascending: true}).
		Execute()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to load data bids from bid table %w", err)
	}

	if count == 0 {
		return nil, 0, constants.ErrNoData
	}

	var res []*Bid

	if err := json.Unmarshal(byteData, &res); err != nil {
		return nil, 0, fmt.Errorf("failed to marshal byte to json")
	}

	return res, count, nil
}

func (sr *SupabaseRepo) GetUserAuctionWithBid(ctx context.Context, userID uuid.UUID, accessToken string) ([]any, error) {
	client, err := sr.GetAuthenticatedClient(accessToken)
	if err != nil {
		return nil, constants.ErrNoClient
	}

	byteDate, count, err := client.From(string(constants.BidTable)).
		Select("*,auctions!inner(*,products!inner(*))", "exact", false).
		Eq("bid_by", userID.String()).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to load user bid data %w", err)
	}

	if count == 0 {
		return nil, constants.ErrNotFound
	}

	var s []map[string]any

	if err := json.Unmarshal(byteDate, &s); err != nil {
		return nil, fmt.Errorf("failed to marshal data into json %w", err)
	}

	// Filter to keep only the latest bid for each unique auction_id
	uniqueAuctions := make(map[string]bool)
	var latestBids []any

	for _, bid := range s {
		auctionID, ok := bid["auction_id"].(string)
		if !ok {
			continue
		}

		if !uniqueAuctions[auctionID] {
			latestBids = append(latestBids, bid)
			uniqueAuctions[auctionID] = true
		}
	}

	return latestBids, nil
}
