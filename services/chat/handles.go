// handlers/chat_handlers.go
package chat

import (
	"fmt"
	"net/http"
	db "raychat/database"
	"raychat/services/auth"

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
	userUUID, _ := c.Get("userUUID")
	userUUIDStr := userUUID.(string)
	var req struct {
		// UserID     string `json:"uuid" binding:"required"`
		RoomName   string `json:"roomname" binding:"required"`
		GuestEmail string `json:"guest_email" binding:"required"`
		IsPrivate  bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get guest user UUID by email using existing function
	guestUUID, err := db.Store.GetUserUUIDByEmail(req.GuestEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Guest user not found with email: " + req.GuestEmail,
		})
		return
	}

	// Create room with creator as admin
	room := CreateRoom(req.RoomName, userUUIDStr, req.IsPrivate)

	// Add guest user as authorized member using existing function
	if err := db.Store.AddUserToRoom(guestUUID, room.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add guest user to room",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"room_id": room.ID,
		"name":    room.Name,
		"message": "Room created successfully and guest user added",
	})
}

func HandleAddUsertoRoom(c *gin.Context) {
	var req struct {
		RoomID    string `json:"room_id" binding:"required"`
		UserEmail string `json:"user_email" binding:"required"`
		MakeAdmin bool   `json:"make_admin"` // Optional: make the added user an admin
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get requesting user UUID from auth middleware
	requestingUserUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	requestingUserUUIDStr := requestingUserUUID.(string)

	// Check if requesting user is admin of the room using existing function
	isAdmin, err := db.Store.IsUserAdmin(requestingUserUUIDStr, req.RoomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can add users to this room"})
		return
	}

	// Get user UUID by email using existing function
	userUUID, err := db.Store.GetUserUUIDByEmail(req.UserEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User not found with email: " + req.UserEmail,
		})
		return
	}

	// Add user to room using existing functions
	if req.MakeAdmin {
		err = db.Store.AddAdminToRoom(userUUID, req.RoomID)
	} else {
		err = db.Store.AddUserToRoom(userUUID, req.RoomID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add user to room",
		})
		return
	}

	role := "member"
	if req.MakeAdmin {
		role = "admin"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("User added to room as %s", role),
		"room_id": req.RoomID,
	})
}

// Update your HandleListRooms function
func HandleListRooms(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUIDStr := userUUID.(string)

	// Get rooms for user from Valkey
	rooms, err := db.Store.GetRoomsForUser(userUUIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve rooms",
		})
		return
	}

	// Add user role information and active count to each room
	roomsWithStatus := make([]map[string]interface{}, 0, len(rooms))
	for _, room := range rooms {
		isAdmin := room.Admins != nil && room.Admins[userUUIDStr] //to see of the user is the admin

		// Get active users count
		activeUsers, err := db.Store.GetActiveUsers(room.ID)
		activeCount := 0
		if err == nil {
			activeCount = len(activeUsers)
		}

		roomInfo := map[string]interface{}{
			"id":           room.ID,
			"name":         room.Name,
			"creator_id":   room.CreatorID,
			"is_private":   room.IsPrivate,
			"created_at":   room.CreatedAt,
			"is_admin":     isAdmin,
			"member_count": len(room.AuthorizedMembers),
			"active_count": activeCount,
		}

		roomsWithStatus = append(roomsWithStatus, roomInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": roomsWithStatus,
	})
}

// success := AddAuthorizedMember(room.ID, req.SecondUser_id, userID)
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

// func HandleAddUsertoRoom(c *gin.Context) {
// 	var req struct {
// 		UserID        string `json:"uuid" binding:"required"`     //senders uuid, admin
// 		RoomID        string `json:"roomname" binding:"required"` //room_id to add user to
// 		SecondUser_id string `json:"guest" binding:"required"`    //guest user to add to the room
// 		IsPrivate     bool   `json:"is_private"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	userID := req.UserID

// 	// Use the exported CreateRoom function, this creates a room with the user who created it
// 	// AddAuthorizedMember
// 	// room := CreateRoom(req.RoomName, userID, req.IsPrivate)

// 	//add the guest user
// 	success := AddAuthorizedMember(req.RoomID, req.SecondUser_id, userID)

// 	if !success {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "Erro occured in the HAnDLECREAateROOM",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"user":    req.SecondUser_id,
// 		"message": "successfuly added user to the room",
// 	})
// }

// RegisterChatRoutes registers all chat-related routes
func RegisterChatRoutes(router *gin.Engine) {
	chatGroup := router.Group("/chat")
	chatGroup.Use(auth.AuthRequired())
	{
		chatGroup.GET("/rooms/:roomId", HandleGetRoom)
		chatGroup.GET("/rooms", HandleListRooms)
		chatGroup.POST("/createroom", HandleCreateRoom)
		chatGroup.POST("/addusertoroom", HandleAddUsertoRoom)
		chatGroup.GET("/ws", HandleWebSocket)
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
