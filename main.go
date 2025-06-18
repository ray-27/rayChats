// main.go
package main

import (
	"fmt"
	"log"
	"raychat/config"
	db "raychat/database"
	"raychat/handler"
	"raychat/services/chat"
)

func main() {
	// Initialize configuration
	server := config.Server{}
	err := server.IntiServer()
	if err != nil {
		fmt.Println("There is an error in initializing the server")
		return
	}
	println("Server init done")

	db.DB_init()
	println("Database init done.")

	chat.Chat_init() 

	// Set up HTTP routes
	handler.Handles(server.Router)

	config.Client, err = config.NewGrpcManager("localhost:8080")
	// config.Client, err = config.NewGrpcManager("api.resnight.tech")
	if err != nil {
		log.Fatalf("Failed to create gprc client manage: %v", err)
	}
	defer config.Client.Close()

	println("gRPC server running...")

	println("Server started....")
	// Start HTTP server
	server.Router.Run(fmt.Sprintf("0.0.0.0:%s", server.Port))
}
