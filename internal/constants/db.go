package constants

type DbConstants string

const (
	DbName       DbConstants = "auction"
	ProductTable DbConstants = "products"
	UserTable    DbConstants = "users"
	ProfileTable DbConstants = "profiles"
	AuctionTable DbConstants = "auctions"
	BidTable     DbConstants = "bids"
)
