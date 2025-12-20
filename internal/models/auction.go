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

type AuctionFilter struct {
	Category string   `form:"category"`
	MinPrice *float64 `form:"min_price"`
	MaxPrice *float64 `form:"max_price"`
	Status   string   `form:"status"`
	SortBy   string   `form:"sort_by"`
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
	ListAuctions(ctx context.Context, limit, offset int) ([]*AuctionResponse, int64, error)
	DeleteAuction(ctx context.Context, accessToken string, auctionID uuid.UUID) (string, error)
	GetActiveAuctionByProductID(ctx context.Context, accessToken string, productID uuid.UUID) (*Auction, error)
	GetAuctionById(ctx context.Context, auctionID uuid.UUID) (*AuctionResponse, error)
	GetAuctionsByProductID(ctx context.Context, accessToken string, productID uuid.UUID, limit, offset int) ([]*Auction, error)
	Recommendation(ctx context.Context, category string, currentID string, limit, offset int) ([]*AuctionResponse, int64, error)
	SearchAuctions(ctx context.Context, query string, limit, offset int) ([]*AuctionResponse, int64, error)
	FilterAuctions(ctx context.Context, filter AuctionFilter, limit, offset int) ([]*AuctionResponse, int64, error)
	UpdateAuctionStatuses(ctx context.Context) (map[string]any, error)
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

func (sr *SupabaseRepo) ListAuctions(ctx context.Context, limit, offset int) ([]*AuctionResponse, int64, error) {
	byteData, count, err := sr.supabase.From(string(constants.AuctionTable)).
		Select("*, products(*)", "exact", false).Limit(limit, "").
		Range(offset, offset+limit-1, "").
		Execute()

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get auctions: %w", err)
	}
	if count == 0 {
		return nil, 0, constants.ErrNoData
	}

	var response []*AuctionResponse
	if err := json.Unmarshal(byteData, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal auctions: %w", err)
	}

	return response, count, nil
}

func (sr *SupabaseRepo) SearchAuctions(ctx context.Context, query string, limit, offset int) ([]*AuctionResponse, int64, error) {
	params := map[string]any{
		"query_text": query,
		"p_limit":    limit,
		"p_offset":   offset,
	}

	byteData, _, err := sr.supabase.From("rpc/search_auctions").Insert(params, false, "", "", "exact").Execute()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search auctions: %w", err)
	}

	if len(byteData) == 0 || string(byteData) == "[]" || string(byteData) == "null" {
		return nil, 0, constants.ErrNoData
	}

	type searchResult struct {
		AuctionData json.RawMessage `json:"auction_data"`
		TotalCount  int64           `json:"total_count"`
	}

	var results []searchResult
	if err := json.Unmarshal(byteData, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	if len(results) == 0 {
		return nil, 0, constants.ErrNoData
	}

	var finalResponse []*AuctionResponse
	for _, res := range results {
		var auctionResp AuctionResponse
		if err := json.Unmarshal(res.AuctionData, &auctionResp); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal auction data: %w", err)
		}
		finalResponse = append(finalResponse, &auctionResp)
	}

	return finalResponse, results[0].TotalCount, nil
}

func (sr *SupabaseRepo) FilterAuctions(ctx context.Context, filter AuctionFilter, limit, offset int) ([]*AuctionResponse, int64, error) {
	params := map[string]any{
		"p_category":  filter.Category,
		"p_min_price": filter.MinPrice,
		"p_max_price": filter.MaxPrice,
		"p_status":    filter.Status,
		"p_sort_by":   filter.SortBy,
		"p_limit":     limit,
		"p_offset":    offset,
	}

	byteData, _, err := sr.supabase.From("rpc/filter_auctions").Insert(params, false, "", "", "exact").Execute()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to filter auctions: %w", err)
	}

	if len(byteData) == 0 || string(byteData) == "[]" || string(byteData) == "null" {
		return nil, 0, constants.ErrNoData
	}

	type searchResult struct {
		AuctionData json.RawMessage `json:"auction_data"`
		TotalCount  int64           `json:"total_count"`
	}

	var results []searchResult
	if err := json.Unmarshal(byteData, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal filter results: %w", err)
	}

	if len(results) == 0 {
		return nil, 0, constants.ErrNoData
	}

	var finalResponse []*AuctionResponse
	for _, res := range results {
		var auctionResp AuctionResponse
		if err := json.Unmarshal(res.AuctionData, &auctionResp); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal auction data: %w", err)
		}
		finalResponse = append(finalResponse, &auctionResp)
	}

	return finalResponse, results[0].TotalCount, nil
}

func (sr *SupabaseRepo) UpdateAuctionStatuses(ctx context.Context) (map[string]any, error) {
	// Call the RPC without any arguments
	res, _, err := sr.supabase.From("rpc/update_auction_statuses").Insert(map[string]any{}, false, "", "", "exact").Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to call update_auction_statuses rpc: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(res), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rpc result: %w", err)
	}

	return result, nil
}

func (sr *SupabaseRepo) Recommendation(ctx context.Context, category string, currentID string, limit, offset int) ([]*AuctionResponse, int64, error) {
	query := sr.supabase.From(string(constants.AuctionTable)).
		Select("*, products!inner(*)", "exact", false).
		Ilike("products.category", category)

	if currentID != "" {
		query = query.Neq("id", currentID)
	}

	byteData, count, err := query.Range(offset, offset+limit-1, "").Execute()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to load items: %w", err)
	}

	if count == 0 {
		return nil, 0, constants.ErrNoData
	}

	var data []*AuctionResponse
	if err := json.Unmarshal(byteData, &data); err != nil {
		return nil, 0, fmt.Errorf("failed to parse byte to json: %w", err)
	}

	return data, count, nil
}
