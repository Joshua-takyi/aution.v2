package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/helpers"
	"github.com/joshua-takyi/auction/internal/models"
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
			"id":        user.ID,
			"email":     user.Email,
			"createdAt": user.CreatedAt,
		}

		logger.Info("User created successfully", map[string]interface{}{
			"userId": user.ID,
			"email":  user.Email,
		})

		utils.Created(c, constants.MsgUserCreated, userResponse)
	}
}

// AuthenticateUserHandler logs in a user via Supabase
func AuthenticateUserHandler(s *service.UserService, logger *utils.Logger, isProduction bool) gin.HandlerFunc {
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

		// Call service which calls Supabase SignIn
		tokenResponse, err := s.AuthenticateUser(ctx, loginRequest.Email, loginRequest.Password)
		if err != nil {
			// Supabase errors might need better mapping, but generic "invalid credentials" is safe
			logger.Warn("Login failed", map[string]interface{}{
				"email": loginRequest.Email,
				"error": err.Error(),
			})
			utils.BadRequest(c, err.Error(), "Login failed")
			return
		}

		csrf, err := helpers.GenerateCsrfToken()
		if err != nil {
			logger.Error("Failed to generate CSRF token", err, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			utils.InternalServerError(c, "Failed to generate CSRF token", "Please try again later")
			return
		}
		// save cookies
		c.SetCookie("access_token", tokenResponse.AccessToken, int(tokenResponse.ExpiresIn), "/", "", isProduction, true)
		c.SetCookie("refresh_token", tokenResponse.RefreshToken, 30*24*60*60, "/", "", isProduction, true)
		c.SetCookie("csrf_token", csrf, int(tokenResponse.ExpiresIn), "/", "", isProduction, false)

		// Return the full Supabase token response (Access Token, Refresh Token, User, etc.)
		utils.StatusOK(c, constants.MsgLoginSuccess, tokenResponse)
	}
}

func SignOut(s *service.UserService, isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err == nil {
			_ = s.SignOut(c.Request.Context(), accessToken)
		}

		c.SetCookie("access_token", "", -1, "/", "", isProduction, true)
		c.SetCookie("refresh_token", "", -1, "/", "", isProduction, true)
		c.SetCookie("csrf_token", "", -1, "/", "", isProduction, false)

		utils.StatusOK(c, constants.MsgLogoutSuccess, nil)
	}
}

func RefreshToken(s *service.UserService, isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		refresh_token, err := c.Cookie("refresh_token")
		if err != nil {
			utils.BadRequest(c, "Refresh token is required", "Please try again later")
			return
		}

		tokenResponse, err := s.RefreshToken(c.Request.Context(), refresh_token)
		if err != nil {
			utils.BadRequest(c, err.Error(), "Please try again later")
			return
		}

		csrf, err := helpers.GenerateCsrfToken()
		if err != nil {
			utils.InternalServerError(c, "Failed to generate CSRF token", "Please try again later")
			return
		}
		// set new TOkens
		c.SetCookie("access_token", tokenResponse.AccessToken, int(tokenResponse.ExpiresIn), "/", "", isProduction, true)
		c.SetCookie("refresh_token", tokenResponse.RefreshToken, 30*24*60*60, "/", "", isProduction, true)
		c.SetCookie("csrf_token", csrf, int(tokenResponse.ExpiresIn), "/", "", isProduction, false)
		utils.StatusOK(c, constants.MsgLoginSuccess, tokenResponse)
	}
}

func CreateProfileDataHandler(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exist := c.Get("user")
		if !exist {
			utils.Unauthorized(c, "unauthenticated user", "error")
			return
		}
		claims, ok := user.(*models.User)
		if !ok {
			utils.Unauthorized(c, "unauthorized access", "error")
			return
		}

		var requestBody models.Profile

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			utils.BadRequest(c, "invalid request body", "error")
			return
		}

		access_token, err := c.Cookie("access_token")
		if err != nil {
			token := c.GetHeader("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				access_token = token[7:]
			} else {
				access_token = token
			}
		}
		if !claims.IsOwner(claims.ID) {
			utils.Unauthorized(c, "unauthorized access", "error")
			return
		}

		res, err := s.CreateProfileData(c.Request.Context(), requestBody, claims.ID, access_token)
		if err != nil {
			utils.BadRequest(c, "failed to create profile data", err.Error())
			return
		}
		utils.StatusOK(c, "profile created successfully", res)
	}

}
