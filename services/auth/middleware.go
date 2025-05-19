// middleware/auth.go
package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired ensures the user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get UUID from header first
		userUUID := c.GetHeader("X-User-UUID")

		// If not in header, try to get from query parameter
		if userUUID == "" {
			userUUID = c.Query("uuid")
		}

		if userUUID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Set the user UUID in the context for handlers to use
		c.Set("userUUID", userUUID)
		c.Next()
	}
}
