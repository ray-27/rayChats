// routes/router.go
package handler

import (
	"raychat/services/auth"
	"raychat/services/chat"

	"github.com/gin-gonic/gin"
)

func Handles(router *gin.Engine) {
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/ping", PingHandler)
	router.Static("/static", "./static")

	// Serve the chat client
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	router.POST("cli/login", auth.LoginCLI)
	router.POST("cli/signup", auth.SignupCLI)
	router.POST("cli/otp", auth.GetOTP)
	cliGroup := router.Group("/cli")
	cliGroup.Use(auth.AuthRequired())
	{
		cliGroup.GET("/validatetoken", auth.ValidateToken)
		cliGroup.GET("/userinfo", auth.GetUserInfo)
	}

	router.POST("app/login", auth.Login)
	// app := router.Group("/app")
	// {

	// }

	chat.RegisterChatRoutes(router)
}
