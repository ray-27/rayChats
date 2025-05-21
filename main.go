// main.go
package main

import (
	"fmt"
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

	// Start gRPC server in a goroutine

	println("Server started....")
	// Start HTTP server
	server.Router.Run(fmt.Sprintf("0.0.0.0:%s", server.Port))
}
