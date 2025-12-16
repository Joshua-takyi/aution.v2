package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/auction/internal/utils"
)

// RequestLogger logs all incoming requests and their responses
func RequestLogger(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Log incoming request
		logger.Info("Incoming request", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
			"agent":  c.Request.UserAgent(),
		})

		// Process request
		c.Next()

		// Log response
		duration := time.Since(startTime)
		logger.Info("Request completed", map[string]interface{}{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"duration": duration.Milliseconds(),
			"ip":       c.ClientIP(),
		})
	}
}

// ErrorHandler handles any panics and returns a proper error response
func ErrorHandler(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", nil, map[string]interface{}{
					"error": err,
					"path":  c.Request.URL.Path,
				})
				utils.InternalServerError(c, "An unexpected error occurred", "Please try again later")
			}
		}()
		c.Next()
	}
}
