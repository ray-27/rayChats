// middleware/auth.go
package auth

import (
	"fmt"
	"net/http"
	db "raychat/database"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// AuthRequired ensures the user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")

		// Check if the header exists and has the correct format
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract the token from the header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Return the secret key used to sign the token
			return []byte("your-secret-key"), nil // Use your actual secret key
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Check if the token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Extract the user UUID from the claims
		userUUID, ok := claims["user_id"].(string)
		if !ok || userUUID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		// Set the user UUID in the context for handlers to use
		c.Set("userUUID", userUUID)
		c.Next()
	}
}

func ValidateToken(c *gin.Context) {
	// The AuthRequired middleware has already validated the token
	// If we reach this handler, the token is valid

	// Get the user UUID from the context (set by AuthRequired middleware)
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User UUID not found in context"})
		return
	}

	// Convert to string
	userUUIDStr, ok := userUUID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user UUID format"})
		return
	}

	// Get user data from your database/store using the UUID
	userData, err := db.Store.GetUserByUUID(userUUIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
		return
	}

	// Return validation status along with user data
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user":  userData,
	})
}

func GetUserInfo(c *gin.Context) {

}
