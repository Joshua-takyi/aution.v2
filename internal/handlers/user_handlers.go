package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/jwt"
	"github.com/joshua-takyi/auction/internal/service"
	"github.com/joshua-takyi/auction/internal/utils"
)

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateUserHandler handles user creation requests
func CreateUserHandler(s *service.UserService, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		utils.LogRequest(c, logger)

		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid request body", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.BadRequest(c, "Invalid request body", err.Error())
			return
		}

		// Create user through service
		user, err := s.CreateUser(ctx, req.Email, req.Password)
		if err != nil {
			// Handle different error types
			switch {
			case errors.Is(err, constants.ErrEmptyFields):
				logger.Warn("Empty fields provided", map[string]interface{}{
					"email": req.Email,
				})
				utils.BadRequest(c, err.Error(), constants.PasswordRequirements)
				return

			case errors.Is(err, constants.ErrWeakPassword):
				logger.Warn("Weak password provided", map[string]interface{}{
					"email": req.Email,
				})
				utils.BadRequest(c, err.Error(), constants.PasswordRequirements)
				return

			case errors.Is(err, constants.ErrUserAlreadyExists):
				logger.Warn("User already exists", map[string]interface{}{
					"email": req.Email,
				})
				utils.Conflict(c, err.Error(), "A user with this email already exists")
				return

			default:
				logger.Error("Failed to create user", err, map[string]interface{}{
					"email": req.Email,
				})
				utils.InternalServerError(c, "Failed to create user", "Please try again later")
				return
			}
		}

		// Prepare response (exclude password)
		userResponse := map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"isVerified": user.IsVerified,
			"createdAt":  user.CreatedAt,
		}

		logger.Info("User created successfully", map[string]interface{}{
			"userId": user.ID,
			"email":  user.Email,
		})

		utils.Created(c, constants.MsgUserCreated, userResponse)
	}
}

func VerifyUserHandler(s *service.UserService, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		email := c.Query("email")
		ctx := c.Request.Context()
		utils.LogRequest(c, logger)

		if token == "" || email == "" {
			logger.Warn("Empty token or email provided", map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.BadRequest(c, "Empty token or email provided", "Please provide a valid token and email")
			return
		}

		if err := s.VerifyUser(ctx, token, email); err != nil {
			logger.Error("Failed to verify user", err, map[string]interface{}{
				"token": token,
			})
			utils.InternalServerError(c, "Failed to verify user", "Please try again later")
			return
		}

		utils.StatusOK(c, constants.MsgUserVerified, nil)
	}
}

func AuthenticateUserHandler(s *service.UserService, logger *utils.Logger, jwtManager *jwt.JWTManager, isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		utils.LogRequest(c, logger)

		var loginRequest CreateUserRequest
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			logger.Error("Failed to bind JSON", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.BadRequest(c, "Failed to bind JSON", "Please try again later")
			return
		}

		user, err := s.AuthenticateUser(ctx, loginRequest.Email, loginRequest.Password)
		if err != nil {
			switch {
			case errors.Is(err, constants.ErrUserNotVerified):
				logger.Warn("User not verified", map[string]interface{}{
					"email": loginRequest.Email,
				})
				utils.BadRequest(c, err.Error(), "Please verify your email address")
				return
			case errors.Is(err, constants.ErrInvalidCredentials):
				logger.Warn("Invalid credentials", map[string]interface{}{
					"email": loginRequest.Email,
				})
				utils.BadRequest(c, err.Error(), "Invalid email or password")
				return
			default:
				logger.Error("Failed to authenticate user", err, map[string]interface{}{
					"path": c.Request.URL.Path,
				})
				utils.InternalServerError(c, "Failed to authenticate user", "Please try again later")
				return
			}
		}

		profile, err := s.GetProfileByUserId(ctx, user.ID)

		if err != nil {
			logger.Error("Failed to get profile", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.InternalServerError(c, "Failed to get profile", "Please try again later")
			return
		}

		userAuth := jwt.UserAuth{
			ID:    user.ID.Hex(),
			Email: user.Email,
			Role:  profile.Role,
		}

		token, err := jwtManager.GenerateToken(userAuth)
		if err != nil {
			logger.Error("Failed to generate token", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.InternalServerError(c, "Failed to generate token", "Please try again later")
			return
		}

		csrf, err := jwt.GenerateCsrfToken()
		if err != nil {
			logger.Error("Failed to generate CSRF token", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.InternalServerError(c, "Failed to generate CSRF token", "Please try again later")
			return
		}

		refresh, err := jwtManager.GenerateRefreshToken(userAuth, 3600*24*30)
		if err != nil {
			logger.Error("Failed to generate refresh token", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.InternalServerError(c, "Failed to generate refresh token", "Please try again later")
			return
		}

		if isProduction {
			c.SetCookie("access_token", token,
				3600*24*7,
				"/",
				"",
				true,
				true,
			)

			c.SetCookie("csrf_token", csrf,
				3600*24*7,
				"/",
				"",
				true,
				true,
			)
			c.SetCookie("refresh_token", refresh,
				3600*24*30,
				"/",
				"",
				true,
				true,
			)
		} else {
			c.SetCookie("access_token", token,
				3600*24*7,
				"/",
				"localhost",
				false,
				true,
			)

			c.SetCookie("csrf_token", csrf,
				3600*24*7,
				"/",
				"localhost",
				false,
				true,
			)

			c.SetCookie("refresh_token", refresh,
				3600*24*30,
				"/",
				"localhost",
				false,
				true,
			)
		}

		userResponse := map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"isVerified": user.IsVerified,
			"createdAt":  user.CreatedAt,
		}
		utils.StatusOK(c, constants.MsgLoginSuccess, userResponse)
	}
}
