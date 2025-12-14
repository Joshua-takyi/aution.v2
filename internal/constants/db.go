package constants

type DbConstants string

const (
	DbName                 DbConstants = "auction"
	ProductCollection      DbConstants = "products"
	UserCollection         DbConstants = "users"
	ProfileCollection      DbConstants = "profiles"
	RefreshTokenCollection DbConstants = "refresh_tokens"
	VerificationCollection DbConstants = "verifications"
)
