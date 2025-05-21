// services/chat/chat_service.go
package chat

import (
	"log"
	"raychat/models"

	"github.com/gorilla/websocket"
)

// Global instance of the chat manager
var manager *ChatManager

// Chat_init initializes the chat service
func Chat_init() {
	// // Initialize Valkey store
	// store := db.NewValkeyChatStore("localhost:6379", "", 0)
	// db.Store = store

	// Create chat manager
	manager = NewChatManager()
	// Load rooms from persistent storage
	LoadRoomsFromStorage()

	// Start chat manager in a goroutine
	go manager.Start()

	log.Println("Chat service running...")
}

// LoadRoomsFromStorage loads all rooms from persistent storage
func LoadRoomsFromStorage() {
	// Get the chat manager
	cm := GetManager()

	// Get all room IDs from Valkey
	roomIDs, err := cm.Store.GetAllRoomIDs()
	if err != nil {
		log.Printf("Error loading room IDs from storage: %v", err)
		return
	}

	log.Printf("Loading %d rooms from persistent storage", len(roomIDs))

	// Load each room
	for _, roomID := range roomIDs {
		roomDetails, err := cm.Store.GetRoomDetails(roomID)
		if err != nil {
			log.Printf("Error loading room %s: %v", roomID, err)
			continue
		}

		// Parse room details
		isPrivate := roomDetails["is_private"] == "1"

		// Create room object
		room := &models.Room{
			ID:                roomID,
			Name:              roomDetails["name"],
			CreatorID:         roomDetails["creator_id"],
			IsPrivate:         isPrivate,
			AuthorizedMembers: make(map[string]bool),
			ActiveMembers:     make(map[string]bool),
			Admins:            make(map[string]bool),
		}

		// Get authorized members
		authMembers, err := cm.Store.GetRoomAuthorizedMembers(roomID)
		if err == nil {
			for _, memberID := range authMembers {
				room.AuthorizedMembers[memberID] = true
			}
		}

		// Get admins
		admins, err := cm.Store.GetRoomAdmins(roomID)
		if err == nil {
			for _, adminID := range admins {
				room.Admins[adminID] = true
			}
		}

		// Add creator as admin if not already
		room.Admins[room.CreatorID] = true

		// Add room to manager
		cm.mutex.Lock()
		cm.Rooms[roomID] = room
		cm.mutex.Unlock()

		log.Printf("Loaded room: %s, name: %s", roomID, room.Name)
	}
}

// GetRoom provides access to the GetRoom functionality of the chat manager
func GetRoom(roomID string) (*models.Room, bool) {
	return manager.GetRoom(roomID)
}

// CreateRoom creates a new chat room
func CreateRoom(name string, creatorID string, isPrivate bool) *models.Room {
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
