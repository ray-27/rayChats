package auth

import (
	"context"
	"net/http"
	"raychat/config"
	"raychat/proto/pb"
	"time"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID    string
	Name  string
	Email string
}

func Login(c *gin.Context) {

	var req struct {
		Token      string `json:"token" binding:"required"`
		AppVersion string `json:"app_version" binding:"required"`
		Provider   string `json:"provider" binding:"required"`
		TimeStamp  int64  `json:"timestamp" binding:"required"`
	}

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		println("Wrong binding...")
		c.JSON(http.StatusBadRequest, gin.H{"error": "nigga"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	authResponse, err := config.Client.AuthClient.GetUserData(ctx, &pb.Token{Token: req.Token})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get user details"})
		return
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"ID":    authResponse.GetId(),
	// 	"Name":  authResponse.GetName(),
	// 	"Email": authResponse.GetEmail(),
	// })
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		
		"jwt_token": "jwtToken123131241112",
		"user": gin.H{
			"user_id":       authResponse.GetId(),
			"id":            authResponse.GetId(),
			"name":          authResponse.GetName(),
			"email":         authResponse.GetEmail(),
			"profile_image": "", // Get from authResponse if available
			"created_at":    time.Now().Unix() * 1000,
			"updated_at":    time.Now().Unix() * 1000,
		},
		"token_expires_at": time.Now().Add(24*time.Hour).Unix() * 1000,
		"refresh_token":    "optional_refresh_token_here",

		"timestamp": time.Now().Unix() * 1000,
	})

}
