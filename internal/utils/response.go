package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standardized API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse sends a standardized success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends a standardized error response
func ErrorResponse(c *gin.Context, statusCode int, code, message, details string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func StatusOK(c *gin.Context, message string, data any) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// Common response helpers
func BadRequest(c *gin.Context, message, details string) {
	ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

func Unauthorized(c *gin.Context, message, details string) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message, details)
}

func Forbidden(c *gin.Context, message, details string) {
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", message, details)
}

func NotFound(c *gin.Context, message, details string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message, details)
}

func Conflict(c *gin.Context, message, details string) {
	ErrorResponse(c, http.StatusConflict, "CONFLICT", message, details)
}

func InternalServerError(c *gin.Context, message, details string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, details)
}

func Created(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

func OK(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       interface{}     `json:"data"`
	Pagination *PaginationMeta `json:"pagination"`
}

// PaginatedOK sends a paginated success response
func PaginatedOK(c *gin.Context, message string, data interface{}, pagination *PaginationMeta) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}
