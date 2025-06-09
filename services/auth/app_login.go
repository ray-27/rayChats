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
		TimeStamp  string `json:"timestamp" binding:"required"`
	}

	// c.JSON(200, gin.H{
	// 	"success": true,
	// 	"message": "Login successful",
	// 	"data": gin.H{
	// 		"jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDg4OTYxMTEsInVzZXJfaWQiOiI5NGU3NDM1Ny1jODU3LTRiMTctYWFkYS1hMzg4YWVmMzY2MDEifQ.rss3_AVik8VWLONr3tD-ylm9-zudkHpmf_4LYZBXS-Y",
	// 		"user": gin.H{
	// 			"user_id":       "12345",
	// 			"id":            "12345",
	// 			"name":          "John Doe",
	// 			"email":         "john.doe@gmail.com",
	// 			"profile_image": "https://lh3.googleusercontent.com/a/profile.jpg",
	// 			"created_at":    1704067200000,
	// 			"updated_at":    1704067200000,
	// 		},
	// 		"token_expires_at": 1704153600000,
	// 		"refresh_token":    "optional_refresh_token_here",
	// 	},
	// 	"timestamp": 1704067200000,
	// })
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	authResponse, err := config.Client.AuthClient.GetUserData(ctx, &pb.Token{Token: "token"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get user details"})
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"ID":    authResponse.GetId(),
	// 	"Name":  authResponse.GetName(),
	// 	"Email": authResponse.GetEmail(),
	// })
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": gin.H{
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
		},
		"timestamp": time.Now().Unix() * 1000,
	})

}
