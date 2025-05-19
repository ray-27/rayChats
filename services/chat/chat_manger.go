package chat

import (
	"log"
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
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mutex      sync.RWMutex
}

// NewChatManager creates a new chat manager
func NewChatManager() *ChatManager {
	return &ChatManager{
		Rooms:      make(map[string]*Room),
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (cm *ChatManager) Start() {
	log.Println("Chat manager started")

	for {
		select {
		case client := <-cm.Register: //client is recieved from the Register channel
			log.Printf("Registering client: %s", client.ID)
			cm.mutex.Lock()
			cm.Clients[client.ID] = client //adds to the Client map
			cm.mutex.Unlock()

		case client := <-cm.Unregister:
			log.Printf("Unregistering client: %s", client.ID)
			cm.mutex.Lock()
			if _, exists := cm.Clients[client.ID]; exists { //if client exists then remove it from all the rooms

				//Remove client from all rooms
				for roomID := range client.Rooms {
					if room, exists := cm.Rooms[roomID]; exists {
						delete(room.ActiveMembers, client.UserID) //delete from the room
						log.Printf("Removed Client %s, from room %s", client.UserID, roomID)
					}
				}

				//Close send channel and delete Client
				close(client.Send)
				delete(cm.Clients, client.ID)
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
				for userID := range room.ActiveMembers {
					//Find all the clients for this user
					for _, client := range cm.Clients {
						if client.UserID == userID {
							select {
							case client.Send <- message:
								// Message sent successfuly, the message is sent through a channel that is read by the WritePump() function
							default:
								// Client's buffer is full, close connection
								close(client.Send)
								delete(cm.Clients, client.ID)
							}
						}
					}
				}
				log.Printf("Message broadcast complete")
			}
		}
	}
}

// CreateRoom creates a new chat room
func (cm *ChatManager) CreateRoom(name string, creatorID string, isPrivate bool) *Room {
	room := NewRoom(name, creatorID, isPrivate) // create a new room

	cm.mutex.Lock()
	cm.Rooms[room.ID] = room // add the room to the ChatManager
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

// JoinRoom adds a user to a room
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
		if _, authorized := room.AuthorizedMembers[userID]; !authorized {
			log.Printf("User %s attempted to join private room %s but is not authorized",
				userID, roomID)
			return false
		}
	}

	// Add to active members
	room.ActiveMembers[userID] = true

	log.Printf("User %s joined room %s, room now has %d active members",
		userID, roomID, len(room.ActiveMembers))

	return true
}

// AddAuthorizedMember adds a user to the authorized members list
func (cm *ChatManager) AddAuthorizedMember(roomID, userID, requestedByID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	room, exists := cm.Rooms[roomID]
	if !exists {
		return false
	}

	// Check if the requesting user has permission (owner or admin)
	if room.CreatorID != requestedByID && !room.Admins[requestedByID] {
		log.Printf("User %s attempted to add member to room %s but lacks permission",
			requestedByID, roomID)
		return false
	}

	room.AuthorizedMembers[userID] = true
	log.Printf("User %s added to authorized members of room %s by %s",
		userID, roomID, requestedByID)

	return true
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
