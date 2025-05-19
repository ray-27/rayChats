// // handler/chat_handler.go
package handler

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"raychat/models"

// 	"github.com/gin-gonic/gin"
// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true // Allow all origins for now
// 	},
// }

// // WebSocketHandler handles WebSocket connections
// // handler/chat_handler.go
// func WebSocketHandler(c *gin.Context) {
// 	// Get user UUID from context (set by auth middleware)
// 	userUUID, exists := c.Get("userUUID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User UUID not found"})
// 		return
// 	}

// 	// Convert userUUID to string if it's not already
// 	userUUIDStr, ok := userUUID.(string)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid UUID format"})
// 		return
// 	}

// 	// Upgrade HTTP connection to WebSocket
// 	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		log.Printf("Failed to upgrade connection: %v", err)
// 		return
// 	}

// 	// Register the connection
// 	services.Registry.Register(userUUIDStr, conn)
// 	log.Printf("User %s connected", userUUIDStr)

// 	// Rest of the handler remains the same...
// 	// Handle disconnect
// 	defer func() {
// 		services.Registry.Unregister(userUUID.(string))
// 		log.Printf("User %s disconnected", userUUID)
// 	}()

// 	// Listen for messages from the client
// 	for {
// 		_, message, err := conn.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				log.Printf("WebSocket error: %v", err)
// 			}
// 			break
// 		}

// 		// Parse the message
// 		var chatMessage struct {
// 			ReceiverUUID string `json:"receiver_uuid"`
// 			Content      string `json:"content"`
// 		}

// 		if err := json.Unmarshal(message, &chatMessage); err != nil {
// 			log.Printf("Error parsing message: %v", err)
// 			continue
// 		}

// 		// Create a new message
// 		newMessage := models.NewMessage(
// 			userUUID.(string),
// 			chatMessage.ReceiverUUID,
// 			chatMessage.Content,
// 		)

// 		// Convert message to JSON
// 		messageJSON, err := json.Marshal(newMessage)
// 		if err != nil {
// 			log.Printf("Error marshaling message: %v", err)
// 			continue
// 		}

// 		// Send message to recipient
// 		if err := services.Registry.SendMessage(chatMessage.ReceiverUUID, messageJSON); err != nil {
// 			log.Printf("Error sending message: %v", err)
// 		}
// 	}
// }
