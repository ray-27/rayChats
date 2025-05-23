package chat

import (
	"encoding/json"
	"log"
	"time"

	"raychat/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

//a client represents a connected chat user, (online user)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 10000
)

type Client struct {
	ID       string
	UserID   string
	UserName string
	Conn     *websocket.Conn
	Manager  *ChatManager
	Send     chan *models.Message
	Rooms    map[string]bool
}

// NewClient creates a new chat client
func NewClient(userID, userName string, conn *websocket.Conn, manager *ChatManager) *Client {
	return &Client{
		ID:       userID,
		UserID:   userID,
		UserName: userName,
		Conn:     conn,
		Manager:  manager,
		Send:     make(chan *models.Message, 256),
		Rooms:    make(map[string]bool),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Manager.Unregister <- c
		c.Conn.Close()
		log.Printf("Client disconnected: %s", c.UserID)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		log.Printf("Received raw message from client %s: %s", c.UserID, string(data))

		var msg models.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		//Set sender ID if not already set
		if msg.SenderID == "" {
			msg.SenderID = c.UserID
		}

		// Set timestamp if not already set
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().Unix()
		}

		log.Printf("Processing message: type=%s, room=%s, sender=%s, content=%s",
			msg.Type, msg.RoomID, msg.SenderID, msg.Content)

		switch msg.Type {
		case "join":
			//join the room
			if c.Manager.JoinRoom(msg.RoomID, c.UserID) {
				c.Rooms[msg.RoomID] = true

				//Notify other members
				joinMsg := &models.Message{
					ID:        uuid.New().String(),
					RoomID:    msg.RoomID,
					SenderID:  c.UserID,
					Content:   c.UserName + " joined the room",
					Type:      "system",
					Timestamp: time.Now().Unix(),
				}
				c.Manager.Broadcast <- joinMsg
			} else {
				// Failure case - send error message back to this client only
				errorMsg := &models.Message{
					ID:        uuid.New().String(),
					RoomID:    msg.RoomID,
					SenderID:  "system",
					Content:   "You are not authorized to join this room",
					Type:      "error",
					Timestamp: time.Now().Unix(),
				}

				// Marshal the message to JSON
				errorData, err := json.Marshal(errorMsg)
				if err != nil {
					log.Printf("Error marshaling error message: %v", err)
					return
				}

				// Send directly to this client's connection
				if err := c.Conn.WriteMessage(websocket.TextMessage, errorData); err != nil {
					log.Printf("Error sending error message: %v", err)
				}

				log.Printf("User %s attempted to join room %s but was unauthorized", c.UserID, msg.RoomID)
			}

		case "leave":
			if _, exists := c.Rooms[msg.RoomID]; exists {
				delete(c.Rooms, msg.RoomID)

				if room, exists := c.Manager.GetRoom(msg.RoomID); exists {
					delete(room.ActiveMembers, c.UserID)
					// Update Valkey
					if err := c.Manager.Store.SetUserInactive(c.UserID, msg.RoomID); err != nil {
						log.Printf("Error marking user inactive: %v", err)
					}
					// Notify other members
					leaveMsg := &models.Message{
						ID:        uuid.New().String(),
						RoomID:    msg.RoomID,
						SenderID:  c.UserID,
						Content:   c.UserName + " left the room",
						Type:      "system",
						Timestamp: time.Now().Unix(),
					}
					c.Manager.Broadcast <- leaveMsg
				}
			}

		case "message":

			room, exists := c.Manager.GetRoom(msg.RoomID)
			if !exists {
				// Room doesn't exist
				errorMsg := &models.Message{
					ID:        uuid.New().String(),
					RoomID:    msg.RoomID,
					SenderID:  "system",
					Content:   "Room does not exist",
					Type:      "error",
					Timestamp: time.Now().Unix(),
				}
				sendErrorToClient(c, errorMsg)
				continue
			}
			// Check if user is an active member of the room
			c.Manager.mutex.RLock()
			isActiveMember := false
			if room.ActiveMembers != nil {
				_, isActiveMember = room.ActiveMembers[c.UserID]
			}
			c.Manager.mutex.RUnlock()

			if !isActiveMember {
				// User is not an active member of the room
				errorMsg := &models.Message{
					ID:        uuid.New().String(),
					RoomID:    msg.RoomID,
					SenderID:  "system",
					Content:   "You must join the room before sending messages",
					Type:      "error",
					Timestamp: time.Now().Unix(),
				}
				sendErrorToClient(c, errorMsg)
				continue
			}
			// Regular message, broadcast to room
			c.Manager.Broadcast <- &msg
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				return
			}

			w.Write(data)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				nextMsg := <-c.Send
				data, err := json.Marshal(nextMsg)
				if err != nil {
					continue
				}
				w.Write([]byte{'\n'})
				w.Write(data)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Helper function to send error messages directly to a client
func sendErrorToClient(c *Client, msg *models.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}
