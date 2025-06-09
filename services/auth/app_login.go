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
		Token      string `json"token" binding:"required"`
		AppVersion string `json"app_version" binding:"required"`
		Provider   string `json"provider" binding:"required"`
		TimeStamp  string `json"timestamp" binding:"required"`
	}

	c.JSON(200, gin.H{
		"token":   "qweqaddsadas",
		"success": "true",
		"user_id": "nigga",
	})
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

	c.JSON(http.StatusOK, gin.H{
		"ID":    authResponse.GetId(),
		"Name":  authResponse.GetName(),
		"Email": authResponse.GetEmail(),
	})

}
