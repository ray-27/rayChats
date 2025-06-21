package chat

import (
	"fmt"
	"log"
	"raychat/models"
	"sync"
)

// ChatManager handles all chat operations
/*
The ChatManager is designed to be the central coordinator for your entire chat system.
It maintains:
- A map of all rooms (Rooms)
- A map of all connected clients (Clients)
- Channels for communication between different parts of the system
*/
type ChatManager struct {
	Rooms      map[string]*Room
	Clients    map[string]*Client //client are the users which are online
	Broadcast  chan *models.Message
	Register   chan *Client
	Unregister chan *Client
	mutex      sync.RWMutex
	// Store      *db.ValkeyChatStore
}

// NewChatManager creates a new chat manager
func NewChatManager() *ChatManager {
	cm := &ChatManager{
		Rooms:      make(map[string]*Room),
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *models.Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	// Load rooms using database package
	if err := cm.loadAllRooms(); err != nil {
		log.Printf("Error loading rooms: %v", err)
	}

	return cm
}

func (cm *ChatManager) Start() {
	log.Println("Chat manager started")

	for {
		select {
		case client := <-cm.Register: //client is recieved from the Register channel
			log.Printf("Registering client: %s", client.UserID)
			cm.mutex.Lock()
			cm.Clients[client.UserID] = client //adds to the Client map
			cm.mutex.Unlock()

		case client := <-cm.Unregister:
			log.Printf("Unregistering client: %s", client.UserID)
			cm.mutex.Lock()
			if _, exists := cm.Clients[client.UserID]; exists { //if client exists then remove it from all the rooms

				//Remove client from all rooms
				for roomID := range client.Rooms {
					if room, exists := cm.Rooms[roomID]; exists {
						delete(room.ActiveMembers, client.UserID) //delete from the room
						// Update Valkey
						// if err := cm.Store.SetUserInactive(client.UserID, roomID); err != nil {
						// 	log.Printf("Error marking user inactive: %v", err)
						// }
						log.Printf("Removed Client %s, from room %s", client.UserID, roomID)
					}
				}

				//Close send channel and delete Client
				close(client.Send)
				delete(cm.Clients, client.UserID)
			}
			cm.mutex.Unlock()

		case message := <-cm.Broadcast:
			log.Printf("Recieved bradcast message for room: %s", message.RoomID)
			cm.mutex.RLock()
			room, exists := cm.Rooms[message.RoomID]
			cm.mutex.RUnlock()

			if exists {
				log.Printf("Broadcasting message to room %s with %d members", message.RoomID, len(room.ActiveMembers))

				//Send Message to all the members in the room
				for userID := range room.AuthorizedMembers {

					//check if the user is currently online
					if client, isActive := room.ActiveMembers[userID]; isActive {

						//Find all the clients for this user
						select {
						case client.Send <- message:

						default:
							// Client's buffer is full, clean up
							cm.mutex.Lock()
							delete(room.ActiveMembers, userID)
							close(client.Send)
							delete(cm.Clients, client.UserID)
							cm.mutex.Unlock()
						}

					} else {
						//User offline
						//make a fucntion to send the message later
					}
				}
				log.Printf("Message broadcast complete")
			}
		}
	}
}

func (cm *ChatManager) loadAllRooms() error {
	rooms, err := LoadAllRoomsWithMembersFromValkey()
	if err != nil {
		return err
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, room := range rooms {
		cm.Rooms[room.ID] = room
	}

	log.Printf("Loaded %d rooms from database", len(rooms))
	return nil
}

// CreateRoom creates a new chat room
func (cm *ChatManager) CreateRoom(roomId, name string, creatorID string, isPrivate bool) *Room {
	//This function will create a room and add it to the Chat Manager

	room := NewRoom(roomId, name, creatorID, isPrivate)

	// Initialize maps
	room.AuthorizedMembers = make(map[string]bool)
	room.Admins = make(map[string]bool)
	room.ActiveMembers = make(map[string]*Client)

	// Creator is both admin and authorized member
	room.AuthorizedMembers[creatorID] = true
	room.Admins[creatorID] = true

	cm.mutex.Lock()
	cm.Rooms[room.ID] = room
	cm.mutex.Unlock()

	log.Printf("Created room: %s, creator: %s", room.ID, creatorID)
	return room
}

// GetRoom returns a room by ID
func (cm *ChatManager) GetRoom(roomID string) (*Room, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	room, exists := cm.Rooms[roomID]
	return room, exists
}

// JoinRoom adds a user to a room if they are authorized
func (cm *ChatManager) JoinRoom(roomID, userID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	room, exists := cm.Rooms[roomID]
	if !exists {
		return false
	}

	// Check if the room is private and if the user is authorized
	if room.IsPrivate {
		if _, ok := room.AuthorizedMembers[userID]; !ok {
			log.Printf("User %s not authorized for private room %s", userID, roomID)
			return false
		}
	} else {
		// Add to authorized members if not already (for public rooms)
		if _, ok := room.AuthorizedMembers[userID]; !ok {
			room.AuthorizedMembers[userID] = true
		}
	}

	userClient, exists := cm.Clients[userID]
	if !exists {
		log.Println("No active client found for user %s", userID)
		return false
	}
	// Add to active members
	room.ActiveMembers[userID] = userClient

	log.Printf("User %s joined room %s, room now has %d active members",
		userID, roomID, len(room.ActiveMembers))

	return true
}

func (cm *ChatManager) AddAuthorizedMemberUnrestricted(roomID, userID, requestedByID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	room, exists := cm.Rooms[roomID]
	if !exists {
		return fmt.Errorf("room does not exists")
	}

	room.AuthorizedMembers[userID] = true
	log.Printf("User %s added to authorized members of room %s by %s",
		userID, roomID, requestedByID)

	return nil
}

// AddAuthorizedMember adds a user to the authorized members list
func (cm *ChatManager) AddAuthorizedMember(roomID, userID, requestedByID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	room, exists := cm.Rooms[roomID]
	if !exists {
		return fmt.Errorf("room does not exists")
	}

	// Check if the requesting user has permission (owner or admin)
	// if room.CreatorID != requestedByID && !room.Admins[requestedByID] {
	if !room.Admins[requestedByID] { // only check if the admin has sent the request
		log.Printf("User %s attempted to add member to room %s but lacks permission",
			requestedByID, roomID)
		return fmt.Errorf("Unauthorized to get added to the room")
	}

	room.AuthorizedMembers[userID] = true
	log.Printf("User %s added to authorized members of room %s by %s",
		userID, roomID, requestedByID)

	return nil
}

// RemoveAuthorizedMember removes a user from the authorized members list
func (cm *ChatManager) RemoveAuthorizedMember(roomID, userID, requestedByID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	room, exists := cm.Rooms[roomID]
	if !exists {
		return false
	}

	// Check if the requesting user has permission (owner or admin)
	if room.CreatorID != requestedByID && !room.Admins[requestedByID] {
		return false
	}

	// Cannot remove the creator
	if userID == room.CreatorID {
		return false
	}

	delete(room.AuthorizedMembers, userID)

	// Also remove from active members if they're currently active
	delete(room.ActiveMembers, userID)

	log.Printf("User %s removed from authorized members of room %s by %s",
		userID, roomID, requestedByID)

	return true
}

// LeaveRoom removes a user from a room's active members
func (cm *ChatManager) LeaveRoom(roomID, userID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	room, exists := cm.Rooms[roomID]
	if !exists {
		return false
	}

	// Check if user is actually in the room
	if _, isActive := room.ActiveMembers[userID]; !isActive {
		return false
	}

	// Remove from active members
	delete(room.ActiveMembers, userID)

	// Update client's room list if they're online
	for _, client := range cm.Clients {
		if client.UserID == userID {
			delete(client.Rooms, roomID)
			break
		}
	}

	log.Printf("User %s left room %s, room now has %d active members",
		userID, roomID, len(room.ActiveMembers))

	// Optional: Check if room is empty and perform cleanup if needed
	if len(room.ActiveMembers) == 0 && !room.IsPrivate {
		log.Printf("Room %s is now empty", roomID)
		// You could add logic here to delete temporary rooms if desired
	}

	return true
}
