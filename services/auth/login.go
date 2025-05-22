package auth

import (
	"net/http"
	db "raychat/database"
	"raychat/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func SignupCLI(c *gin.Context) {

	var request struct {
		Email       string `json:"email" binding:"required"`
		Password    string `json:"password" binding:"required"`
		UserName    string `json:"username" binding:"required"`
		PhoneNumber string `json:"phoneno" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Email doesn't exist, create new user
	newUser := &models.UserCred{
		UUID:        uuid.New().String(),
		Username:    request.UserName,
		Email:       request.Email,
		Password:    request.Password, // In a real app, hash this password
		PhoneNumber: request.PhoneNumber,
		CreatedAt:   time.Now(),
	}

	// Save the new user
	err := db.Store.SaveUserCredentials(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token for the new user
	token, err := generateJWT(newUser.UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Store the token in Valkey with expiration
	tokenKey := "token:" + newUser.UUID
	err = db.Store.Client.Set(db.Store.Ctx, tokenKey, token, 10*24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"token":   token,
		"user":    newUser,
	})
}

func LoginCLI(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if the email already exists
	existingUser, err := db.Store.GetUserByEmail(request.Email)
	if err == nil && existingUser != nil {
		// User exists, verify password
		// In a real app, use bcrypt.CompareHashAndPassword for password comparison
		if existingUser.Password != request.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
			return
		}

		// Password is correct, generate JWT token
		token, err := generateJWT(existingUser.UUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Store the token in Valkey with expiration
		tokenKey := "token:" + existingUser.UUID
		err = db.Store.Client.Set(db.Store.Ctx, tokenKey, token, 10*24*time.Hour).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   token,
			"user":    existingUser,
			"exists":  "true",
		})
		return
	} else if err != redis.Nil {
		// Some other error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else if existingUser == nil {
		c.JSON(http.StatusOK, gin.H{"message": "acount does not exists, create account", "exists": "false"})
	}

	// // Email doesn't exist, create new user
	// newUser := &models.UserCred{
	// 	UUID:      uuid.New().String(),
	// 	Email:     request.Email,
	// 	Password:  request.Password, // In a real app, hash this password
	// 	CreatedAt: time.Now(),
	// }

	// // Save the new user
	// err = db.Store.SaveUserCredentials(newUser)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
	// 	return
	// }

	// // Generate JWT token for the new user
	// token, err := generateJWT(newUser.UUID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
	// 	return
	// }

	// // Store the token in Valkey with expiration
	// tokenKey := "token:" + newUser.UUID
	// err = db.Store.Client.Set(db.Store.Ctx, tokenKey, token, 10*24*time.Hour).Err()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
	// 	return
	// }

	// c.JSON(http.StatusCreated, gin.H{
	// 	"message": "Account created successfully",
	// 	"token":   token,
	// 	"user":    newUser,
	// })
}

// Helper function to generate JWT token
func generateJWT(userID string) (string, error) {
	// Create a new token object
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(10 * 24 * time.Hour).Unix() // 10 days expiration

	// Generate encoded token
	tokenString, err := token.SignedString([]byte("your-secret-key")) // Use a secure secret key in production
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
