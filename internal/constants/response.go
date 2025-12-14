package constants

import "errors"

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Common error messages
var (
	ErrInvalidInput       = errors.New("invalid user input")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInternalServer     = errors.New("internal server error")
	ErrEmptyFields        = errors.New("required fields cannot be empty")
	ErrWeakPassword       = errors.New("password does not meet security requirements")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("forbidden")
	ErrUserNotVerified    = errors.New("user not verified")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidToken       = errors.New("invalid verification token")
	ErrTokenExpired       = errors.New("verification token expired")

	ErrNoClient = errors.New("no database client available")
)

// Success messages
const (
	MsgUserCreated      = "User created successfully"
	MsgUserUpdated      = "User updated successfully"
	MsgUserVerified     = "User verified successfully"
	MsgUserDeleted      = "User deleted successfully"
	MsgLoginSuccess     = "Login successful"
	MsgLogoutSuccess    = "Logout successful"
	MsgPasswordChanged  = "Password changed successfully"
	MsgEmailVerified    = "Email verified successfully"
	MsgOperationSuccess = "Operation completed successfully"
)

// Password requirements message
const PasswordRequirements = "Password must be at least 8 characters and contain uppercase, lowercase, digit, and special character (@$!%*#?&)"
