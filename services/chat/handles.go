// handlers/chat_handlers.go
package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// HandleGetRoom gets details about a specific room
func HandleGetRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	userID := c.GetString("user_id")

	// Use the exported GetRoom function
	room, exists := GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Check if user is authorized for this room
	if room.IsPrivate {
		if _, ok := room.AuthorizedMembers[userID]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized for this room"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             room.ID,
		"name":           room.Name,
		"is_private":     room.IsPrivate,
		"active_members": len(room.ActiveMembers),
	})
}

// HandleCreateRoom creates a new chat room
// a room is created between two users (at least)
func HandleCreateRoom(c *gin.Context) {

	var req struct {
		UserID        string `json:"uuid" binding:"required"`
		RoomName      string `json:"roomname" binding:"required"`
		SecondUser_id string `json:"guest" binding:"required"`
		IsPrivate     bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := req.UserID

	// Use the exported CreateRoom function, this creates a room with the user who created it
	room := CreateRoom(req.RoomName, userID, req.IsPrivate)

	//add the guest user
	success := AddAuthorizedMember(room.ID, req.SecondUser_id, userID)

	if !success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro occured in the HAnDLECREAateROOM",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"room_id": room.ID,
		"name":    room.Name,
	})
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(c *gin.Context) {

	userID := c.Query("user_id")
	userName := c.Query("username")

	if userID == "" || userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Use the exported HandleWebSocketConnection function
	HandleWebSocketConnection(userID, userName, conn)
}

func HandleAddUsertoRoom(c *gin.Context) {
	var req struct {
		UserID        string `json:"uuid" binding:"required"`     //senders uuid, admin
		RoomID        string `json:"roomname" binding:"required"` //room_id to add user to
		SecondUser_id string `json:"guest" binding:"required"`    //guest user to add to the room
		IsPrivate     bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := req.UserID

	// Use the exported CreateRoom function, this creates a room with the user who created it
	// AddAuthorizedMember
	// room := CreateRoom(req.RoomName, userID, req.IsPrivate)

	//add the guest user
	success := AddAuthorizedMember(req.RoomID, req.SecondUser_id, userID)

	if !success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro occured in the HAnDLECREAateROOM",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    req.SecondUser_id,
		"message": "successfuly added user to the room",
	})
}

// RegisterChatRoutes registers all chat-related routes
func RegisterChatRoutes(router *gin.Engine) {
	chatGroup := router.Group("/")
	{
		chatGroup.GET("/rooms/:roomId", HandleGetRoom)
		chatGroup.POST("/createroom", HandleCreateRoom)
		chatGroup.POST("/addusertoroom", HandleAddUsertoRoom)
		chatGroup.GET("/ws", HandleWebSocket)
		// Add more routes as needed
	}
}

// func AuthorizationMiddleWare() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "Authorization header is required",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Check for Bearer prefix
// 		tokenParts := strings.Split(authHeader, "Bearer ")
// 		if len(tokenParts) != 2 {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "Authorization header must be in format 'Bearer {token}'",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		c.Set("userID", )
// 	}
// }
