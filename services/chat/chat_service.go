// services/chat/chat_service.go
package chat

import (
	"log"

	"github.com/gorilla/websocket"
)

// Global instance of the chat manager
var manager *ChatManager

// Chat_init initializes the chat service
func Chat_init() {

	// Create chat manager
	manager = NewChatManager()
	// Load rooms from persistent storage

	// Start chat manager in a goroutine
	go manager.Start()

	log.Println("Chat service running...")
}

// GetRoom provides access to the GetRoom functionality of the chat manager
func GetRoom(roomID string) (*Room, bool) {
	return manager.GetRoom(roomID)
}

// CreateRoom creates a new chat room
func CreateRoom(name string, creatorID string, isPrivate bool) *Room {
	return manager.CreateRoom(name, creatorID, isPrivate)
}

// JoinRoom adds a user to a room if they are authorized
func JoinRoom(roomID, userID string) bool {
	return manager.JoinRoom(roomID, userID)
}

// LeaveRoom removes a user from a room's active members
func LeaveRoom(roomID, userID string) bool {
	return manager.LeaveRoom(roomID, userID)
}

// AddAuthorizedMember adds a user to the authorized members list
func AddAuthorizedMember(roomID, userID, requestedByID string) bool {
	return manager.AddAuthorizedMember(roomID, userID, requestedByID)
}

// RemoveAuthorizedMember removes a user from the authorized members list
func RemoveAuthorizedMember(roomID, userID, requestedByID string) bool {
	return manager.RemoveAuthorizedMember(roomID, userID, requestedByID)
}

// GetManager returns the chat manager instance
// This can be useful for direct access when needed
func GetManager() *ChatManager {
	return manager
}

// HandleWebSocketConnection creates a new client and sets up the connection
func HandleWebSocketConnection(userID, userName string, conn *websocket.Conn) {
	// Create a new client
	client := NewClient(userID, userName, conn, manager)

	// Register the client with the manager
	manager.Register <- client

	// Start the client's read and write pumps
	go client.WritePump()
	go client.ReadPump()
}
