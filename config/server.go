package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Server struct {
	Router *gin.Engine
	Port   string
}

func (s *Server) IntiServer() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}

	//init router
	gin.SetMode(os.Getenv("GIN_MODE"))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if PORT env is not set
	}
	router := gin.Default()
	log.Printf("Server running at port: %s", port)

	s.Router = router
	s.Port = port

	return nil
}
