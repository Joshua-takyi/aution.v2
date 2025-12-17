package constants

type DbConstants string

const (
	DbName       DbConstants = "auction"
	ProductTable DbConstants = "products"
	UserTable    DbConstants = "users"
	ProfileTable DbConstants = "profiles"
	// RefreshTokenTable DbConstants = "refresh_tokens"
	// VerificationTable DbConstants = "verifications"
	AuctionTable DbConstants = "auctions"
)
