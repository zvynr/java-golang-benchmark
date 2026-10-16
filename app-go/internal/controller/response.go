package controller

import "github.com/gin-gonic/gin"

// RespondError sends a standardized JSON error response idiomatic to Go Gin
func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status": status,
		"error":  message,
		"path":   c.Request.URL.Path,
	})
}

// RespondJSON sends a successful JSON response
func RespondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}
